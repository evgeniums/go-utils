package auth_login_phash

import (
	"errors"
	"net/http"

	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/crypt_utils"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

const LoginProtocol = "login_phash"
const LoginName = "login"
const SaltName = "login-salt"
const PasswordHashName = "login-phash"

// Auth handler for login processing. The AuthTokenHandler MUST ALWAYS follow this handler in session scheme with AND conjunction.
type LoginHandler struct {
	auth.AuthHandlerBase
}

func (l *LoginHandler) Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {

	l.AuthHandlerBase.Init(LoginProtocol)

	return nil
}

const ErrorCodeLoginFailed = "login_failed"
const ErrorCodeCredentialsRequired = "login_credentials_required"

func (a *LoginHandler) ErrorDescriptions() map[string]string {
	m := map[string]string{
		ErrorCodeLoginFailed:         "Invalid login or password",
		ErrorCodeCredentialsRequired: "Credentials hash must be provided in request",
	}
	return m
}

func (a *LoginHandler) ErrorProtocolCodes() map[string]int {
	m := map[string]int{
		ErrorCodeLoginFailed:         http.StatusUnauthorized,
		ErrorCodeCredentialsRequired: http.StatusUnauthorized,
	}
	return m
}

func (l *LoginHandler) Handle(ctx auth.AuthContext) (bool, error) {

	// setup
	c := ctx.TraceInMethod("LoginHandler.Handle")
	var err error
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// get login from request
	login := ctx.GetAuthParameter(l.Protocol(), LoginName)
	if login == "" {
		return false, nil
	}
	ctx.SetLoggerField("login", login)

	// load user
	notfound, err := ctx.LoadUser(login)
	if !db.CheckFoundNoError(notfound, &err) {
		if err != nil {
			c.SetMessage("failed to load user")
			ctx.SetGenericErrorCode(generic_error.ErrorCodeInternalServerError)
			return true, err
		}
		c.SetMessage("user not found")
		ctx.SetGenericErrorCode(ErrorCodeLoginFailed)
		return true, err
	}
	ctx.SetLoggerField("user", ctx.AuthUser().Display())

	// extract user salt
	salt := ctx.AuthUser().GetAuthParameter(l.Protocol(), "password_salt")

	// get password hash from request
	phash := ctx.GetAuthParameter(l.Protocol(), PasswordHashName)
	if phash != "" {

		// extract user password
		password := ctx.AuthUser().GetAuthParameter(l.Protocol(), "password")

		// check password hash
		hash := crypt_utils.NewHash()
		hash.CalcStrIn(login, password, salt)
		err = hash.CheckStr(phash)
		if err != nil {
			ctx.UnloadUser()
			c.SetMessage("invalid password hash")
			ctx.SetGenericErrorCode(ErrorCodeLoginFailed)
			return true, err
		}

		// done
		return true, nil
	}

	// add salt to auth parameters
	ctx.SetAuthParameter(l.Protocol(), SaltName, salt)
	ctx.SetGenericErrorCode(ErrorCodeCredentialsRequired)

	// done
	return true, errors.New("credentials not provided yet")
}
