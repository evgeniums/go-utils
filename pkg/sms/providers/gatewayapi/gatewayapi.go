package gatewayapi

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/http_request"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/sms"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

type Recipient struct {
	Msisdn string `json:"msisdn"`
}

type Message struct {
	Message    string      `json:"message"`
	Recipients []Recipient `json:"recipients"`
	Usereref   string      `json:"userref"`
}

type GoodResponse struct {
	Ids []string `json:"ids"`
}

type SmsGatewayapiConfig struct {
	URL   string `validation:"required,url"`
	TOKEN string `validation:"required"`
}

type SmsGatewayapi struct {
	SmsGatewayapiConfig
	sendUrl string
}

func (s *SmsGatewayapi) Config() interface{} {
	return &s.SmsGatewayapiConfig
}

func (s *SmsGatewayapi) Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {

	err := object_config.LoadLogValidate(cfg, log, vld, s, "sms.gatewayapi")
	if err != nil {
		return log.Fatal("failed to init SmsGatewayapi", err)
	}

	s.sendUrl = fmt.Sprintf("%s/rest/mtsms", s.URL)
	return nil
}

func (s *SmsGatewayapi) Send(ctx op_context.Context, message string, recipient string, smsID ...string) (*sms.ProviderResponse, error) {

	c := ctx.TraceInMethod("SmsGatewayapi.Send", logger.Fields{"recipient": recipient})
	var err error
	onExit := func() {
		if err != nil {
			ctx.SetGenericErrorCode(sms.ErrorCodeSmsSendingFailed)
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	recipients := []Recipient{{Msisdn: recipient}}
	msg := &Message{Message: message, Recipients: recipients}
	if len(smsID) > 0 {
		msg.Usereref = smsID[0]
		c.Fields()["sms_id"] = msg.Usereref
	}

	request, err := http_request.NewPost(ctx, s.sendUrl, msg)
	if err != nil {
		return nil, err
	}
	request.SetAuthHeader("Basic", s.TOKEN)

	response := &GoodResponse{}
	request.GoodResponse = response

	err = request.Send(ctx)
	c.Fields()["response_content"] = request.ResponseContent
	c.Fields()["response_status"] = request.ResponseStatus
	if err != nil {
		return nil, err
	}

	result := &sms.ProviderResponse{RawContent: request.ResponseContent}
	if len(response.Ids) > 0 {
		result.ProviderMessageID = response.Ids[0]
	}

	c.Logger().Debug("success")
	return result, nil
}
