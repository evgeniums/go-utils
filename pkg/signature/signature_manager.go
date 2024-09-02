package signature

import (
	"errors"
	"net/http"
	"strings"

	"github.com/evgeniums/go-utils/pkg/auth"
	"github.com/evgeniums/go-utils/pkg/config"
	"github.com/evgeniums/go-utils/pkg/config/object_config"
	"github.com/evgeniums/go-utils/pkg/crypt_utils"
	"github.com/evgeniums/go-utils/pkg/db"
	"github.com/evgeniums/go-utils/pkg/generic_error"
	"github.com/evgeniums/go-utils/pkg/logger"
	"github.com/evgeniums/go-utils/pkg/op_context"
	"github.com/evgeniums/go-utils/pkg/utils"
	"github.com/evgeniums/go-utils/pkg/validator"
	"github.com/klauspost/compress/zstd"
)

type UserWithPubkey interface {
	PubKey() string
	PubKeyHash() string
}

type SignatureManager interface {
	generic_error.ErrorDefinitions

	Verify(ctx auth.AuthContext, signature string, message []byte, extraData ...string) error
	CheckPubKey(ctx op_context.Context, key string) error
	SetUserKeyFinder(userKeyFinder func(ctx auth.AuthContext) (UserWithPubkey, error))
}

type WithSignatureManager interface {
	SignatureManager() SignatureManager
}

const (
	ErrorCodeInvalidKey       string = "invalid_key"
	ErrorCodeInvalidSignature string = "invalid_signature"
)

var ErrorDescriptions = map[string]string{
	ErrorCodeInvalidSignature: "Invalid signature",
	ErrorCodeInvalidKey:       "Invalid key",
}

var ErrorHttpCodes = map[string]int{
	ErrorCodeInvalidSignature: http.StatusUnauthorized,
}

type SignatureManagerBaseConfig struct {
	ALGORITHM               string `validate:"required,oneof=rsa_h256_signature" default:"rsa_h256_signature"`
	ENCRYPT_MESSAGE_STORE   bool
	COMPRESS_BEFORE_ENCRYPT bool   `default:"true"`
	SECRET                  string `mask:"true"`
	SALT                    string `mask:"true"`
}

type SignatureManagerBase struct {
	SignatureManagerBaseConfig
	cipher *crypt_utils.AEAD

	zstdEncoder *zstd.Encoder
	zstdDecoder *zstd.Decoder

	userKeyFinder func(ctx auth.AuthContext) (UserWithPubkey, error)
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

	// init compressor
	if s.COMPRESS_BEFORE_ENCRYPT {
		s.zstdEncoder, _ = zstd.NewWriter(nil)
		s.zstdDecoder, _ = zstd.NewReader(nil, zstd.WithDecoderConcurrency(1))
	}

	// done
	return nil
}

func (s *SignatureManagerBase) SetUserKeyFinder(userKeyFinder func(ctx auth.AuthContext) (UserWithPubkey, error)) {
	s.userKeyFinder = userKeyFinder
}

func (s *SignatureManagerBase) CheckPubKey(ctx op_context.Context, key string) error {

	// setup
	c := ctx.TraceInMethod("SignatureManagerBase.CheckPubKey")
	var err error
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// try tp make verifier
	_, err = s.MakeVerifier(ctx, key)
	if err != nil {
		return err
	}

	// done
	return nil
}

func (s *SignatureManagerBase) MakeVerifier(ctx op_context.Context, key string) (crypt_utils.EVerifier, error) {

	// setup
	c := ctx.TraceInMethod("SignatureManagerBase.MakeVerfier")
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
		return nil, err
	}

	// create verifier
	verifier := crypt_utils.NewRsaVerifier()

	// load public key
	err = verifier.LoadKey([]byte(key))
	if err != nil {
		ctx.SetGenericErrorCode(ErrorCodeInvalidKey)
		c.SetMessage("failed to load public key")
		return nil, err
	}

	// done
	return verifier, nil
}

func (s *SignatureManagerBase) Compress(src []byte) []byte {
	return s.zstdEncoder.EncodeAll(src, make([]byte, 0, len(src)))
}

func (s *SignatureManagerBase) Decompress(src []byte) ([]byte, error) {
	return s.zstdDecoder.DecodeAll(src, nil)
}

func (s *SignatureManagerBase) Verify(ctx auth.AuthContext, signature string, message []byte, extraData ...string) error {

	// setup
	c := ctx.TraceInMethod("SignatureManagerBase.Verify")
	var err error
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// extract auth user from context
	if ctx.AuthUser() == nil {
		err = errors.New("user must be authorized")
		ctx.SetGenericErrorCode(generic_error.ErrorCodeInternalServerError)
		return err
	}

	// find user pubkey
	userKey, err := s.userKeyFinder(ctx)
	if err != nil {
		c.SetMessage("failed to find user key")
		ctx.SetGenericErrorCode(generic_error.ErrorCodeInternalServerError)
		return err
	}

	// make verifier
	verifier, err := s.MakeVerifier(ctx, userKey.PubKey())
	if err != nil {
		return err
	}

	// verify
	err = crypt_utils.VerifySignature(verifier, []byte(message), signature, extraData...)
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
	obj.ExtraData = strings.Join(extraData, "+")
	obj.PubKeyHash = userKey.PubKeyHash()
	if s.ENCRYPT_MESSAGE_STORE {
		src := []byte(message)
		if s.COMPRESS_BEFORE_ENCRYPT {
			src = s.Compress(src)
		}
		ciphertext, err := s.cipher.Encrypt(src)
		if err != nil {
			c.SetMessage("failed to encrypt message")
			ctx.SetGenericErrorCode(generic_error.ErrorCodeInternalServerError)
			return err
		}
		enc := utils.Base64StringCoding{}
		obj.Message = enc.Encode(ciphertext)
	} else {
		obj.Message = string(message)
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

	month, err := utils.MonthFromId(contextId)
	if err != nil {
		c.SetMessage("invalid context ID")
		return nil, err
	}

	obj := &MessageSignature{}
	fields := db.Fields{}
	fields["month"] = month
	fields["context"] = contextId
	found, err := op_context.DB(ctx).FindByFields(ctx, fields, obj)
	if err != nil {
		c.SetMessage("failed to find signature in database")
		return nil, err
	}
	if !found {
		err = errors.New("signature not found")
		return nil, err
	}

	// TODO decrypt and decompress data

	return obj, nil
}
