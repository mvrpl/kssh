package main

import (
	"context"
	"crypto"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	kms "github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/demille/termsize"
	"github.com/mvrpl/kssh/signer"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/term"
)

var kmsKeyAlias = os.Getenv("KSSH_KEY")

var defaultUsername = func() string {
	cu, _ := user.Current()
	return cu.Username
}()

var (
	username = flag.String("l", defaultUsername, "Specifies the user to log in as on the remote machine.")
	port     = flag.Int("p", 22, "Port to connect to on the remote host.")
	key      = flag.String("i", kmsKeyAlias, "Selects a AWS KMS resource Alias.")

	printAuthorizedKey = flag.Bool("authorized_key", false, `print authorized_key (public key)
You can copy the public key to ~/.ssh/authorized_keys in your home directory on the remote machine.`)
)

func main() {
	flag.Parse()

	if *key == "" {
		fmt.Println("Please set kms key")
		os.Exit(2)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	ctx, cancel := context.WithCancel(context.Background())

	signer, err := loadSigner(ctx, *key)
	if err != nil {
		log.Printf("load key: %s", err)
		os.Exit(1)
	}

	s, err := ssh.NewSignerFromSigner(signer)
	if err != nil {
		log.Printf("load signer: %s", err)
		os.Exit(1)
	}

	if *printAuthorizedKey {
		fmt.Printf("%s\n", ssh.MarshalAuthorizedKey(s.PublicKey()))
		return
	}

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}

	hostname := flag.Arg(0)
	if ss := strings.SplitN(hostname, "@", 2); len(ss) == 2 {
		*username = ss[0]
		hostname = ss[1]
	}

	go func() {
		if err := run(ctx, s, *username, hostname, *port); err != nil {
			log.Print(err)
		}
		fmt.Printf("Connection to %s closed.\n", hostname)
		cancel()
	}()

	select {
	case <-sig:
		cancel()
	case <-ctx.Done():
	}
}

func addHostKey(remote net.Addr, pubKey ssh.PublicKey) error {
	khFilePath := filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts")

	f, fErr := os.OpenFile(khFilePath, os.O_APPEND|os.O_WRONLY, 0600)
	if fErr != nil {
		return fErr
	}
	defer f.Close()

	knownHosts := knownhosts.Normalize(remote.String())
	_, fileErr := f.WriteString(knownhosts.Line([]string{knownHosts}, pubKey))
	return fileErr
}

func createKnownHosts() {
	khFilePath := filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts")
	_, err := os.Stat(khFilePath)
	if os.IsNotExist(err) {
		fmt.Printf("File %s does not exist.\n", khFilePath)
	}

	f, fErr := os.OpenFile(khFilePath, os.O_CREATE, 0600)
	if fErr != nil {
		panic(fErr)
	}
	f.Close()
}

func checkKnownHosts() ssh.HostKeyCallback {
	createKnownHosts()
	kh, err := knownhosts.New(filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"))
	if err != nil {
		panic(err)
	}
	return kh
}

func hostKeyString(pubKey ssh.PublicKey) string {
	authorizedKeyBytes := ssh.MarshalAuthorizedKey(pubKey)
	return string(authorizedKeyBytes)
}

func hostKeyCallback() ssh.HostKeyCallback {
	var (
		keyErr *knownhosts.KeyError
	)
	return ssh.HostKeyCallback(func(host string, remote net.Addr, pubKey ssh.PublicKey) error {
		kh := checkKnownHosts()
		hErr := kh(host, remote, pubKey)
		if errors.As(hErr, &keyErr) && len(keyErr.Want) > 0 {
			log.Printf("WARNING: %v is not a key of %s, either a MiTM attack or %s has reconfigured the host pub key.", hostKeyString(pubKey), host, host)
			return keyErr
		} else if errors.As(hErr, &keyErr) && len(keyErr.Want) == 0 {
			log.Printf("WARNING: %s is not trusted, adding this key: %q to known_hosts file.", host, hostKeyString(pubKey))
			return addHostKey(remote, pubKey)
		}
		log.Printf("Pub key exists for %s.", host)
		return nil
	})
}

func run(ctx context.Context, signer ssh.Signer, user, host string, port int) error {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		Timeout:         5 * time.Second,
		HostKeyCallback: hostKeyCallback(),
	}

	hostport := fmt.Sprintf("%s:%d", host, port)
	conn, err := ssh.Dial("tcp", hostport, config)
	if err != nil {
		return fmt.Errorf("cannot connect %v: %v", hostport, err)
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return fmt.Errorf("cannot open new session: %v", err)
	}
	defer session.Close()

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	fd := int(os.Stdin.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("terminal make raw: %s", err)
	}
	defer term.Restore(fd, state)

	if err := termsize.Init(); err != nil {
		panic(err)
	}
	w, h, err := termsize.Size()
	if err != nil {
		return fmt.Errorf("terminal get size: %s", err)
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	term := os.Getenv("TERM")
	if term == "" {
		term = "xterm-256color"
	}
	if err := session.RequestPty(term, h, w, modes); err != nil {
		return fmt.Errorf("session xterm: %s", err)
	}

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	if err := session.Shell(); err != nil {
		return fmt.Errorf("session shell: %s", err)
	}

	if err := session.Wait(); err != nil {
		if e, ok := err.(*ssh.ExitError); ok {
			switch e.ExitStatus() {
			case 130:
				return nil
			}
		}
		return fmt.Errorf("ssh: %s", err)
	}
	return nil
}

func loadSigner(ctx context.Context, key string) (crypto.Signer, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("sa-east-1"))
	if err != nil {
		panic(err)
	}
	client := kms.NewFromConfig(cfg)
	if err != nil {
		return nil, err
	}
	return signer.NewSigner(client, key)
}
