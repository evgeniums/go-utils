package test_utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/api_server/rest_api_gin_server"
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

func CheckResponse(t *testing.T, resp *httptest.ResponseRecorder, code int, message string) func(expected *Expected) {
	return func(expected *Expected) {
		if expected == nil {
			assert.Contains(t, []int{http.StatusOK, http.StatusCreated}, code)
		} else {
			if expected.HttpCode != 0 {
				assert.Equal(t, expected.HttpCode, code)
			}
			if expected.Error != "" {
				errResp := &rest_api_gin_server.ResponseError{}
				require.NoError(t, json.Unmarshal([]byte(message), errResp))
				assert.Equal(t, expected.Error, errResp.Code)
				if expected.Details != "" {
					assert.Equal(t, expected.Details, errResp.Details)
				}
				if expected.Message != "" {
					assert.Equal(t, expected.Message, errResp.Message)
				}
			} else {
				if expected.Message != "" {
					assert.Equal(t, expected.Message, message)
				}
			}
		}
	}
}

type HttpClient struct {
	BaseUrl      string
	AccessToken  string
	RefreshToken string
}

func (a *HttpClient) Url(path string) string {
	return a.BaseUrl + path
}

func (a *HttpClient) Login(t *testing.T, g *gin.Engine, user string, password string, expectedErrorCode ...string) {

	errorCode := utils.OptionalArg("", expectedErrorCode...)
	path := a.Url("/auth/login")

	checkInvalidResponse := func(resp *httptest.ResponseRecorder, code int, message string) {
		require.Equal(t, http.StatusUnauthorized, code)
		assert.NotEmpty(t, message)
		errResp := &rest_api_gin_server.ResponseError{}
		require.NoError(t, json.Unmarshal([]byte(message), errResp))
		assert.Equal(t, errorCode, errResp.Code)
	}

	if errorCode == auth_login_phash.ErrorCodeCredentialsRequired {
		// request without headers
		resp, code, message := HttpPost(t, g, path, nil)
		checkInvalidResponse(resp, code, message)
		return
	}

	// first step
	headers := map[string]string{"x-auth-login": user}
	resp, code, message := HttpPost(t, g, path, nil, headers)
	require.Equal(t, http.StatusUnauthorized, code)
	errResp := &rest_api_gin_server.ResponseError{}
	require.NoError(t, json.Unmarshal([]byte(message), errResp))
	assert.Equal(t, auth_login_phash.ErrorCodeCredentialsRequired, errResp.Code)

	salt := resp.Header().Get("x-auth-salt")
	require.NotEmpty(t, salt)

	// second
	phash := auth_login_phash.Phash(password, salt)
	headers["x-auth-phash"] = phash
	resp, code, message = HttpPost(t, g, path, nil, headers)

	if errorCode != "" {
		checkInvalidResponse(resp, code, message)
	} else {
		require.Equal(t, http.StatusOK, code)
		assert.Empty(t, message)
		a.AccessToken = resp.Header().Get("x-auth-access-token")
		require.NotEmpty(t, a.AccessToken)
		a.RefreshToken = resp.Header().Get("x-auth-refresh-token")
		require.NotEmpty(t, a.RefreshToken)
	}
}

func (a *HttpClient) addToken(headers ...map[string]string) map[string]string {

	h := map[string]string{"x-auth-access-token": a.AccessToken}
	if len(headers) > 0 {
		utils.AppendMap(h, headers[0])
	}
	return h
}

func (a *HttpClient) updateToken(resp *httptest.ResponseRecorder, code int) {
	accessToken := resp.Header().Get("x-auth-access-token")
	if accessToken != "" {
		a.AccessToken = accessToken
	}
	refreshToken := resp.Header().Get("x-auth-refresh-token")
	if refreshToken != "" {
		a.RefreshToken = refreshToken
	}
}

func (a *HttpClient) Post(t *testing.T, g *gin.Engine, path string, cmd interface{}, headers ...map[string]string) (*httptest.ResponseRecorder, int, string) {
	resp, code, message := HttpPost(t, g, path, cmd, a.addToken(headers...))
	a.updateToken(resp, code)
	return resp, code, message
}

func (a *HttpClient) Put(t *testing.T, g *gin.Engine, path string, cmd interface{}, headers ...map[string]string) (*httptest.ResponseRecorder, int, string) {
	resp, code, message := HttpPut(t, g, path, cmd, a.addToken(headers...))
	a.updateToken(resp, code)
	return resp, code, message
}

func (a *HttpClient) Patch(t *testing.T, g *gin.Engine, path string, cmd interface{}, headers ...map[string]string) (*httptest.ResponseRecorder, int, string) {
	resp, code, message := HttpPatch(t, g, path, cmd, a.addToken(headers...))
	a.updateToken(resp, code)
	return resp, code, message
}

func (a *HttpClient) Get(t *testing.T, g *gin.Engine, path string, cmd interface{}, headers ...map[string]string) (*httptest.ResponseRecorder, int, string) {
	resp, code, message := HttpGet(t, g, path, cmd, a.addToken(headers...))
	a.updateToken(resp, code)
	return resp, code, message
}

func (a *HttpClient) Delete(t *testing.T, g *gin.Engine, path string, cmd interface{}, headers ...map[string]string) (*httptest.ResponseRecorder, int, string) {
	resp, code, message := HttpDelete(t, g, path, cmd, a.addToken(headers...))
	a.updateToken(resp, code)
	return resp, code, message
}

func (a *HttpClient) Logout(t *testing.T, g *gin.Engine) {
	a.Post(t, g, "/auth/logout", nil)
}

func (a *HttpClient) RequestRefreshToken(t *testing.T, g *gin.Engine, expectedErrorCode ...string) {

	errorCode := utils.OptionalArg("", expectedErrorCode...)

	h := map[string]string{"x-auth-refresh-token": a.RefreshToken}
	resp, code, message := HttpPost(t, g, "/auth/refresh-token", nil, h)

	if errorCode != "" {
		require.Equal(t, http.StatusUnauthorized, code)
		assert.NotEmpty(t, message)
		errResp := &rest_api_gin_server.ResponseError{}
		require.NoError(t, json.Unmarshal([]byte(message), errResp))
		assert.Equal(t, errorCode, errResp.Code)
	} else {
		require.Equal(t, http.StatusOK, code)
		assert.Empty(t, message)
		a.AccessToken = resp.Header().Get("x-auth-access-token")
		require.NotEmpty(t, a.AccessToken)
		a.updateToken(resp, code)
	}
}
