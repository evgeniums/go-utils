package rest_api_client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
	"golang.org/x/term"
)

type InteractiveClientConfig struct {
	ClientBase
	TOKEN_FILE_NAME string `validate:"required,file"`
}

type InteractiveClient struct {
	app_context.WithAppBase
	InteractiveClientConfig
	*RestApiClientBase
}

func NewInteractiveClient(app app_context.Context) *InteractiveClient {
	c := &InteractiveClient{}
	c.WithAppBase.Init(app)
	c.RestApiClientBase = AutoReconnectRestApiClient(c)
	return c
}

func (a *InteractiveClient) Config() interface{} {
	return a.InteractiveClientConfig
}

func (a *InteractiveClient) Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {

	err := object_config.LoadLogValidate(cfg, log, vld, a, "rest_api_client", configPath...)
	if err != nil {
		return log.PushFatalStack("failed to load configuration of rest api client", err)
	}

	a.RestApiClientBase.Init(a.BASE_URL, a.USER_AGENT)

	return nil
}

func (e *InteractiveClient) GetRefreshToken() string {

	content, err := os.ReadFile(e.TOKEN_FILE_NAME)
	if err != nil || content == nil {
		e.App().Logger().Warn("failed to read refresh token from file", db.Fields{"error": err})
		return ""
	}

	tokenKeeper := &TokenKeeper{}
	err = json.Unmarshal(content, tokenKeeper)
	if err != nil {
		e.App().Logger().Warn("failed to unarshal refresh token from file", db.Fields{"error": err})
		return ""
	}

	return tokenKeeper.Token
}

func (e *InteractiveClient) SaveRefreshToken(ctx op_context.Context, token string) {

	c := ctx.TraceInMethod("rest_api_client.InteractiveClient")
	defer ctx.TraceOutMethod()

	tokenKeeper := &TokenKeeper{Token: token}

	content, err := json.MarshalIndent(tokenKeeper, "", " ")
	if err != nil {
		c.Logger().Error("failed to marshal refresh token", err)
	}

	err = os.WriteFile(e.TOKEN_FILE_NAME, content, 0644)
	if err != nil {
		c.Logger().Error("failed to save refresh token to file", err)
	}
}

func (e *InteractiveClient) GetCredentials(ctx op_context.Context) (string, string, error) {

	c := ctx.TraceInMethod("rest_api_client.InteractiveClient")
	defer ctx.TraceOutMethod()

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Please, enter login:")
	login, err := reader.ReadString('\n')
	if err != nil {
		c.SetMessage("failed to enter login")
		return "", "", c.SetError(err)
	}
	login = strings.TrimSuffix(login, "\n")

	fmt.Println("Please, enter password:")
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		c.SetMessage("failed to enter password")
		return "", "", c.SetError(err)
	}

	return login, string(password), nil
}
