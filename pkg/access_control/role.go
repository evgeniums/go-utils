package access_control

import "github.com/evgeniums/go-utils/pkg/common"

type Role interface {
	common.WithName
}

type RoleBase struct {
	common.WithNameBase
}
