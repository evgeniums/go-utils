package admin

import "github.com/evgeniums/go-utils/pkg/user"

type OpLogAdmin struct {
	user.OpLogUser
}

func NewOplog() user.OpLogUserI {
	return &OpLogAdmin{}
}
