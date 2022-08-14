package utils

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
)

func LoadRsaPrivateKey(rsaPublicKeyLocation, rsaPrivateKeyLocation, rsaPrivateKeyPassword string) (*rsa.PrivateKey, error) {
	if rsaPrivateKeyLocation == "" {
		return nil, errors.New("no RSA Key given")
	}

	priv, err := ioutil.ReadFile(rsaPrivateKeyLocation)
	if err != nil {
		return nil, errors.New("no RSA private key found")
	}

	privPem, _ := pem.Decode(priv)
	var privPemBytes []byte
	if privPem.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("RSA private key is of the wrong type")
	}

	if rsaPrivateKeyPassword != "" {
		privPemBytes, err = x509.DecryptPEMBlock(privPem, []byte(rsaPrivateKeyPassword))
		if err != nil {
			return nil, errors.New("unable to decrypt passphrase")
		}
	} else {
		privPemBytes = privPem.Bytes
	}

	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKCS1PrivateKey(privPemBytes); err != nil {
		if parsedKey, err = x509.ParsePKCS8PrivateKey(privPemBytes); err != nil {
			return nil, errors.New("unable to parse RSA private key")
		}
	}

	var privateKey *rsa.PrivateKey
	var ok bool
	privateKey, ok = parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("unable to parse RSA private key")
	}

	if rsaPublicKeyLocation != "" {
		pubKey, err := LoadRsaPublicKey(rsaPublicKeyLocation)
		if err != nil {
			return nil, err
		}
		privateKey.PublicKey = *pubKey
	}

	return privateKey, nil
}
