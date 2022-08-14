package utils

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
)

func ParseRsaPublicKey(rsaPublicKey string) (*rsa.PublicKey, error) {

	var err error

	pubPem, _ := pem.Decode([]byte(rsaPublicKey))
	if pubPem == nil {
		return nil, errors.New("RSA public key not in pem format")
	}
	if pubPem.Type != "RSA PUBLIC KEY" && pubPem.Type != "PUBLIC KEY" {
		return nil, errors.New("RSA public key is of the wrong type")
	}

	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKIXPublicKey(pubPem.Bytes); err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("unable to parse RSA public key: %v", err.Error()))
	}

	var ok bool
	var pubKey *rsa.PublicKey
	if pubKey, ok = parsedKey.(*rsa.PublicKey); !ok {
		return nil, errors.New("unable to parse RSA public key")
	}

	return pubKey, nil
}

func LoadRsaPublicKey(rsaPublicKeyLocation string) (*rsa.PublicKey, error) {
	pub, err := ioutil.ReadFile(rsaPublicKeyLocation)
	if err != nil {
		return nil, errors.New("no RSA public key found")
	}
	return ParseRsaPublicKey(string(pub))
}
