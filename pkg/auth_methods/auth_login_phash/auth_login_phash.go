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

type User interface {
	PasswordHash() string
	PasswordSalt() string
	SetPassword(password string, salt string) string
	CheckPasswordHash(phash string) bool
}

type UserBase struct {
	PASSWORD_HASH string
	PASSWORD_SALT string
}

func (u *UserBase) PasswordHash() string {
	return u.PASSWORD_HASH
}

func (u *UserBase) PasswordSalt() string {
	return u.PASSWORD_SALT
}

func (u *UserBase) SetPassword(password string) {
	u.PASSWORD_SALT = crypt_utils.GenerateString()
	u.PASSWORD_HASH = Phash(password, u.PASSWORD_SALT)
}

func (u *UserBase) CheckPasswordHash(phash string) bool {
	return crypt_utils.Check(u.PASSWORD_HASH, phash)
}

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

	// user must be of User interface
	user, ok := ctx.AuthUser().(User)
	if !ok {
		c.SetMessage("user must be of UserWithPasswordHash interface")
		ctx.SetGenericErrorCode(generic_error.ErrorCodeInternalServerError)
		return true, err
	}

	// extract user salt
	salt := user.PasswordSalt()

	// get password hash from request
	phash := ctx.GetAuthParameter(l.Protocol(), PasswordHashName)
	if phash != "" {

		// extract user password
		password := user.PasswordHash()

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

func Phash(password string, salt string) string {
	h := crypt_utils.NewHash()
	return h.CalcStrStr(salt, password)
}
