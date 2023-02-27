package crypt_utils

import (
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type ESigner interface {
	utils.WithStringCoder
	Sign(data []byte, extraData ...string) ([]byte, error)
}

type EVerifier interface {
	utils.WithStringCoder
	Verify(data []byte, signature []byte, extraData ...string) error
}

func Sign(signer ESigner, data []byte, extraData ...string) (string, error) {
	signature, err := signer.Sign(data)
	if err != nil {
		return "", err
	}
	return signer.Coder().Encode(signature), nil
}

func VerifySignature(verifier EVerifier, data []byte, signature string, extraData ...string) error {
	sig, err := verifier.Coder().Decode(signature)
	if err != nil {
		return err
	}
	return verifier.Verify(data, sig, extraData...)
}
