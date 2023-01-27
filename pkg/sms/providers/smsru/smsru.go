package smsru

import (
	"errors"

	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/http_request"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/sms"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

const Protocol string = "smsru"

const (
	CodeOk                   = 100
	CodeInvalidRoute         = 150
	CodeInvalidPhone         = 202
	CodeForeignPhone         = 214
	CodePhoneDayLimit        = 230
	CodePhoneSameMinuteLimit = 231
	CodePhoneSameDayLimit    = 232
	CodePhoneCodeSpamLimit   = 233
)

type SmsruConfig struct {
	URL    string `validate:"required,url"`
	API_ID string `validate:"required" mask:"true"`
	TEST   int
	NAME   string
}

type Smsru struct {
	SmsruConfig
	sms.ProviderBase
}

func New() *Smsru {
	return &Smsru{}
}

func (s *Smsru) Config() interface{} {
	return &s.SmsruConfig
}

type request struct {
	ApiId   string `url:"api_id"`
	To      string `url:"to"`
	Message string `url:"msg"`
	Json    int    `url:"json"`
	Test    int    `url:"test"`
}

type responseItem struct {
	Status       string `json:"status"`
	StatusCode   int    `json:"status_code"`
	SmsId        string `json:"sms_id"`
	ErrorMessage string `json:"status_text"`
}

type response struct {
	Status     string                  `json:"status"`
	StatusCode int                     `json:"status_code"`
	Balance    float64                 `json:"balance"`
	Items      map[string]responseItem `json:"sms"`
}

func (s *Smsru) Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {

	err := object_config.LoadLogValidate(cfg, log, vld, s, "sms.smsru", configPath...)
	if err != nil {
		return log.PushFatalStack("failed to init SmsGatewayapi", err)
	}

	s.ProviderBase.SetProtocolAndName(Protocol, utils.OptionalString(Protocol, s.NAME))
	return nil
}

func (s *Smsru) Send(ctx op_context.Context, message string, recipient string, smsID ...string) (*sms.ProviderResponse, error) {

	c := ctx.TraceInMethod("Smsru.Send", logger.Fields{"recipient": recipient})
	var err error
	onExit := func() {
		if err != nil {
			ctx.SetGenericErrorCode(sms.ErrorCodeSmsSendingFailed)
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// prepare request
	smsReq := &request{ApiId: s.API_ID, To: recipient, Message: message, Json: 1, Test: s.TEST}
	url := s.URL + "/send"
	req, err := http_request.NewGet(ctx, url, smsReq)
	if err != nil {
		return nil, err
	}
	resp := &response{}
	req.GoodResponse = resp
	req.BadResponse = resp

	// send request
	err = req.Send(ctx)
	c.Fields()["response_content"] = req.ResponseContent
	c.Fields()["response_status"] = req.ResponseStatus
	c.Fields()["status_code"] = resp.StatusCode
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != CodeOk {
		err = errors.New("failed status code")
	}

	// fill result
	result := &sms.ProviderResponse{RawContent: req.ResponseContent}
	item, ok := resp.Items[recipient]
	if ok {
		result.ProviderMessageID = item.SmsId
		ctx.SetLoggerField("provider_sms_id", result.ProviderMessageID)
		c.Fields()["sms_status_code"] = item.StatusCode
		if err == nil && item.StatusCode != CodeOk {
			err = errors.New("failed item status code")
		}
	} else {
		if err == nil {
			err = errors.New("phone not found in response")
		} else {
			c.Logger().Warn("phone not found in response")
		}
	}

	// return result
	return result, err
}
