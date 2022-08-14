package utils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

func RsaSign(key *rsa.PrivateKey, data string) (string, error) {

	return RsaSignBytes(key, []byte(data))
}

func RsaSignBytes(key *rsa.PrivateKey, data []byte) (string, error) {

	hashed := sha256.Sum256(data)

	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}

	b64 := base64.RawStdEncoding.WithPadding(base64.StdPadding).EncodeToString(signature)
	return b64, nil
}

func RsaSignSha1(key *rsa.PrivateKey, data string) (string, error) {

	hashed := sha1.Sum([]byte(data))

	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA1, hashed[:])
	if err != nil {
		return "", err
	}

	b64 := base64.RawStdEncoding.WithPadding(base64.StdPadding).EncodeToString(signature)
	return b64, nil
}

func RsaVerifyBytes(key *rsa.PublicKey, data []byte, signature string) error {

	hashed := sha256.Sum256(data)
	sig, err := base64.RawStdEncoding.WithPadding(base64.StdPadding).DecodeString(signature)
	if err != nil {
		return err
	}

	err = rsa.VerifyPKCS1v15(key, crypto.SHA256, hashed[:], sig)
	if err != nil {
		return err
	}

	return nil
}

func RsaVerify(key *rsa.PublicKey, data string, signature string) error {
	return RsaVerifyBytes(key, []byte(data), signature)
}

func PrepareString(method string, uri string, data string) string {
	return fmt.Sprintf("%v\n%v\n%v", method, uri, data)
}

func SignRequest(key *rsa.PrivateKey, method string, uri string, data string) (string, error) {

	str := PrepareString(method, uri, data)

	return RsaSign(key, str)
}

func VerifyRequest(key *rsa.PublicKey, method string, uri string, data string, signature string) error {

	str := PrepareString(method, uri, data)

	return RsaVerify(key, str, signature)
}
