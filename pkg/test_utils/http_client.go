package test_utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/api_server/rest_api_gin_server"
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_methods/auth_login_phash"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Expected struct {
	Error    string
	Message  string
	Details  string
	HttpCode int
}

type HttpResponse struct {
	Object  *httptest.ResponseRecorder
	Code    int
	Message string
}

func CheckResponse(t *testing.T, resp *HttpResponse, expected *Expected) {
	require.NotNil(t, resp)
	if expected == nil {
		assert.Contains(t, []int{http.StatusOK, http.StatusCreated}, resp.Code)
	} else {
		if expected.HttpCode != 0 {
			assert.Equal(t, expected.HttpCode, resp.Code)
		}
		if expected.Error != "" {
			assert.NotEmpty(t, resp.Message)
			if len(resp.Message) != 0 {
				errResp := &rest_api_gin_server.ResponseError{}
				require.NoError(t, json.Unmarshal([]byte(resp.Message), errResp))
				assert.Equal(t, expected.Error, errResp.Code)
				if expected.Details != "" {
					assert.Equal(t, expected.Details, errResp.Details)
				}
				if expected.Message != "" {
					assert.Equal(t, expected.Message, errResp.Message)
				}
			}
		} else {
			if expected.Message != "" {
				assert.Equal(t, expected.Message, resp.Message)
			}
		}
	}
}

type HttpClient struct {
	BaseUrl      string
	AccessToken  string
	RefreshToken string
	CsrfToken    string

	Gin *gin.Engine
}

func NewHttpClient(gin *gin.Engine, baseUrl ...string) *HttpClient {
	c := &HttpClient{}
	c.BaseUrl = utils.OptionalArg("/api/1.0.0", baseUrl...)
	c.Gin = gin
	return c
}

func PrepareHttpClient(t *testing.T, gin *gin.Engine, baseUrl ...string) *HttpClient {
	c := &HttpClient{}
	c.BaseUrl = utils.OptionalArg("/api/1.0.0", baseUrl...)
	c.Gin = gin
	c.Prepare(t)
	return c
}

func (c *HttpClient) Url(path string) string {
	return c.BaseUrl + path
}

func (c *HttpClient) Login(t *testing.T, user string, password string, expectedErrorCode ...string) {

	errorCode := utils.OptionalArg("", expectedErrorCode...)
	path := "/auth/login"

	if errorCode == auth.ErrorCodeUnauthorized {
		// request without headers
		resp := c.Post(t, path, nil)
		CheckResponse(t, resp, &Expected{HttpCode: http.StatusUnauthorized, Error: errorCode})
		return
	}

	// first step
	headers := map[string]string{"x-auth-login": user}
	resp := c.Post(t, path, nil, headers)
	CheckResponse(t, resp, &Expected{HttpCode: http.StatusUnauthorized, Error: auth_login_phash.ErrorCodeCredentialsRequired})

	salt := resp.Object.Header().Get("x-auth-login-salt")
	require.NotEmpty(t, salt)

	// second
	phash := auth_login_phash.Phash(password, salt)
	headers["x-auth-login-phash"] = phash
	resp = c.Post(t, path, nil, headers)

	if errorCode != "" {
		CheckResponse(t, resp, &Expected{HttpCode: http.StatusUnauthorized, Error: errorCode})
	} else {
		CheckResponse(t, resp, &Expected{HttpCode: http.StatusOK})
		assert.Empty(t, resp.Message)
		c.AccessToken = resp.Object.Header().Get("x-auth-access-token")
		require.NotEmpty(t, c.AccessToken)
		c.RefreshToken = resp.Object.Header().Get("x-auth-refresh-token")
		require.NotEmpty(t, c.RefreshToken)
		assert.NotEqual(t, c.AccessToken, c.RefreshToken)
	}
}

func (c *HttpClient) addTokens(headers ...map[string]string) map[string]string {

	h := map[string]string{}
	if c.AccessToken != "" {
		h["x-auth-access-token"] = c.AccessToken
	}
	if c.CsrfToken != "" {
		h["x-csrf"] = c.CsrfToken
	}
	if len(headers) > 0 {
		utils.AppendMap(h, headers[0])
	}
	return h
}

func (c *HttpClient) updateToken(resp *httptest.ResponseRecorder, code int) {
	accessToken := resp.Header().Get("x-auth-access-token")
	if accessToken != "" {
		c.AccessToken = accessToken
	}
	refreshToken := resp.Header().Get("x-auth-refresh-token")
	if refreshToken != "" {
		c.RefreshToken = refreshToken
	}
	csrfToken := resp.Header().Get("x-csrf")
	if csrfToken != "" {
		c.CsrfToken = csrfToken
	}
}

func (c *HttpClient) RequestBody(t *testing.T, method string, path string, cmd interface{}, headers ...map[string]string) *HttpResponse {
	h := c.addTokens(headers...)
	c.addTokens(headers...)
	resp, code, message := HttpRequestBody(t, c.Gin, method, c.Url(path), cmd, h)
	c.updateToken(resp, code)
	return &HttpResponse{resp, code, message}
}

func (c *HttpClient) RequestQuery(t *testing.T, method string, path string, cmd interface{}, headers ...map[string]string) *HttpResponse {
	h := c.addTokens(headers...)
	resp, code, message := HttpRequestQuery(t, c.Gin, method, c.Url(path), cmd, h)
	c.updateToken(resp, code)
	return &HttpResponse{resp, code, message}
}

func (c *HttpClient) Post(t *testing.T, path string, cmd interface{}, headers ...map[string]string) *HttpResponse {
	return c.RequestBody(t, http.MethodPost, path, cmd, headers...)
}

func (c *HttpClient) Put(t *testing.T, path string, cmd interface{}, headers ...map[string]string) *HttpResponse {
	return c.RequestBody(t, http.MethodPut, path, cmd, headers...)
}

func (c *HttpClient) Patch(t *testing.T, path string, cmd interface{}, headers ...map[string]string) *HttpResponse {
	return c.RequestBody(t, http.MethodPatch, path, cmd, headers...)
}

func (c *HttpClient) Get(t *testing.T, path string, cmd interface{}, headers ...map[string]string) *HttpResponse {
	return c.RequestQuery(t, http.MethodGet, path, cmd, headers...)
}

func (c *HttpClient) Delete(t *testing.T, path string, cmd interface{}, headers ...map[string]string) *HttpResponse {
	return c.RequestQuery(t, http.MethodDelete, path, cmd, headers...)
}

func (c *HttpClient) Logout(t *testing.T) {
	c.Post(t, "/auth/logout", nil)
}

func (c *HttpClient) RequestRefreshToken(t *testing.T, expectedErrorCode ...string) {

	errorCode := utils.OptionalArg("", expectedErrorCode...)

	h := map[string]string{"x-auth-refresh-token": c.RefreshToken}
	resp := c.Post(t, "/auth/refresh", nil, h)

	if errorCode != "" {
		CheckResponse(t, resp, &Expected{Error: errorCode, HttpCode: http.StatusUnauthorized})
	} else {
		CheckResponse(t, resp, &Expected{HttpCode: http.StatusOK})
		assert.Empty(t, resp.Message)
		c.AccessToken = resp.Object.Header().Get("x-auth-access-token")
		require.NotEmpty(t, c.AccessToken)
	}
}

func (c *HttpClient) Prepare(t *testing.T) {
	resp := c.Get(t, "/status/check", nil)
	CheckResponse(t, resp, &Expected{
		HttpCode: http.StatusOK,
		Message:  `{"status":"running"}`})
}
