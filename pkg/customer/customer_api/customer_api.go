package customer_api

import (
	"github.com/evgeniums/go-backend-helpers/pkg/access_control"
	"github.com/evgeniums/go-backend-helpers/pkg/api"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/customer"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
)

type FieldsSetter[T customer.User] struct {
	user.UserFieldsSetterBase[T]
	common.WithNameBase
	common.WithDescriptionBase
}

func (c *FieldsSetter[T]) SetUserFields(ctx op_context.Context, user T) ([]user.CheckDuplicateField, error) {
	user.SetName(c.Name())
	user.SetDescription(c.Description())
	return c.UserFieldsSetterBase.SetUserFields(ctx, user)
}

func NewFieldsSetter[T customer.User]() user.UserFieldsSetter[T] {
	s := &FieldsSetter[T]{}
	return s
}

func SetName() api.Operation {
	return api.NewOperation("set_name", access_control.Put)
}

func SetDescription() api.Operation {
	return api.NewOperation("set_description", access_control.Put)
}
