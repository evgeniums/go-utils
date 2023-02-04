package rest_api_client

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/evgeniums/go-backend-helpers/pkg/api_server/rest_api_gin_server"
	"github.com/evgeniums/go-backend-helpers/pkg/auth_methods/auth_login_phash"
	"github.com/evgeniums/go-backend-helpers/pkg/http_request"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/google/go-querystring/query"
)

type DoRequest = func(ctx op_context.Context, method string, path string, cmd interface{}, headers ...map[string]string) (Response, error)

func fillResponseError(resp Response) error {
	b := resp.Body()
	if b != nil {
		errResp := &rest_api_gin_server.ResponseError{}
		err := json.Unmarshal(b, errResp)
		if err != nil {
			return err
		}
		resp.SetError(errResp)
		return nil
	}
	return nil
}

type HttpClient struct {
	BaseUrl      string
	AccessToken  string
	RefreshToken string
	CsrfToken    string

	SendWithBody  DoRequest
	SendWithQuery DoRequest
}

func NewHttpClient(baseUrl string, withBodySender DoRequest, withQuerySender DoRequest) *HttpClient {
	c := &HttpClient{}
	c.BaseUrl = baseUrl
	c.SendWithBody = withBodySender
	c.SendWithQuery = withQuerySender
	return c
}

func (c *HttpClient) Url(path string) string {
	return c.BaseUrl + path
}

func (h *HttpClient) Login(ctx op_context.Context, user string, password string) (Response, error) {

	var err error
	c := ctx.TraceInMethod("HttpClient.Login", logger.Fields{"user": user})
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	path := "/auth/login"

	// first step
	headers := map[string]string{"x-auth-login": user}
	resp, err := h.Post(ctx, path, nil, headers)
	if err != nil {
		c.SetMessage("failed to send first request")
		return nil, err
	}
	if resp.Error().Code != auth_login_phash.ErrorCodeCredentialsRequired {
		err = errors.New("unexpected error code")
		c.SetLoggerField("error_code", resp.Error().Code)
		return resp, err
	}

	// second
	salt := resp.Header().Get("x-auth-login-salt")
	phash := auth_login_phash.Phash(password, salt)
	headers["x-auth-login-phash"] = phash
	resp, err = h.Post(ctx, path, nil, headers)
	if err != nil {
		c.SetMessage("failed to send second request")
		return nil, err
	}
	if resp.Code() != http.StatusOK {
		err = errors.New("login failed")
		c.SetLoggerField("error_code", resp.Error().Code)
		return resp, err
	}

	// done
	return resp, nil
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

func (c *HttpClient) updateTokens(resp Response) {
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

func (h *HttpClient) SendSmsConfirmation(send DoRequest, ctx op_context.Context, resp Response, code string, method string, path string, cmd interface{}, headers ...map[string]string) (Response, error) {

	c := ctx.TraceInMethod("HttpClient.SendSmsConfirmation")
	defer ctx.TraceOutMethod()

	hs := h.addTokens(headers...)
	hs["x-auth-sms-code"] = code
	token := resp.Header().Get("x-auth-sms-token")
	if token != "" {
		hs["x-auth-sms-token"] = token
	}
	nextResp, err := h.sendRequest(send, ctx, method, path, cmd, hs)
	if err != nil {
		return nil, c.SetError(err)
	}
	return nextResp, nil
}

func (h *HttpClient) sendRequest(send DoRequest, ctx op_context.Context, method string, path string, cmd interface{}, headers ...map[string]string) (Response, error) {

	c := ctx.TraceInMethod("HttpClient.sendRequest", logger.Fields{"method": method, "path": path})
	defer ctx.TraceOutMethod()

	hs := h.addTokens(headers...)
	h.addTokens(headers...)
	resp, err := send(ctx, method, h.Url(path), cmd, hs)
	if err != nil {
		c.SetMessage("failed to send request")
		return nil, c.SetError(err)
	}
	h.updateTokens(resp)

	err = fillResponseError(resp)
	if err != nil {
		c.SetMessage("failed to parse response error")
		return nil, c.SetError(err)
	}

	return resp, nil
}

func (h *HttpClient) RequestBody(ctx op_context.Context, method string, path string, cmd interface{}, headers ...map[string]string) (Response, error) {
	return h.sendRequest(h.SendWithBody, ctx, method, path, cmd, headers...)
}

func (h *HttpClient) RequestQuery(ctx op_context.Context, method string, path string, cmd interface{}, headers ...map[string]string) (Response, error) {
	return h.sendRequest(h.SendWithQuery, ctx, method, path, cmd, headers...)
}

func (h *HttpClient) Post(ctx op_context.Context, path string, cmd interface{}, headers ...map[string]string) (Response, error) {
	return h.RequestBody(ctx, http.MethodPost, path, cmd, headers...)
}

func (h *HttpClient) Put(ctx op_context.Context, path string, cmd interface{}, headers ...map[string]string) (Response, error) {
	return h.RequestBody(ctx, http.MethodPut, path, cmd, headers...)
}

func (h *HttpClient) Patch(ctx op_context.Context, path string, cmd interface{}, headers ...map[string]string) (Response, error) {
	return h.RequestBody(ctx, http.MethodPatch, path, cmd, headers...)
}

func (h *HttpClient) Get(ctx op_context.Context, path string, cmd interface{}, headers ...map[string]string) (Response, error) {
	return h.RequestQuery(ctx, http.MethodGet, path, cmd, headers...)
}

func (h *HttpClient) Delete(ctx op_context.Context, path string, cmd interface{}, headers ...map[string]string) (Response, error) {
	return h.RequestQuery(ctx, http.MethodGet, path, cmd, headers...)
}

func (h *HttpClient) Logout(ctx op_context.Context) (Response, error) {
	c := ctx.TraceInMethod("HttpClient.Logout")
	defer ctx.TraceOutMethod()
	resp, err := h.Post(ctx, "/auth/logout", nil)
	if err != nil {
		return nil, c.SetError(err)
	}
	return resp, nil
}

func (h *HttpClient) UpdateCsrfToken(ctx op_context.Context) (Response, error) {
	c := ctx.TraceInMethod("HttpClient.UpdateCsrfToken")
	defer ctx.TraceOutMethod()
	resp, err := h.Get(ctx, "/status/check", nil)
	if err != nil {
		return nil, c.SetError(err)
	}
	if resp.Code() != http.StatusOK {
		err = errors.New("failed to update CSRF")
		c.SetLoggerField("error_code", resp.Error().Code)
		return resp, err
	}
	return resp, nil
}

func (h *HttpClient) UpdateTokens(ctx op_context.Context) (Response, error) {

	c := ctx.TraceInMethod("HttpClient.UpdateTokens")
	defer ctx.TraceOutMethod()

	resp, err := h.UpdateCsrfToken(ctx)
	if err != nil {
		return resp, c.SetError(err)
	}

	resp, err = h.RequestRefreshToken(ctx)
	if err != nil {
		return resp, c.SetError(err)
	}

	return resp, nil
}

func (h *HttpClient) RequestRefreshToken(ctx op_context.Context) (Response, error) {

	c := ctx.TraceInMethod("HttpClient.RequestRefreshToken")
	defer ctx.TraceOutMethod()

	headers := map[string]string{"x-auth-refresh-token": h.RefreshToken}
	resp, err := h.Post(ctx, "/auth/refresh", nil, headers)
	if err != nil {
		return nil, c.SetError(err)
	}
	if resp.Code() != http.StatusOK {
		err = errors.New("failed to update CSRF")
		c.SetLoggerField("error_code", resp.Error().Code)
		return resp, err
	}
	return resp, nil
}

func (h *HttpClient) Prepare(ctx op_context.Context) (Response, error) {
	return h.UpdateCsrfToken(ctx)
}

func DefaultSendWithBody(ctx op_context.Context, method string, url string, cmd interface{}, headers ...map[string]string) (Response, error) {

	// setup
	var err error
	c := ctx.TraceInMethod("http_request.DefaultSendWithBody", logger.Fields{"method": method, "url": url})
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// prepare data
	cmdByte, err := json.Marshal(cmd)
	if err != nil {
		c.SetMessage("failed to marshal message")
		return nil, c.SetError(err)
	}

	// create request
	req, err := http.NewRequest(method, url, bytes.NewBuffer(cmdByte))
	if err != nil {
		c.SetMessage("failed to create request")
		return nil, c.SetError(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// send request
	rawResp, err := http_request.SendRawRequest(ctx, req)
	if err != nil {
		c.SetMessage("failed to send request")
		return nil, c.SetError(err)
	}

	// done
	resp := NewResponse(rawResp)
	return resp, nil
}

func DefaultSendWithQuery(ctx op_context.Context, method string, url string, cmd interface{}, headers ...map[string]string) (Response, error) {

	// setup
	var err error
	c := ctx.TraceInMethod("http_request.DefaultSendWithQuery", logger.Fields{"method": method, "url": url})
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// create request
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		c.SetMessage("failed to create request")
		return nil, c.SetError(err)
	}

	// prepare data
	v, err := query.Values(cmd)
	if err != nil {
		c.SetMessage("failed to build query")
		return nil, c.SetError(err)
	}
	req.URL.RawQuery = v.Encode()
	req.Header.Set("Accept", "application/json")

	// send request
	rawResp, err := http_request.SendRawRequest(ctx, req)
	if err != nil {
		c.SetMessage("failed to send request")
		return nil, c.SetError(err)
	}

	// done
	resp := NewResponse(rawResp)
	return resp, nil
}

func DefaultHttpClient(baseUrl string) *HttpClient {
	return NewHttpClient(baseUrl, DefaultSendWithBody, DefaultSendWithQuery)
}
