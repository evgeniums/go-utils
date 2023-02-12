package customer

import "github.com/evgeniums/go-backend-helpers/pkg/user"

type OpLogCustomer struct {
	user.OpLogUser
}

func NewOplog() user.OpLogUserI {
	return &OpLogCustomer{}
}
