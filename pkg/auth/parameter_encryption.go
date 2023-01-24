package auth

import (
	"errors"

	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/crypt_utils"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/message"
	"github.com/evgeniums/go-backend-helpers/pkg/message/message_json"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

type AuthParameterEncryption interface {
	SetAuthParameter(ctx AuthContext, authMethodProtocol string, name string, obj interface{}) error
	GetAuthParameter(ctx AuthContext, authMethodProtocol string, name string, obj interface{}) (bool, error)
}

type AuthParameterEncryptionBaseConfig struct {
	SECRET            string `validate:"required" mask:"true"`
	PBKDF2_ITERATIONS uint   `default:"256"`
	SALT_SIZE         int    `default:"8" validate:"lte=32,gte=4"`
}

type AuthParameterEncryptionBase struct {
	AuthParameterEncryptionBaseConfig
	Serializer   message.Serializer
	StringCoding utils.StringCoding
}

func (a *AuthParameterEncryptionBase) Config() interface{} {
	return a.AuthParameterEncryptionBaseConfig
}

func (a *AuthParameterEncryptionBase) Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {
	a.Serializer = &message_json.JsonSerializer{}
	a.StringCoding = &utils.Base64StringCoding{}

	err := object_config.LoadLogValidate(cfg, log, vld, a, "auth.params_encryption", configPath...)
	if err != nil {
		return log.Fatal("Failed to load configuration of auth parameters encryption", err)
	}
	return nil
}

func (a *AuthParameterEncryptionBase) createCipher(salt []byte) (*crypt_utils.AEAD, error) {
	pbkdfCfg := crypt_utils.DefaultPbkdfConfig()
	pbkdfCfg.Iter = int(a.PBKDF2_ITERATIONS)
	aeadCfg := crypt_utils.DefaultAEADConfig(pbkdfCfg)
	cipher, err := crypt_utils.NewAEAD(a.SECRET, salt, aeadCfg)
	return cipher, err
}

func (a *AuthParameterEncryptionBase) SetAuthParameter(ctx AuthContext, authMethodProtocol string, name string, obj interface{}) error {

	// setup
	c := ctx.TraceInMethod("AuthParameterEncryptionBase.SetAuthParameter", logger.Fields{"name": name})
	var err error
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// serialize object to plaintext
	plaintext, err := a.Serializer.SerializeMessage(obj)
	if err != nil {
		c.SetMessage("failed to serialize object")
		return err
	}

	// generate salt
	salt, err := crypt_utils.GenerateCryptoRand(a.SALT_SIZE)
	if err != nil {
		c.SetMessage("failed to generate salt")
		return err
	}

	// create cipher
	cipher, err := a.createCipher(salt)
	if err != nil {
		c.SetMessage("failed to create AEAD cipher")
		return err
	}

	// encrypt data
	ciphertext, err := cipher.Encrypt(plaintext)
	if err != nil {
		c.SetMessage("failed to encrypt data")
		return err
	}

	// append salt to ciphertext
	ciphertext = append(ciphertext, salt...)

	// encode data to string
	data := a.StringCoding.Encode(ciphertext)

	// write result to  auth parameter
	ctx.SetAuthParameter(authMethodProtocol, name, data)

	// done
	return nil
}

func (a *AuthParameterEncryptionBase) GetAuthParameter(ctx AuthContext, authMethodProtocol string, name string, obj interface{}) (bool, error) {

	// setup
	c := ctx.TraceInMethod("AuthParameterEncryptionBase.GetAuthParameter", logger.Fields{"name": name})
	var err error
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// read auth parameter
	data := ctx.GetAuthParameter(authMethodProtocol, name)
	if data == "" {
		return false, nil
	}

	// decode data
	ciphertext, err := a.StringCoding.Decode(data)
	if err != nil {
		c.SetMessage("failed to decode data")
		return true, err
	}

	// split data to salt and ciphertext
	if len(ciphertext) < a.SALT_SIZE {
		err := errors.New("ciphertext too short for salt")
		return true, err
	}
	salt := ciphertext[len(ciphertext)-a.SALT_SIZE:]
	ciphertext = ciphertext[:len(ciphertext)-len(salt)]

	// create cipher
	cipher, err := a.createCipher(salt)
	if err != nil {
		c.SetMessage("failed to create AEAD cipher")
		return true, err
	}

	// decrypt data
	plaintext, err := cipher.Decrypt(ciphertext)
	if err != nil {
		c.SetMessage("failed to decrypt ciphertext")
		return true, err
	}

	// parse message
	err = a.Serializer.ParseMessage(plaintext, obj)
	if err != nil {
		c.SetMessage("failed to parse plaintext")
		return true, err
	}

	// done
	return true, nil
}
