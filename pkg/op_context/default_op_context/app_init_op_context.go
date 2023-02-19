package default_op_context

import (
	"net/http"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

func NewAppInitContext(app app_context.Context) op_context.Context {

	opCtx := NewContext()
	opCtx.Init(app, app.Logger(), app.Db())
	opCtx.SetName("init_app")
	errManager := &generic_error.ErrorManagerBase{}
	errManager.Init(http.StatusInternalServerError)
	opCtx.SetErrorManager(errManager)
	opCtx.SetWriteCloseLog(false)

	origin := NewOrigin(app)
	origin.SetUserType("auto_init")
	opCtx.SetOrigin(origin)

	return opCtx
}
