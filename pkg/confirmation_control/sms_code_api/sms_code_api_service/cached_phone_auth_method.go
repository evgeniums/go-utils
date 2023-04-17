package sms_code_api_service

import (
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/confirmation_control/confirmation_control_api"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

type User struct {
	auth.UserBase
	phone string
}

func (u *User) Phone() string {
	return u.phone
}

const CachedPhoneAuthProtocol = "cached_phone"

type CachedPhoneAuthMethod struct {
	auth.AuthHandlerBase
}

func NewCachedPhoneAuthMethod() *CachedPhoneAuthMethod {
	c := &CachedPhoneAuthMethod{}
	return c
}

func (a *CachedPhoneAuthMethod) Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {

	a.AuthHandlerBase.Init(CachedPhoneAuthProtocol)

	return nil
}

func (a *CachedPhoneAuthMethod) Handle(ctx auth.AuthContext) (bool, error) {

	// setup
	c := ctx.TraceInMethod("CachedPhoneAuthMethod.Handle")
	defer ctx.TraceOutMethod()

	// get token from cache
	cacheToken, err := confirmation_control_api.GetTokenFromCache(ctx)
	if err != nil {
		return true, c.SetError(err)
	}

	// set pseudo user to context
	user := &User{}
	user.UserId = cacheToken.Id
	user.UserDisplay = cacheToken.Id
	user.UserLogin = "_none_"
	user.phone = cacheToken.Recipient
	ctx.SetAuthUser(user)

	// done
	return true, nil
}

func (a *CachedPhoneAuthMethod) SetAuthManager(manager auth.AuthManager) {
	manager.Schemas().AddHandler(a)
}
