package customer

import (
	"github.com/evgeniums/go-backend-helpers/pkg/auth/auth_session"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
	"github.com/evgeniums/go-backend-helpers/pkg/user/user_session_default"
)

type Customer struct {
	user_session_default.User
	common.WithNameBase
	common.WithDescriptionBase
}

func NewCustomer() *Customer {
	c := &Customer{}
	return c
}

type CustomerSession struct {
	auth_session.SessionBase
}

func NewCustomerSession() *CustomerSession {
	return &CustomerSession{}
}

type CustomerSessionClient struct {
	user_session_default.UserSessionClient
}

func NewCustomerSessionClient() *CustomerSessionClient {
	return &CustomerSessionClient{}
}

func Name(name string, sample ...*Customer) user.SetUserFields[*Customer] {
	return func(ctx op_context.Context, customer *Customer) ([]user.CheckDuplicateField, error) {
		customer.SetName(name)
		return nil, nil
	}
}

func Description(description string, sample ...*Customer) user.SetUserFields[*Customer] {
	return func(ctx op_context.Context, customer *Customer) ([]user.CheckDuplicateField, error) {
		customer.SetDescription(description)
		return nil, nil
	}
}
