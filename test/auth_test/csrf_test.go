package auth_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/evgeniums/go-backend-helpers/pkg/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestCsrf(t *testing.T) {
	app, server := initAuthServer(t)
	defer app.Close()

	client := test_utils.NewHttpClient(server.RestApiServer.GinEngine())

	resp := client.Get(t, "/just-check-404", nil)
	test_utils.CheckResponse(t, resp, &test_utils.Expected{
		Error:    "not_found",
		HttpCode: http.StatusNotFound,
		Message:  "Requested resource was not found"})
	assert.Empty(t, client.CsrfToken)
	assert.Empty(t, client.AccessToken)
	assert.Empty(t, client.RefreshToken)

	resp = client.Get(t, "/status/csrf", nil)
	test_utils.CheckResponse(t, resp, &test_utils.Expected{
		Error:    "anti_csrf_token_required",
		HttpCode: http.StatusForbidden,
		Message:  "Request must be protected with anti-CSRF token"})
	assert.Empty(t, client.CsrfToken)
	assert.Empty(t, client.AccessToken)
	assert.Empty(t, client.RefreshToken)

	resp = client.Get(t, "/status/check", nil)
	test_utils.CheckResponse(t, resp, &test_utils.Expected{
		HttpCode: http.StatusOK,
		Message:  `{"status":"running"}`})
	assert.NotEmpty(t, client.CsrfToken)
	assert.Empty(t, client.AccessToken)
	assert.Empty(t, client.RefreshToken)
	prevToken := client.CsrfToken

	resp = client.Get(t, "/status/csrf", nil)
	test_utils.CheckResponse(t, resp, &test_utils.Expected{
		HttpCode: http.StatusOK,
		Message:  `{"status":"success"}`})
	assert.NotEqual(t, prevToken, client.CsrfToken)

	t.Logf("Wait for token expiration for 3 seconds...")
	time.Sleep(time.Second * 3)

	resp = client.Get(t, "/status/csrf", nil)
	test_utils.CheckResponse(t, resp, &test_utils.Expected{
		Error:    "anti_csrf_token_expired",
		HttpCode: http.StatusForbidden,
		Message:  "Anti-CSRF token expired"})

	resp = client.Get(t, "/status/check", nil)
	test_utils.CheckResponse(t, resp, &test_utils.Expected{
		HttpCode: http.StatusOK,
		Message:  `{"status":"running"}`})

	resp = client.Get(t, "/status/csrf", nil)
	test_utils.CheckResponse(t, resp, &test_utils.Expected{
		HttpCode: http.StatusOK,
		Message:  `{"status":"success"}`})
}
