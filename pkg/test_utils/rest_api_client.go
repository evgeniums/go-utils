package test_utils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client/rest_api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/http_request"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/go-querystring/query"
	"github.com/stretchr/testify/require"
)

type RestApiTestResponse struct {
	Raw         *httptest.ResponseRecorder
	body        []byte
	serverError *api.ResponseError
}

func NewRestApiTestResponse(raw *httptest.ResponseRecorder) *RestApiTestResponse {
	return &RestApiTestResponse{Raw: raw}
}

func (r *RestApiTestResponse) Code() int {
	return r.Raw.Code
}

func (r *RestApiTestResponse) Header() http.Header {
	return r.Raw.Header()
}

func (r *RestApiTestResponse) Body() []byte {
	if r.Raw.Body != nil {
		r.body = r.Raw.Body.Bytes()
	}
	return r.body
}

func (r *RestApiTestResponse) Message() string {
	return string(r.Body())
}

func (r *RestApiTestResponse) Error() *api.ResponseError {
	return r.serverError
}

func (r *RestApiTestResponse) SetError(err *api.ResponseError) {
	r.serverError = err
}

func HttptestSendWithBody(t *testing.T, g *gin.Engine, method string, url string, cmd interface{}, headers ...map[string]string) *RestApiTestResponse {

	// prepare data
	cmdByte, err := json.Marshal(cmd)
	require.NoErrorf(t, err, "failed to marshal message")

	// create request
	req, err := http.NewRequest(method, url, bytes.NewBuffer(cmdByte))
	require.NoErrorf(t, err, "failed to create HTTP request")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	http_request.HttpHeadersSet(req, headers...)

	// send request
	httprec := httptest.NewRecorder()
	g.ServeHTTP(httprec, req)

	// done
	return NewRestApiTestResponse(httprec)
}

func HttptestSendWithQuery(t *testing.T, g *gin.Engine, method string, url string, cmd interface{}, headers ...map[string]string) *RestApiTestResponse {

	// create request
	req, err := http.NewRequest(method, url, nil)
	require.NoErrorf(t, err, "failed to create request")

	// prepare data
	v, err := query.Values(cmd)
	require.NoErrorf(t, err, "failed to build query")
	req.URL.RawQuery = v.Encode()
	req.Header.Set("Accept", "application/json")
	http_request.HttpHeadersSet(req, headers...)

	// send request
	httprec := httptest.NewRecorder()
	g.ServeHTTP(httprec, req)

	// done
	return NewRestApiTestResponse(httprec)
}

func RestApiTestClient(t *testing.T, g *gin.Engine, baseUrl string, userAgent ...string) *rest_api_client.RestApiClientBase {

	sendWithBody := func(ctx op_context.Context, method string, url string, cmd interface{}, headers ...map[string]string) (rest_api_client.Response, error) {
		resp := HttptestSendWithBody(t, g, method, url, cmd, headers...)
		return resp, nil
	}
	sendWithQuery := func(ctx op_context.Context, method string, url string, cmd interface{}, headers ...map[string]string) (rest_api_client.Response, error) {
		resp := HttptestSendWithQuery(t, g, method, url, cmd, headers...)
		return resp, nil
	}

	c := rest_api_client.NewRestApiClientBase(sendWithBody, sendWithQuery)
	c.Init(baseUrl, utils.OptionalArg("go-backend-helpers", userAgent...))
	return c
}
