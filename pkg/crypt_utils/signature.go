package crypt_utils

import (
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type ESigner interface {
	utils.WithStringCoder
	Sign(data []byte) ([]byte, error)
}

type EVerifier interface {
	utils.WithStringCoder
	Verify(data []byte, signature []byte) error
}

func Sign(signer ESigner, data []byte) (string, error) {
	signature, err := signer.Sign(data)
	if err != nil {
		return "", err
	}
	return signer.Coder().Encode(signature), nil
}

func Verify(verifier EVerifier, data []byte, signature string) error {
	sig, err := verifier.Coder().Decode(signature)
	if err != nil {
		return err
	}
	return verifier.Verify(data, sig)
}
