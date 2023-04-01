package auth

import (
	"github.com/evgeniums/go-backend-helpers/pkg/access_control"
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

const NoAuthProtocol = "noauth"

type NoAuthMethod struct {
	AuthHandlerBase
}

func (n *NoAuthMethod) Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {

	n.AuthHandlerBase.Init(NoAuthProtocol)

	return nil
}

func (n *NoAuthMethod) Handle(ctx AuthContext) (bool, error) {

	ctx.TraceInMethod("NoAuth.Handle")
	defer ctx.TraceOutMethod()

	user := &UserBase{}
	user.UserId = "0"
	user.UserDisplay = "_none_"
	user.UserLogin = "_noauth_"
	ctx.SetAuthUser(user)

	return true, nil
}

func (n *NoAuthMethod) SetAuthManager(manager AuthManager) {
	manager.Schemas().AddHandler(n)
}

type NoAuth struct {
	handler *NoAuthMethod
}

func NewNoAuth() *NoAuth {
	a := &NoAuth{}
	a.handler = &NoAuthMethod{}
	return a
}

func (a *NoAuth) HandleRequest(ctx AuthContext, path string, access access_control.AccessType) error {
	a.handler.Handle(ctx)
	return nil
}

func (a *NoAuth) AttachToErrorManager(errManager generic_error.ErrorManager) {
}
