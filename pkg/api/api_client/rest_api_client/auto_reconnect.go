package rest_api_client

import (
	"errors"
	"net/http"

	"github.com/evgeniums/go-backend-helpers/pkg/api/api_client"
	"github.com/evgeniums/go-backend-helpers/pkg/auth/auth_methods/auth_csrf"
	"github.com/evgeniums/go-backend-helpers/pkg/auth/auth_methods/auth_login_phash"
	"github.com/evgeniums/go-backend-helpers/pkg/auth/auth_methods/auth_token"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type autoReconnect struct {
	client   *RestApiClientBase
	handlers api_client.AutoReconnectHandlers

	inLogin bool
}

func newAutoReconnectHelper(handlers api_client.AutoReconnectHandlers) *autoReconnect {
	a := &autoReconnect{}
	a.handlers = handlers
	return a
}

func (a *autoReconnect) init() {
	token := a.handlers.GetRefreshToken()
	if token != "" {
		a.client.RefreshToken = token
	}
}

func (a *autoReconnect) resend(ctx op_context.Context, send func(opCtx op_context.Context) (Response, error)) (Response, error) {
	ctx.ClearError()
	resp, err := send(ctx)
	return a.checkResponse(ctx, send, resp, err)
}

func (a *autoReconnect) checkResponse(ctx op_context.Context, send func(opCtx op_context.Context) (Response, error), lastResp Response, lastErr error) (Response, error) {

	// setup
	var err error
	c := ctx.TraceInMethod("autoReconnect.checkResponse")
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// check last response
	if lastResp == nil || lastResp.Error() == nil {
		return lastResp, lastErr
	}

	// refresh CSRF token
	if lastResp.Code() == http.StatusForbidden && auth_csrf.IsCsrfError(lastResp.Error().Code) {
		resp, err := a.client.UpdateCsrfToken(ctx)
		if !IsResponseOK(resp, err) {
			return resp, err
		}
		return a.resend(ctx, send)
	}

	// only unauthorized errors can be processed
	if lastResp.Code() != http.StatusUnauthorized {
		err = errors.New(lastResp.Error().Message)
		return lastResp, err
	}

	// login
	if a.client.RefreshToken == "" || auth_token.ReloginRequired(lastResp.Error().Code) || lastResp.Error().Code == auth_login_phash.ErrorCodeLoginFailed {

		if a.inLogin {
			err = errors.New(lastResp.Error().Message)
			return lastResp, err
		}

		login, password, err := a.handlers.GetCredentials(ctx)
		if err != nil {
			c.SetMessage("failed to get credentials")
			return lastResp, err
		}
		if login == "" {
			err = errors.New("login must be specified in client credentials")
			return lastResp, err
		}
		a.inLogin = true
		resp, err := a.client.Login(ctx, login, password)
		a.inLogin = false
		if err != nil {
			c.SetMessage("failed to login")
			return resp, err
		}
		if resp == nil {
			err = errors.New("nil login response")
			return nil, err
		}
		a.handlers.SaveRefreshToken(ctx, a.client.RefreshToken)
		return a.resend(ctx, send)
	}

	// refresh token
	if a.client.AccessToken == "" || auth_token.RefreshRequired(lastResp.Error().Code) {
		resp, err := a.client.RequestRefreshToken(ctx)
		if !IsResponseOK(resp, err) {
			return resp, err
		}
		a.handlers.SaveRefreshToken(ctx, a.client.RefreshToken)
		return a.resend(ctx, send)
	}

	// done
	return lastResp, lastErr
}

func NewAutoReconnectRestApiClient(reconnectHandlers api_client.AutoReconnectHandlers) *RestApiClientWithConfig {

	reconnect := newAutoReconnectHelper(reconnectHandlers)
	var client *RestApiClientWithConfig

	sendWithBody := func(ctx op_context.Context, method string, url string, cmd interface{}, headers ...map[string]string) (Response, error) {
		send := func(opCtx op_context.Context) (Response, error) {
			hs := client.addTokens(headers...)
			return DefaultSendWithBody(opCtx, method, url, cmd, hs)
		}
		hs := client.addTokens(headers...)
		resp, err := DefaultSendWithBody(ctx, method, url, cmd, hs)
		return reconnect.checkResponse(ctx, send, resp, err)
	}
	sendWithQuery := func(ctx op_context.Context, method string, url string, cmd interface{}, headers ...map[string]string) (Response, error) {
		send := func(opCtx op_context.Context) (Response, error) {
			return DefaultSendWithQuery(opCtx, method, url, cmd, headers...)
		}
		resp, err := DefaultSendWithQuery(ctx, method, url, cmd, headers...)
		return reconnect.checkResponse(ctx, send, resp, err)
	}

	client = NewRestApiClientWithConfig(sendWithBody, sendWithQuery)
	reconnect.client = client.RestApiClientBase
	reconnect.init()
	return client
}
