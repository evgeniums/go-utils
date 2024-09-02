package customer

import "github.com/evgeniums/go-utils/pkg/user"

type OpLogCustomer struct {
	user.OpLogUser
}

func NewOplog() user.OpLogUserI {
	return &OpLogCustomer{}
}
