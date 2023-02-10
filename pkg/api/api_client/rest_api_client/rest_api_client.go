package rest_api_client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/auth/auth_methods/auth_login_phash"
	"github.com/evgeniums/go-backend-helpers/pkg/http_request"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/google/go-querystring/query"
)

type RestApiClient interface {
	Url(path string) string

	Login(ctx op_context.Context, user string, password string) (Response, error)
	Logout(ctx op_context.Context) (Response, error)

	UpdateTokens(ctx op_context.Context) (Response, error)
	UpdateCsrfToken(ctx op_context.Context) (Response, error)
	RequestRefreshToken(ctx op_context.Context) (Response, error)

	Post(ctx op_context.Context, path string, cmd interface{}, response interface{}, headers ...map[string]string) (Response, error)
	Put(ctx op_context.Context, path string, cmd interface{}, response interface{}, headers ...map[string]string) (Response, error)
	Patch(ctx op_context.Context, path string, cmd interface{}, response interface{}, headers ...map[string]string) (Response, error)
	Get(ctx op_context.Context, path string, cmd interface{}, response interface{}, headers ...map[string]string) (Response, error)
	Delete(ctx op_context.Context, path string, cmd interface{}, response interface{}, headers ...map[string]string) (Response, error)

	SendSmsConfirmation(send DoRequest, ctx op_context.Context, resp Response, code string, method string, url string, cmd interface{}, headers ...map[string]string) (Response, error)
}

type DoRequest = func(ctx op_context.Context, method string, path string, cmd interface{}, headers ...map[string]string) (Response, error)

type RestApiClientBase struct {
	BaseUrl   string
	UserAgent string

	AccessToken  string
	RefreshToken string
	CsrfToken    string

	SendWithBody  DoRequest
	SendWithQuery DoRequest
}

func NewRestApiClientBase(withBodySender DoRequest, withQuerySender DoRequest) *RestApiClientBase {
	r := &RestApiClientBase{}
	r.Construct(withBodySender, withQuerySender)
	return r
}

func (r *RestApiClientBase) Init(baseUrl string, userAgent string) {
	r.BaseUrl = baseUrl
	r.UserAgent = userAgent
}

func (r *RestApiClientBase) Construct(withBodySender DoRequest, withQuerySender DoRequest) {
	r.SendWithBody = withBodySender
	r.SendWithQuery = withQuerySender
}

func (r *RestApiClientBase) Url(path string) string {
	return utils.ConcatStrings(r.BaseUrl, path)
}

func (r *RestApiClientBase) SetRefreshToken(token string) {
	r.RefreshToken = token
}

func (r *RestApiClientBase) Login(ctx op_context.Context, user string, password string) (Response, error) {

	var err error
	c := ctx.TraceInMethod("RestApiClientBase.Login", logger.Fields{"user": user})
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
	resp, err := r.Post(ctx, path, nil, nil, headers)
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
	resp, err = r.Post(ctx, path, nil, nil, headers)
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

func (r *RestApiClientBase) addTokens(headers ...map[string]string) map[string]string {

	h := map[string]string{}
	if r.AccessToken != "" {
		h["x-auth-access-token"] = r.AccessToken
	}
	if r.CsrfToken != "" {
		h["x-csrf"] = r.CsrfToken
	}
	if len(headers) > 0 {
		utils.AppendMap(h, headers[0])
	}
	return h
}

func (r *RestApiClientBase) updateTokens(resp Response) {
	accessToken := resp.Header().Get("x-auth-access-token")
	if accessToken != "" {
		r.AccessToken = accessToken
	}
	refreshToken := resp.Header().Get("x-auth-refresh-token")
	if refreshToken != "" {
		r.RefreshToken = refreshToken
	}
	csrfToken := resp.Header().Get("x-csrf")
	if csrfToken != "" {
		r.CsrfToken = csrfToken
	}
}

func (r *RestApiClientBase) SendSmsConfirmation(send DoRequest, ctx op_context.Context, resp Response, code string, method string, path string, cmd interface{}, headers ...map[string]string) (Response, error) {

	c := ctx.TraceInMethod("RestApiClientBase.SendSmsConfirmation")
	defer ctx.TraceOutMethod()

	hs := r.addTokens(headers...)
	hs["x-auth-sms-code"] = code
	token := resp.Header().Get("x-auth-sms-token")
	if token != "" {
		hs["x-auth-sms-token"] = token
	}
	nextResp, err := r.sendRequest(send, ctx, method, path, cmd, nil, hs)
	if err != nil {
		return nil, c.SetError(err)
	}
	return nextResp, nil
}

func (r *RestApiClientBase) sendRequest(send DoRequest, ctx op_context.Context, method string, path string, cmd interface{}, response interface{}, headers ...map[string]string) (Response, error) {

	// setup
	c := ctx.TraceInMethod("RestApiClientBase.sendRequest", logger.Fields{"method": method, "path": path})
	defer ctx.TraceOutMethod()

	// prepare tokens
	hs := r.addTokens(headers...)
	if r.UserAgent != "" {
		hs["User-Agent"] = r.UserAgent
	}

	// send request
	resp, err := send(ctx, method, r.Url(path), cmd, hs)
	if err != nil {
		c.SetMessage("failed to send request")
		return nil, c.SetError(err)
	}
	r.updateTokens(resp)

	// fill good response
	if resp.Code() < http.StatusBadRequest {
		if response != nil {
			if resp.Code() < http.StatusBadRequest {

				b := resp.Body()
				if len(b) == 0 {
					return nil, c.SetError(errors.New("failed to parse empty response"))
				}

				err = json.Unmarshal(b, response)
				if err != nil {
					fmt.Printf("message: %s\n", err)
					return nil, c.SetError(errors.New("failed to parse response message"))
				}
			}
		}
	} else {
		// fill response error
		err = fillResponseError(resp)
		if err != nil {
			c.SetMessage("failed to parse response error")
			return resp, c.SetError(err)
		}
		ctx.SetGenericError(api.ResponseGenericError(resp.Error()))
	}

	// done
	return resp, nil
}

func (r *RestApiClientBase) RequestBody(ctx op_context.Context, method string, path string, cmd interface{}, response interface{}, headers ...map[string]string) (Response, error) {
	return r.sendRequest(r.SendWithBody, ctx, method, path, cmd, response, headers...)
}

func (r *RestApiClientBase) RequestQuery(ctx op_context.Context, method string, path string, cmd interface{}, response interface{}, headers ...map[string]string) (Response, error) {
	return r.sendRequest(r.SendWithQuery, ctx, method, path, cmd, response, headers...)
}

func (r *RestApiClientBase) Post(ctx op_context.Context, path string, cmd interface{}, response interface{}, headers ...map[string]string) (Response, error) {
	return r.RequestBody(ctx, http.MethodPost, path, cmd, response, headers...)
}

func (r *RestApiClientBase) Put(ctx op_context.Context, path string, cmd interface{}, response interface{}, headers ...map[string]string) (Response, error) {
	return r.RequestBody(ctx, http.MethodPut, path, cmd, response, headers...)
}

func (r *RestApiClientBase) Patch(ctx op_context.Context, path string, cmd interface{}, response interface{}, headers ...map[string]string) (Response, error) {
	return r.RequestBody(ctx, http.MethodPatch, path, cmd, response, headers...)
}

func (r *RestApiClientBase) Get(ctx op_context.Context, path string, cmd interface{}, response interface{}, headers ...map[string]string) (Response, error) {
	return r.RequestQuery(ctx, http.MethodGet, path, cmd, response, headers...)
}

func (r *RestApiClientBase) Delete(ctx op_context.Context, path string, cmd interface{}, response interface{}, headers ...map[string]string) (Response, error) {
	return r.RequestQuery(ctx, http.MethodGet, path, cmd, response, headers...)
}

func (r *RestApiClientBase) Logout(ctx op_context.Context) (Response, error) {
	c := ctx.TraceInMethod("RestApiClientBase.Logout")
	defer ctx.TraceOutMethod()
	resp, err := r.Post(ctx, "/auth/logout", nil, nil)
	if err != nil {
		return nil, c.SetError(err)
	}
	return resp, nil
}

func (r *RestApiClientBase) UpdateCsrfToken(ctx op_context.Context) (Response, error) {
	c := ctx.TraceInMethod("RestApiClientBase.UpdateCsrfToken")
	defer ctx.TraceOutMethod()
	resp, err := r.Get(ctx, "/status/check", nil, nil)
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

func (r *RestApiClientBase) UpdateTokens(ctx op_context.Context) (Response, error) {

	c := ctx.TraceInMethod("RestApiClientBase.UpdateTokens")
	defer ctx.TraceOutMethod()

	resp, err := r.UpdateCsrfToken(ctx)
	if err != nil {
		return resp, c.SetError(err)
	}

	resp, err = r.RequestRefreshToken(ctx)
	if err != nil {
		return resp, c.SetError(err)
	}

	return resp, nil
}

func (r *RestApiClientBase) RequestRefreshToken(ctx op_context.Context) (Response, error) {

	c := ctx.TraceInMethod("RestApiClientBase.RequestRefreshToken")
	defer ctx.TraceOutMethod()

	headers := map[string]string{"x-auth-refresh-token": r.RefreshToken}
	resp, err := r.Post(ctx, "/auth/refresh", nil, nil, headers)
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

func (r *RestApiClientBase) Prepare(ctx op_context.Context) (Response, error) {
	return r.UpdateCsrfToken(ctx)
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
	http_request.HttpHeadersSet(req, headers...)

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
	http_request.HttpHeadersSet(req, headers...)

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

func DefaultRestApiClient(baseUrl string, userAgent ...string) *RestApiClientBase {
	c := NewRestApiClientBase(DefaultSendWithBody, DefaultSendWithQuery)
	c.Init(baseUrl, utils.OptionalArg("go-backend-helpers", userAgent...))
	return c
}

func fillResponseError(resp Response) error {
	b := resp.Body()
	if b != nil {
		errResp := &api.ResponseError{}
		err := json.Unmarshal(b, errResp)
		if err != nil {
			return err
		}
		resp.SetError(errResp)
		return nil
	}
	return nil
}

type ClientBase struct {
	BASE_URL   string `validate:"required"`
	USER_AGENT string `default:"go-backend-helpers"`
}
