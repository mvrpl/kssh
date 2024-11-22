package main

import (
	"context"
	"crypto"
	"crypto/sha256"
	"crypto/x509"
	"fmt"
	"io"
	"time"

	kms "github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
)

type Signer struct {
	keyId       string
	client      *kms.Client
	signTimeout time.Duration

	pubKey crypto.PublicKey
}

func NewSigner(client *kms.Client, keyId string) (*Signer, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pubKeypb, err := client.GetPublicKey(ctx, &kms.GetPublicKeyInput{
		KeyId: &keyId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %s", err)
	}

	pubKey, err := x509.ParsePKIXPublicKey(pubKeypb.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %s", err)
	}

	return &Signer{
		keyId:       keyId,
		client:      client,
		signTimeout: 15 * time.Second,
		pubKey:      pubKey,
	}, nil
}

func (s *Signer) Public() crypto.PublicKey {
	return s.pubKey
}

func (s *Signer) HashFunc() crypto.Hash {
	return crypto.SHA256
}

func (s *Signer) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) (signature []byte, err error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, s.signTimeout)
	defer cancel()

	fmt.Println("=======SIGN======")

	hash := sha256.New()
	hash.Write(digest)
	hashedMessage := hash.Sum(nil)

	res, err := s.client.Sign(ctx, &kms.SignInput{
		KeyId:            &s.keyId,
		Message:          hashedMessage,
		SigningAlgorithm: types.SigningAlgorithmSpecRsassaPkcs1V15Sha256,
		MessageType:      types.MessageTypeDigest,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to sign: %s", err)
	}

	return res.Signature, nil
}
