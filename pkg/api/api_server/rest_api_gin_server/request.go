package rest_api_gin_server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/evgeniums/go-backend-helpers/pkg/access_control"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/http_request"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
	"github.com/gin-gonic/gin"
)

type Request struct {
	api_server.RequestBase

	server   *Server
	ginCtx   *gin.Context
	response *Response

	initialPath string
	start       time.Time

	clientIp string
}

func (r *Request) Init(s *Server, ginCtx *gin.Context, ep api_server.Endpoint, fields ...logger.Fields) {

	r.start = time.Now()
	r.server = s

	r.RequestBase.Init(s.App(), s.App().Logger(), s.App().Db(), ep, fields...)
	r.RequestBase.SetErrorManager(s)

	r.ginCtx = ginCtx
	r.response = &Response{httpCode: http.StatusOK}
	r.response.SetRequest(r)

	r.initialPath = ginCtx.Request.URL.Path
	r.clientIp = ginCtx.ClientIP()
}

func (r *Request) Server() api_server.Server {
	return r.server
}

func (r *Request) GetParameter(key string) (any, bool) {
	return r.ginCtx.Get(key)
}

func (r *Request) SetParameter(key string, value any) {
	r.ginCtx.Set(key, value)
}

func (r *Request) Response() api_server.Response {
	return r.response
}

func (r *Request) GetRequestMethod() string {
	return r.ginCtx.Request.Method
}

func (r *Request) GetRequestClientIp() string {
	return r.clientIp
}

func (r *Request) GetRequestUserAgent() string {
	return r.ginCtx.Request.UserAgent()
}

func (r *Request) Close(successMessage ...string) {
	var reponseBody interface{}
	if r.GenericError() == nil {
		if r.response.RedirectPath() != "" {
			r.ginCtx.Redirect(http.StatusFound, r.response.RedirectPath())
		} else if r.response.Text() != "" {
			r.ginCtx.String(r.response.httpCode, r.response.Text())
		} else {
			if r.response.Message() != nil {
				reponseBody = r.response.Message()
				r.ginCtx.JSON(r.response.httpCode, r.response.Message())
			} else {
				r.ginCtx.Status(r.response.httpCode)
			}
		}

		r.SetLoggerField("status", "success")
	} else {
		code, err := r.server.MakeResponseError(r.GenericError())
		if code < http.StatusInternalServerError {
			r.SetErrorAsWarn(true)
		}
		reponseBody = err
		r.SetLoggerField("status", err.Code)
		r.ginCtx.JSON(code, reponseBody)
	}

	if r.server.VERBOSE {
		h := r.ginCtx.Writer.Header()
		body := ""
		if reponseBody != nil {
			b, e := json.Marshal(reponseBody)
			if e == nil {
				body = string(b)
			}
		}
		r.Logger().Debug("Dump HTTP response", logger.Fields{"response_header": h, "response_body": body})
	}

	r.RequestBase.Close("")
	r.server.logGinRequest(r.Logger(), r.initialPath, r.start, r.ginCtx)
}

func (r *Request) GetRequestContent() []byte {
	if r.ginCtx.Request.Body != nil {
		b, _ := io.ReadAll(r.ginCtx.Request.Body)
		r.ginCtx.Request.Body = io.NopCloser(bytes.NewBuffer(b))
		return b
	}
	return nil
}

func AuthKey(key string) string {
	return utils.ConcatStrings("x-auth-", key)
}

func (r *Request) SetAuthParameter(authMethodProtocol string, key string, value string) {
	handler := r.server.AuthParameterSetter(authMethodProtocol)
	if handler != nil {
		handler(r, key, value)
		return
	}
	r.ginCtx.Header(AuthKey(key), value)
}

func (r *Request) GetAuthParameter(authMethodProtocol string, key string) string {
	handler := r.server.AuthParameterGetter(authMethodProtocol)
	if handler != nil {
		return handler(r, key)
	}
	return getHttpHeader(r.ginCtx, AuthKey(key))
}

func (r *Request) CheckRequestContent(smsMessage *string) error {
	return r.Endpoint().PrecheckRequestBeforeAuth(r, smsMessage)
}

func (r *Request) ResourceIds() map[string]string {
	m := make(map[string]string, 0)
	for _, param := range r.ginCtx.Params {
		m[param.Key] = param.Value
	}
	return m
}

func (r *Request) GetRequestPath() string {
	return api_server.FullRequestServicePath(r)
}

func (r *Request) GetResourceId(resourceType string) string {
	return r.ginCtx.Param(resourceType)
}

func (r *Request) Validate(cmd interface{}) error {

	c := r.TraceInMethod("Request.Validate")
	defer r.TraceOutMethod()

	err := r.App().Validator().Validate(cmd)
	if err != nil {
		vErr, ok := err.(*validator.ValidationError)
		if ok {
			r.SetGenericError(vErr.GenericError(), true)
		}
		return c.SetError(err)
	}
	return nil
}

func (r *Request) ParseValidateQuery(cmd interface{}) error {

	if cmd == nil {
		return nil
	}

	c := r.TraceInMethod("Request.ParseValidateQuery")
	defer r.TraceOutMethod()

	err := http_request.ParseQuery(r, r.ginCtx.Request, cmd)
	if err != nil {
		return c.SetError(err)
	}

	err = r.Validate(cmd)
	if err != nil {
		return c.SetError(err)
	}

	return nil
}

func (r *Request) ParseValidateBody(cmd interface{}) error {

	if cmd == nil {
		return nil
	}

	c := r.TraceInMethod("Request.ParseValidateBody")
	defer r.TraceOutMethod()

	err := http_request.ParseBody(r, r.ginCtx.Request, cmd)
	if err != nil {
		return c.SetError(err)
	}

	err = r.Validate(cmd)
	if err != nil {
		return c.SetError(err)
	}

	return nil
}

func (r *Request) ParseValidate(cmd interface{}) error {

	if access_control.HttpContentInQuery(r.Endpoint().AccessType()) {
		return r.ParseValidateQuery(cmd)
	}

	return r.ParseValidateBody(cmd)
}
