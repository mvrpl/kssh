package signer

import (
	"context"
	"crypto/rand"
	"crypto/x509"
	"fmt"
	"log"
	"math/big"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	kms "github.com/aws/aws-sdk-go-v2/service/kms"
)

func TestSigner(t *testing.T) {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("sa-east-1"))
	if err != nil {
		log.Fatal(err)
	}
	client := kms.NewFromConfig(cfg)

	kmsKeyId := os.Getenv("KSSH_KEY_ID")

	signer, err := NewSigner(client, kmsKeyId)
	if err != nil {
		log.Fatal(err)
	}

	rootCa := &x509.Certificate{
		SerialNumber: big.NewInt(1),
	}

	data, _ := x509.CreateCertificate(rand.Reader, rootCa, rootCa, signer.Public(), signer)
	cert, _ := x509.ParseCertificate(data)

	msg := "olá Mundo!"
	h := signer.HashFunc().New()
	h.Write([]byte(msg))
	digest := h.Sum(nil)
	signature, err := signer.Sign(rand.Reader, digest, nil)
	if err != nil {
		log.Fatal(err)
	}

	if err := cert.CheckSignature(cert.SignatureAlgorithm, []byte(msg), signature); err != nil {
		log.Fatal(err)
	}

	fmt.Println("OK")
}
