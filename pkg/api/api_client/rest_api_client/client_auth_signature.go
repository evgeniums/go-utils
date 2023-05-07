package rest_api_client

import (
	"encoding/json"

	"github.com/evgeniums/go-backend-helpers/pkg/access_control"
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/auth/auth_methods/auth_signature"
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/crypt_utils"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

type ClientAuthSignature struct {
	Signer          crypt_utils.ESigner
	EndpoindsConfig auth.EndpointsAuthConfig
}

func (a *ClientAuthSignature) MakeHeaders(ctx op_context.Context, operation api.Operation, cmd interface{}) (map[string]string, error) {

	// setup
	c := ctx.TraceInMethod("ClientAuthSignature.MakeHeaders")
	defer ctx.TraceOutMethod()

	// find auth schema for operation
	path := operation.Resource().PathPrototype()
	schema, found := a.EndpoindsConfig.Schema(path, operation.AccessType())
	if !found || schema != auth_signature.SignatureProtocol {
		return nil, nil
	}

	// serialize command
	content, err := json.Marshal(cmd)
	if err != nil {
		c.SetMessage("failed to marshal command")
		return nil, c.SetError(err)
	}

	// sign request
	sig, err := a.Signer.SignB64(content, access_control.Access2HttpMethod(operation.AccessType()), path)
	if err != nil {
		c.SetMessage("failed to sign request")
		return nil, c.SetError(err)
	}

	// put signature to header
	h := map[string]string{"x-auth-signature": sig}

	// done
	return h, nil
}

type ClientAuthSignatureBaseConfig struct {
	PRIVATE_KEY_FILE     string `validate:"required"`
	PRIVATE_KEY_PASSWORD string `mask:"true"`
}

type ClientAuthSignatureBase struct {
	ClientAuthSignatureBaseConfig
	ClientAuthSignature

	rsaSigner       *crypt_utils.RsaSigner
	endpointsConfig *auth.EndpointsAuthConfigBase
}

func (a *ClientAuthSignatureBase) Config() interface{} {
	return &a.ClientAuthSignatureBaseConfig
}

func (a *ClientAuthSignatureBase) Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {

	path := utils.OptionalArg("auth_signature", configPath...)

	// load config
	err := object_config.LoadLogValidate(cfg, log, vld, a, path)
	if err != nil {
		return log.PushFatalStack("failed to load configuration of client auth signature", err)
	}

	// load key
	if a.rsaSigner != nil {
		err = a.rsaSigner.LoadKeyFromFile(a.PRIVATE_KEY_FILE, a.PRIVATE_KEY_PASSWORD)
		if err != nil {
			return log.PushFatalStack("failed to load private key for RSA signer of client auth signature", err)
		}
	}

	// load endpoints configuration
	endpointesPath := object_config.Key(path, "endpoints")
	err = a.endpointsConfig.Init(cfg, log, vld, endpointesPath)
	if err != nil {
		return log.PushFatalStack("failed to load endpoints configuration of client auth signature", err)
	}

	// done
	return nil
}

func NewClientAuthSignature() *ClientAuthSignatureBase {

	c := &ClientAuthSignatureBase{}

	c.rsaSigner = crypt_utils.NewRsaSigner()
	c.Signer = c.rsaSigner
	c.endpointsConfig = auth.NewEndpointsAuthConfigBase()
	c.EndpoindsConfig = c.endpointsConfig

	return c
}
