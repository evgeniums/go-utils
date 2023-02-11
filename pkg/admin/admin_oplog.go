package admin

import "github.com/evgeniums/go-backend-helpers/pkg/user"

type OpLogAdmin struct {
	user.OpLogUser
}

func NewOplog() user.OpLogUserI {
	return &OpLogAdmin{}
}
