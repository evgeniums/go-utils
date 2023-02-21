package customer_api

import (
	"github.com/evgeniums/go-backend-helpers/pkg/access_control"
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/customer"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
)

type CustomerFieldsSetter struct {
	user.UserFieldsSetterBase[*customer.Customer]
	common.WithNameBase
	common.WithDescriptionBase
}

func (c *CustomerFieldsSetter) SetUserFields(ctx op_context.Context, customer *customer.Customer) ([]user.CheckDuplicateField, error) {
	customer.SetName(c.Name())
	customer.SetDescription(c.Description())
	return c.UserFieldsSetterBase.SetUserFields(ctx, customer)
}

func NewCustomerFieldsSetter() user.UserFieldsSetter[*customer.Customer] {
	s := &CustomerFieldsSetter{}
	return s
}

func SetName() api.Operation {
	return api.NewOperation("set_name", access_control.Put)
}

func SetDescription() api.Operation {
	return api.NewOperation("set_description", access_control.Put)
}
