package crud

import (
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type CRUD interface {
	Create(ctx op_context.Context, object common.Object) error
	Read(ctx op_context.Context, fields db.Fields, object interface{}, dest ...interface{}) (bool, error)
	ReadByField(ctx op_context.Context, fieldName string, fieldValue interface{}, object interface{}, dest ...interface{}) (bool, error)
	Update(ctx op_context.Context, object common.Object, fields db.Fields) error
	Delete(ctx op_context.Context, object common.Object) error

	List(ctx op_context.Context, filter *db.Filter, object interface{}, dest ...interface{}) error
}

type WithCRUD interface {
	CRUD() CRUD
}

type DbCRUD struct {
}

func (crud *DbCRUD) Create(ctx op_context.Context, object common.Object) error {
	c := ctx.TraceInMethod("CRUD.Create")
	defer ctx.TraceOutMethod()

	err := op_context.DB(ctx).Create(ctx, object)
	if err != nil {
		return c.SetError(err)
	}

	return nil
}

func (d *DbCRUD) Read(ctx op_context.Context, fields db.Fields, object interface{}, dest ...interface{}) (bool, error) {

	c := ctx.TraceInMethod("CRUD.Read")
	defer ctx.TraceOutMethod()

	found, err := op_context.DB(ctx).FindByFields(ctx, fields, object, dest...)
	if err != nil {
		return found, c.SetError(err)
	}

	return found, nil
}

func (d *DbCRUD) ReadByField(ctx op_context.Context, fieldName string, fieldValue interface{}, object interface{}, dest ...interface{}) (bool, error) {

	c := ctx.TraceInMethod("CRUD.Read", logger.Fields{fieldName: fieldValue})
	defer ctx.TraceOutMethod()

	found, err := op_context.DB(ctx).FindByField(ctx, fieldName, fieldValue, object, dest...)
	if err != nil {
		return found, c.SetError(err)
	}

	return found, nil
}

func (d *DbCRUD) Update(ctx op_context.Context, user common.Object, fields db.Fields) error {
	c := ctx.TraceInMethod("CRUD.Update")
	defer ctx.TraceOutMethod()

	err := db.Update(op_context.DB(ctx), ctx, user, fields)
	if err != nil {
		return c.SetError(err)
	}

	return nil
}

func (crud *DbCRUD) Delete(ctx op_context.Context, object common.Object) error {
	c := ctx.TraceInMethod("CRUD.Delete")
	defer ctx.TraceOutMethod()

	err := op_context.DB(ctx).DeleteByField(ctx, "id", object.GetID(), object)
	if err != nil {
		return c.SetError(err)
	}

	return nil
}

func (d *DbCRUD) List(ctx op_context.Context, filter *db.Filter, objects interface{}, dest ...interface{}) error {
	c := ctx.TraceInMethod("CRUD.List")
	defer ctx.TraceOutMethod()
	err := ctx.DB().FindWithFilter(ctx, filter, objects, dest...)
	if err != nil {
		return c.SetError(err)
	}
	return nil
}

func List(crud CRUD, ctx op_context.Context, methodName string, filter *db.Filter, objects interface{}, dest ...interface{}) error {

	c := ctx.TraceInMethod(methodName)
	defer ctx.TraceOutMethod()

	err := crud.List(ctx, filter, objects)
	if err != nil {
		return c.SetError(err)
	}

	return nil
}

func Find[T common.Object](crud CRUD, ctx op_context.Context, methodName string, fields db.Fields, object T, dest ...interface{}) (T, error) {
	c := ctx.TraceInMethod(methodName)
	defer ctx.TraceOutMethod()

	var nilT T

	found, err := crud.Read(ctx, fields, object, dest...)
	if err != nil {
		return nilT, c.SetError(err)
	}
	if !found {
		return nilT, nil
	}
	return object, nil
}

func FindByField[T common.Object](crud CRUD, ctx op_context.Context, methodName string, fieldName string, fieldValue interface{}, object T, dest ...interface{}) (T, error) {
	c := ctx.TraceInMethod(methodName)
	defer ctx.TraceOutMethod()

	var nilT T

	found, err := crud.ReadByField(ctx, fieldName, fieldValue, object, dest...)
	if err != nil {
		return nilT, c.SetError(err)
	}
	if !found {
		return nilT, nil
	}
	return object, nil
}

func Update(crud CRUD, ctx op_context.Context, methodName string, obj common.Object, fields db.Fields, loggerFields ...logger.Fields) error {
	c := ctx.TraceInMethod(methodName, loggerFields...)
	defer ctx.TraceOutMethod()
	err := crud.Update(ctx, obj, fields)
	if err != nil {
		return c.SetError(err)
	}
	return nil
}

func FindUpdate[T common.Object](crud CRUD, ctx op_context.Context, methodName string, fieldName string, fieldValue interface{}, fields db.Fields, object T, dest ...interface{}) error {
	c := ctx.TraceInMethod(methodName)
	defer ctx.TraceOutMethod()

	obj, err := FindByField(crud, ctx, "Find", fieldName, fieldValue, object)
	if err != nil {
		return c.SetError(err)
	}

	err = Update(crud, ctx, "Update", obj, fields)
	if err != nil {
		return c.SetError(err)
	}

	return nil
}
