package user_manager

import (
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type UserController interface {
	Find(ctx op_context.Context, fields db.Fields, user interface{}) (bool, error)
	Create(ctx op_context.Context, user common.Object) error
	Update(ctx op_context.Context, user common.Object, fields db.Fields) error
	List(ctx op_context.Context, filter *db.Filter, users interface{}) error
}

type UserControllerBase struct {
}

func (m *UserControllerBase) Find(ctx op_context.Context, fields db.Fields, user interface{}) (bool, error) {

	c := ctx.TraceInMethod("UserControllerBase.Find")
	defer ctx.TraceOutMethod()

	found, err := op_context.DB(ctx).FindByFields(ctx, fields, user)
	if err != nil {
		return found, c.SetError(err)
	}

	return found, nil
}

func (m *UserControllerBase) Create(ctx op_context.Context, user common.Object) error {
	c := ctx.TraceInMethod("UserControllerBase.Create")
	defer ctx.TraceOutMethod()

	err := op_context.DB(ctx).Create(ctx, user)
	if err != nil {
		return c.SetError(err)
	}

	return nil
}

func (m *UserControllerBase) Update(ctx op_context.Context, user common.Object, fields db.Fields) error {
	c := ctx.TraceInMethod("UserControllerBase.Update")
	defer ctx.TraceOutMethod()

	err := db.Update(op_context.DB(ctx), ctx, user, fields)
	if err != nil {
		return c.SetError(err)
	}

	return nil
}

func (m *UserControllerBase) List(ctx op_context.Context, filter *db.Filter, users interface{}) error {
	return op_context.LoadObjects(ctx, "", filter, users)
}

func FindByLogin(controller UserController, ctx op_context.Context, login string, user interface{}) (bool, error) {
	return controller.Find(ctx, db.Fields{"login": login}, user)
}
