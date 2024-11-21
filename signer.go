package main

import (
	"context"
	"crypto"
	"crypto/x509"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
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
	ctx, _ = context.WithTimeout(ctx, 10*time.Second)

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
	ctx, _ = context.WithTimeout(ctx, s.signTimeout)

	kmsDigest := &kms.GenerateMacInput{
		KeyId:        aws.String(s.keyId),
		Message:      digest,
		MacAlgorithm: types.MacAlgorithmSpecHmacSha256,
	}

	result, err := s.client.GenerateMac(ctx, kmsDigest)
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %s", err)
	}

	res, err := s.client.Sign(ctx, &kms.SignInput{
		KeyId:   &s.keyId,
		Message: result.Mac,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %s", err)
	}

	return res.Signature, nil
}
