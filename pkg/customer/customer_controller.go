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

type NameAndDescriptionSetter interface {
	SetName(ctx op_context.Context, id string, name string, idIsLogin ...bool) error
	SetDescription(ctx op_context.Context, id string, description string, idIsLogin ...bool) error
}

type UserNameAndDescriptionController[T user.User] interface {
	user.UserController[T]
	NameAndDescriptionSetter
}

type UserNameAndDescriptionControllerB[T user.User] struct {
	*user.UserControllerBase[T]
}

func (cu *UserNameAndDescriptionControllerB[T]) SetName(ctx op_context.Context, id string, name string, idIsLogin ...bool) error {

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

func (cu *UserNameAndDescriptionControllerB[T]) SetDescription(ctx op_context.Context, id string, description string, idIsLogin ...bool) error {
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
	c.SetUserBuilder(NewCustomer)
	c.SetOplogBuilder(NewOplog)
	return c
}

type CustomerController interface {
	UserNameAndDescriptionController[*Customer]
}

type CustomersControllerBase struct {
	UserNameAndDescriptionControllerB[*Customer]
}
