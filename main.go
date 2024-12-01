package main

import (
	"context"
	"crypto"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"os/user"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	kms "github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/mvrpl/kssh/signer"
	"golang.org/x/crypto/ssh"
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

func run(ctx context.Context, signer ssh.Signer, user, host string, port int) error {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		Timeout:         5 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
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

	w, h, err := term.GetSize(fd)
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
		log.Fatal(err)
	}
	client := kms.NewFromConfig(cfg)
	if err != nil {
		return nil, err
	}
	return signer.NewSigner(client, key)
}
