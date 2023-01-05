package access_control

import (
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type Resource interface {
	common.WithNameAndPath
}

type ResourceManager interface {
	FindResource(ctx op_context.Context, path string) (Resource, error)
	ResourceTags(ctx op_context.Context, path string) ([]string, error)
}

type ResourceBase struct {
	common.WithNameAndPathBase
}
