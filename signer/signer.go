package signer

import (
	"context"
	"crypto"
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

	algorithm types.SigningAlgorithmSpec

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
		algorithm:   types.SigningAlgorithmSpecRsassaPkcs1V15Sha256,
		pubKey:      pubKey,
	}, nil
}

func (s *Signer) Public() crypto.PublicKey {
	return s.pubKey
}

func (s *Signer) HashFunc() crypto.Hash {
	switch s.algorithm {
	case types.SigningAlgorithmSpecEcdsaSha256,
		types.SigningAlgorithmSpecRsassaPkcs1V15Sha256:
		return crypto.SHA256
	case types.SigningAlgorithmSpecEcdsaSha384,
		types.SigningAlgorithmSpecRsassaPkcs1V15Sha384:
		return crypto.SHA384
	default:
		return 0
	}
}

func (s *Signer) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) (signature []byte, err error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, s.signTimeout)
	defer cancel()

	res, err := s.client.Sign(ctx, &kms.SignInput{
		KeyId:            &s.keyId,
		Message:          digest,
		SigningAlgorithm: s.algorithm,
		MessageType:      types.MessageTypeDigest,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to sign: %s", err)
	}

	return res.Signature, nil
}
