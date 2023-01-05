package access_control

import "github.com/evgeniums/go-backend-helpers/pkg/common"

type Role interface {
	common.WithName
}

type RoleBase struct {
	common.WithNameBase
}
