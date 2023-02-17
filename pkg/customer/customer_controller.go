package customer

import (
	"net/http"

	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
)

const (
	ErrorCodeCustomerNotFound string = "customer_not_found"
)

var ErrorDescriptions = map[string]string{
	ErrorCodeCustomerNotFound: "Customer not found.",
}

var ErrorHttpCodes = map[string]int{
	ErrorCodeCustomerNotFound: http.StatusNotFound,
}

type CustomerFieldSetter interface {
	SetName(ctx op_context.Context, id string, name string, idIsLogin ...bool) error
	SetDescription(ctx op_context.Context, id string, description string, idIsLogin ...bool) error
}

type CustomerController interface {
	user.UserController[*Customer]
	CustomerFieldSetter
}

type CustomersControllerBase struct {
	*user.UserControllerBase[*Customer]
}

func (cu *CustomersControllerBase) SetName(ctx op_context.Context, id string, name string, idIsLogin ...bool) error {

	// setup
	ctx.SetLoggerField("name", name)
	c := ctx.TraceInMethod("Users.SetName")
	var err error
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// find user
	user, err := user.FindUser(cu.UserControllerBase, ctx, id, idIsLogin...)
	if err != nil {
		return err
	}

	// set name
	err = cu.CRUD().Update(ctx, user, db.Fields{"name": name})
	if err != nil {
		return err
	}

	// done
	cu.OpLog(ctx, "set_name", user.GetID(), user.Login())
	return nil
}

func (cu *CustomersControllerBase) SetDescription(ctx op_context.Context, id string, description string, idIsLogin ...bool) error {
	// setup
	c := ctx.TraceInMethod("Users.SetDescription")
	var err error
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	// find user
	user, err := user.FindUser(cu.UserControllerBase, ctx, id, idIsLogin...)
	if err != nil {
		return err
	}

	// set description
	err = cu.CRUD().Update(ctx, user, db.Fields{"description": description})
	if err != nil {
		return err
	}

	// done
	cu.OpLog(ctx, "set_description", user.GetID(), user.Login())
	return nil
}

func LocalCustomerController() *CustomersControllerBase {
	c := &CustomersControllerBase{}
	c.UserControllerBase = user.LocalUserController[*Customer]()
	return c
}
