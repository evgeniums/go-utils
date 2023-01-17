package utils

import (
	"crypto/cipher"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"errors"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/pbkdf2"
)

type Pbkdf2Config struct {
	Iter        int
	HashBuilder DigestBuilder
}

func DefaultPbkdfConfig() Pbkdf2Config {
	return Pbkdf2Config{32, sha256.New}
}

type PbkdfFnc = func(password []byte, salt []byte, keySize int) []byte

type PbkdfConfig struct {
	Iter        int
	HashBuilder func() DigestBuilder
}

type MakeAeadFnc = func(key []byte) (cipher.AEAD, error)

type AEADConfig struct {
	MakeAead  MakeAeadFnc
	DeriveKey PbkdfFnc
	KeySize   int
}

type AEAD struct {
	Cipher cipher.AEAD
}

func NewAEAD(secret string, salt string, config ...*AEADConfig) (*AEAD, error) {
	var err error
	a := &AEAD{}

	var cfg *AEADConfig
	if len(config) == 1 {
		cfg = config[0]
	} else {
		cfg = &AEADConfig{}
		cfg.MakeAead = chacha20poly1305.NewX
		cfg.KeySize = chacha20poly1305.KeySize
		cfg.DeriveKey = func(password []byte, salt []byte, keySize int) []byte {
			pbkdfCfg := DefaultPbkdfConfig()
			return pbkdf2.Key(password, salt, pbkdfCfg.Iter, keySize, pbkdfCfg.HashBuilder)
		}
	}

	key := cfg.DeriveKey([]byte(secret), []byte(salt), cfg.KeySize)
	a.Cipher, err = cfg.MakeAead(key)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (a *AEAD) Encrypt(plaintext []byte, additionalData ...[]byte) ([]byte, error) {

	// prepare buffer for nonce and ciphertext
	nonce := make([]byte, a.Cipher.NonceSize(), a.Cipher.NonceSize()+len(plaintext)+a.Cipher.Overhead())
	if _, err := cryptorand.Read(nonce); err != nil {
		return nil, err
	}

	// encrypt and seal data
	extraData := OptionalArg(nil, additionalData...)
	ciphertext := a.Cipher.Seal(nonce, nonce, plaintext, extraData)

	// done
	return ciphertext, nil
}

func (a *AEAD) Decrypt(ciphertext []byte, additionalData ...[]byte) ([]byte, error) {

	// check length
	if len(ciphertext) < a.Cipher.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}

	// split nonce and ciphertext
	nonce, ciphertext := ciphertext[:a.Cipher.NonceSize()], ciphertext[a.Cipher.NonceSize():]

	// decrypt the message and check it wasn't tampered with
	extraData := OptionalArg(nil, additionalData...)
	plaintext, err := a.Cipher.Open(nil, nonce, ciphertext, extraData)
	if err != nil {
		return nil, err
	}

	// done
	return plaintext, nil
}
