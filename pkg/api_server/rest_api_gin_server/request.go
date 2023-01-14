package rest_api_gin_server

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/evgeniums/go-backend-helpers/pkg/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/message/message_json"
	"github.com/gin-gonic/gin"
)

type Request struct {
	api_server.RequestBase
	message_json.WithMessageJson

	server   *Server
	ginCtx   *gin.Context
	response *Response

	initialPath string
	start       time.Time
}

func (r *Request) Init(s *Server, ginCtx *gin.Context, fields ...logger.Fields) {

	r.start = time.Now()

	r.ContextBase.Init(s.App(), s.App().Logger(), s.App().DB(), fields...)

	r.ginCtx = ginCtx
	r.response = &Response{request: r, httpCode: http.StatusOK}

	r.initialPath = ginCtx.Request.URL.Path
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

func (r *Request) Close() {
	r.RequestBase.Close()
	if r.GenericError() == nil {
		r.ginCtx.JSON(r.response.httpCode, r.response.Message())
	} else {
		code, err := r.server.MakeResponseError(r.GenericError())
		r.ginCtx.JSON(code, err)
	}
	r.server.logGinRequest(r.Logger(), r.initialPath, r.start, r.ginCtx)
}

func (r *Request) TenancyInPath() string {
	return r.ginCtx.Param(TenancyParameter)
}

type AuthParameters struct {
	request *Request
}

func (a *AuthParameters) SetAuthParameter(key string, value string) {
	a.request.ginCtx.Header(key, value)
}

func (a *AuthParameters) GetAuthParameter(key string) string {
	return a.request.ginCtx.GetHeader(key)
}

func (a *AuthParameters) GetRequestContent() []byte {
	// TODO check if body is accessible after that
	b, _ := ioutil.ReadAll(a.request.ginCtx.Request.Body)
	return b
}

func (r *Request) makeAuthParamsResolver(authProtocolName string) auth.AuthParameters {
	return &AuthParameters{request: r}
}
