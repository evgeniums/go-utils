package noauth

import (
	"github.com/evgeniums/go-backend-helpers/pkg/auth"
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

const NoAuthProtocol = "noauth"

type NoAuth struct {
	auth.AuthHandlerBase
}

func (n *NoAuth) Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {

	n.AuthHandlerBase.Init(NoAuthProtocol)

	return nil
}

func (n *NoAuth) Handle(ctx auth.AuthContext) (bool, error) {

	ctx.TraceInMethod("NoAuth.Handle")
	defer ctx.TraceOutMethod()

	return true, nil
}

func (n *NoAuth) SetAuthManager(manager auth.AuthManager) {
	manager.Schemas().AddHandler(n)
}
