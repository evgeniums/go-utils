package signature

import (
	"errors"
	"net/http"

	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/crypt_utils"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

type SignatureManager interface {
	generic_error.ErrorDefinitions

	Verify(ctx auth.AuthContext, message string, signature string, extraData string) error
	Find(ctx op_context.Context, context string) (*MessageSignature, error)
}

const (
	ErrorCodeInvalidSignature string = "invalid_signature"
)

var ErrorDescriptions = map[string]string{
	ErrorCodeInvalidSignature: "Invalid signature.",
}

var ErrorHttpCodes = map[string]int{
	ErrorCodeInvalidSignature: http.StatusUnauthorized,
}

type SignatureManagerBaseConfig struct {
	ALGORITHM             string `validate:"required,oneof:rsa_h256_signature" default:"rsa_h256_signature"`
	ENCRYPT_MESSAGE_STORE bool
	SECRET                string `mask:"true"`
	SALT                  string `mask:"true"`
}

type SignatureManagerBase struct {
	SignatureManagerBaseConfig
	cipher *crypt_utils.AEAD
}

func NewSignatureManager() *SignatureManagerBase {
	return &SignatureManagerBase{}
}

func (s *SignatureManagerBase) Config() interface{} {
	return &s.SignatureManagerBaseConfig
}

func (s *SignatureManagerBase) Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {

	// load configuration
	path := utils.OptionalArg("signature", configPath...)
	err := object_config.LoadLogValidate(cfg, log, vld, s, path)
	if err != nil {
		return log.PushFatalStack("failed to init signature manager", err)
	}

	// init cipher
	if s.ENCRYPT_MESSAGE_STORE {
		if s.SECRET == "" {
			return log.PushFatalStack("encryption secret must not be empty", nil)
		}
		if s.SALT == "" {
			return log.PushFatalStack("encryption salt must not be empty", nil)
		}
		s.cipher, err = crypt_utils.NewAEAD(s.SECRET, []byte(s.SALT))
		if err != nil {
			return log.PushFatalStack("failed to init cipher for signature manager", err)
		}
	}

	// done
	return nil
}

func (s *SignatureManagerBase) Verify(ctx auth.AuthContext, message string, signature string, extraData string) error {

	// setup
	c := ctx.TraceInMethod("SignatureManagerBase.Verify", logger.Fields{"user": ctx.AuthUser().Display(), "extra_data": extraData})
	var err error
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// TODO support other algorithms
	if s.ALGORITHM != crypt_utils.RSA_H256_SIGNATURE {
		err = errors.New("unsupported algorithm")
		return err
	}

	// verify
	key := ctx.GetAuthParameter(s.ALGORITHM, "pubkey")
	verifier := crypt_utils.NewRsaVerifier()
	err = verifier.LoadKey([]byte(key))
	if err != nil {
		c.SetMessage("failed to load public key")
		return err
	}
	err = crypt_utils.VerifySignature(verifier, []byte(message), signature)
	if err != nil {
		c.SetMessage("invalid signature")
		ctx.SetGenericErrorCode(ErrorCodeInvalidSignature)
		return err
	}

	// keep signature
	obj := &MessageSignature{}
	obj.InitObject()
	obj.Context = ctx.ID()
	obj.SetUser(ctx.AuthUser())
	obj.Operation = ctx.Name()
	obj.Algorithm = s.ALGORITHM
	obj.Signature = signature
	obj.ExtraData = extraData
	obj.PubKeyHash = ctx.GetAuthParameter(s.ALGORITHM, "pubkey_hash")
	if s.ENCRYPT_MESSAGE_STORE {
		ciphertext, err := s.cipher.Encrypt([]byte(message))
		if err != nil {
			c.SetMessage("failed to encrypt message")
			ctx.SetGenericErrorCode(generic_error.ErrorCodeInternalServerError)
			return err
		}
		enc := utils.Base64StringCoding{}
		obj.Message = enc.Encode(ciphertext)
	} else {
		obj.Message = message
	}
	err = op_context.DB(ctx).Create(ctx, obj)
	if err != nil {
		c.SetMessage("failed to save message signature in database")
		ctx.SetGenericErrorCode(generic_error.ErrorCodeInternalServerError)
		return err
	}

	// done
	return err
}

func (s *SignatureManagerBase) AttachToErrorManager(errManager generic_error.ErrorManager) {
	errManager.AddErrorDescriptions(ErrorDescriptions)
	errManager.AddErrorProtocolCodes(ErrorHttpCodes)
}

func (s *SignatureManagerBase) Find(ctx op_context.Context, contextId string) (*MessageSignature, error) {

	c := ctx.TraceInMethod("SmsManagerBase.FindSms", logger.Fields{"signature_context_id": contextId})
	var err error
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	obj := &MessageSignature{}
	found, err := op_context.DB(ctx).FindByField(ctx, "context", contextId, obj)
	if err != nil {
		c.SetMessage("failed to find signature in database")
		return nil, err
	}
	if !found {
		err = errors.New("signature not found")
		return nil, err
	}

	return obj, nil
}
