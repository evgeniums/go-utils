package rest_api_gin_server

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/evgeniums/go-backend-helpers/pkg/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/gin-gonic/gin"
)

type Request struct {
	api_server.RequestBase

	server   *Server
	ginCtx   *gin.Context
	response *Response

	initialPath string
	start       time.Time

	endpoint api_server.Endpoint
}

func (r *Request) Init(s *Server, ginCtx *gin.Context, ep api_server.Endpoint, fields ...logger.Fields) {

	r.start = time.Now()
	r.server = s

	r.RequestBase.Init(s.App(), s.App().Logger(), s.App().DB(), fields...)
	r.RequestBase.SetErrorManager(s)

	r.ginCtx = ginCtx
	r.response = &Response{request: r, httpCode: http.StatusOK}

	r.initialPath = ginCtx.Request.URL.Path

	r.endpoint = ep
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

func (r *Request) GetRequestPath() string {
	return r.endpoint.FullPath()
}

func (r *Request) GetRequestMethod() string {
	return r.ginCtx.Request.Method
}

func (r *Request) GetRequestClientIp() string {
	return r.ginCtx.ClientIP()
}

func (r *Request) GetRequestUserAgent() string {
	return r.ginCtx.Request.UserAgent()
}

func (r *Request) Close(successMessage ...string) {
	if r.GenericError() == nil {
		if r.response.Message() != nil {
			r.ginCtx.JSON(r.response.httpCode, r.response.Message())
		} else {
			r.ginCtx.Status(r.response.httpCode)
		}
		r.SetLoggerField("status", "success")
	} else {
		code, err := r.server.MakeResponseError(r.GenericError())
		if code < http.StatusInternalServerError {
			r.SetErrorAsWarn(true)
		}
		r.SetLoggerField("status", err.Code)
		r.ginCtx.JSON(code, err)
	}
	r.RequestBase.Close("")
	r.server.logGinRequest(r.Logger(), r.initialPath, r.start, r.ginCtx)
}

func (r *Request) TenancyInPath() string {
	return r.ginCtx.Param(TenancyParameter)
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
	return r.endpoint.PrecheckRequestBeforeAuth(r, smsMessage)
}
