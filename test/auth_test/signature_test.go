package auth_test

import (
	"net/http"
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/signature"
	"github.com/evgeniums/go-backend-helpers/pkg/signature/user_pubkey"
	"github.com/evgeniums/go-backend-helpers/pkg/test_utils"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
	"github.com/stretchr/testify/require"
)

func TestSignature(t *testing.T) {
	app, users, server, opCtx := initOpTest(t)
	defer app.Close()

	pubKeyBuilder := func() *UserPubKey { return &UserPubKey{} }
	pubkeyController := user_pubkey.NewPubkeyController[*UserPubKey, *User](pubKeyBuilder, server.SignatureManager(), nil)
	pubkeyController.AttachToErrorManager(server.ApiServer())
	pubKeyFinder := func(ctx auth.AuthContext) (signature.UserWithPubkey, error) {
		return user_pubkey.FindUserPubKey[*UserPubKey](pubkeyController, ctx)
	}
	server.SignatureManager().SetUserKeyFinder(pubKeyFinder)

	// create user1
	login1 := "user1@example.com"
	password1 := "password1"
	user1, err := users.Add(opCtx, login1, password1, user.Phone("12345678", &User{}), user.Email("user1@example.com", &User{}))
	require.NoErrorf(t, err, "failed to add user")
	require.NotNil(t, user1)

	// prepare client
	client := test_utils.PrepareHttpClient(t, test_utils.BBGinEngine(t, server))

	// good login
	client.Login(login1, password1)

	// send altered data or path
	t.Logf("No pubkey")
	cmd1 := &Cmd{Param1: "value1_1", Param2: "value1_2"}
	// cmd2 := &Cmd{Param1: "value2_1", Param2: "value2_2"}
	resp := client.Post("/status/echo", cmd1)
	test_utils.CheckResponse(t, resp, &test_utils.Expected{HttpCode: http.StatusUnauthorized, Error: auth.ErrorCodeUnauthorized})
	// smsToken := resp.Object.Header().Get("x-auth-sms-token")
	// smsDelay := resp.Object.Header().Get("x-auth-sms-delay")
	// t.Logf("Current delay: %s", smsDelay)
	// resp = client.SendSmsConfirmation(resp, auth_sms.LastSmsCode, http.MethodPost, "/status/sms", cmd2)
	// test_utils.CheckResponse(t, resp, &test_utils.Expected{HttpCode: http.StatusUnauthorized, Error: auth_sms.ErrorCodeContentMismatch})
	// headers := map[string]string{"x-auth-sms-token": smsToken}
	// resp = client.SendSmsConfirmation(resp, auth_sms.LastSmsCode, http.MethodPost, "/status/sms-alt", cmd1, headers)
	// test_utils.CheckResponse(t, resp, &test_utils.Expected{HttpCode: http.StatusUnauthorized, Error: auth_sms.ErrorCodeContentMismatch})
	// resp = client.SendSmsConfirmation(resp, auth_sms.LastSmsCode, http.MethodPut, "/status/sms", cmd1, headers)
	// test_utils.CheckResponse(t, resp, &test_utils.Expected{HttpCode: http.StatusUnauthorized, Error: auth_sms.ErrorCodeContentMismatch})
	// resp = client.SendSmsConfirmation(resp, auth_sms.LastSmsCode, http.MethodPost, "/status/sms", cmd1, headers)
	// test_utils.CheckResponse(t, resp, &test_utils.Expected{HttpCode: http.StatusOK})

}
