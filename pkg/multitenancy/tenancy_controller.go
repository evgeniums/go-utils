package multitenancy

import (
	"github.com/evgeniums/go-backend-helpers/pkg/crud"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub/pubsub_subscriber"
)

const (
	OpAdd        string = "add"
	OpDelete     string = "delete"
	OpUpdate     string = "update"
	OpActivate   string = "activate"
	OpDeactivate string = "deactivate"
)

type Multitenancy interface {

	// Check if multiple tenancies are enabled
	IsMultiTenancy() bool

	// Find tenancy by ID.
	Tenancy(id string) (Tenancy, error)

	// Find tenancy by path.
	TenancyByPath(path string) (Tenancy, error)

	// Load tenancy.
	LoadTenancy(ctx op_context.Context, id string) (Tenancy, error)

	// Unload tenancy.
	UnloadTenancy(id string)
}

type PubsubNotification struct {
	Tenancy   string `json:"tenancy"`
	Operation string `json:"operation"`
}

const PubsubTopicName = "tenancy"

type PubsubTopic struct {
	*pubsub_subscriber.TopicBase[*PubsubNotification]
}

func NewPubsubNotification() *PubsubNotification {
	return &PubsubNotification{}
}

type TenancyController interface {
	Add(ctx op_context.Context, tenancy *TenancyDb) (TenancyDb, error)
	Find(ctx op_context.Context, id string, idIsDisplay ...bool) (*TenancyDb, error)
	Update(ctx op_context.Context, id string, fields db.Fields, idIsDisplay ...bool) error
	Delete(ctx op_context.Context, id string, idIsDisplay ...bool) error
	List(ctx op_context.Context, filter *db.Filter) ([]*TenancyDb, int64, error)

	Activate(ctx op_context.Context, id string, idIsDisplay ...bool) error
	Deactivate(ctx op_context.Context, id string, idIsDisplay ...bool) error
}

type TenancyControllerBase struct {
	CRUD crud.CRUD
}

func NewTenancyController(crud crud.CRUD) *TenancyControllerBase {
	c := &TenancyControllerBase{}
	c.CRUD = crud
	return c
}
