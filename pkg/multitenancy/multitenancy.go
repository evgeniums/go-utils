package multitenancy

import (
	"net/http"

	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/pubsub/pubsub_subscriber"
)

const (
	OpAdd            string = "add"
	OpDelete         string = "delete"
	OpActivate       string = "activate"
	OpDeactivate     string = "deactivate"
	OpSetPath        string = "set_path"
	OpSetRole        string = "set_role"
	OpSetCustomer    string = "set_customer"
	OpChangePoolOrDb string = "change_pool_or_db"
)

const (
	TENANCY_DATABASE_ROLE string = "tenancy_db"
)

const (
	ErrorCodeTenancyConflictRole           = "tenancy_conflict_role"
	ErrorCodeTenancyConflictPath           = "tenancy_conflict_path"
	ErrorCodeTenancyNotFound               = "tenancy_not_found"
	ErrorCodeTenancyDbInitializationFailed = "tenancy_db_initialization_failed"
	ErrorCodeForeignDatabase               = "foreign_tenancy_database"
	ErrorCodeNoDbserviceInPool             = "no_db_service_in_pool"
	ErrorCodeCreateTenancyDatabaseFailed   = "create_tenancy_database_dailed"
)

var ErrorDescriptions = map[string]string{
	ErrorCodeTenancyNotFound:               "Tenancy not found.",
	ErrorCodeTenancyConflictRole:           "Tenancy with such role already exists for that customer.",
	ErrorCodeTenancyConflictPath:           "Tenancy with such path already exists in that pool.",
	ErrorCodeTenancyDbInitializationFailed: "Failed to initialize tenancy database.",
	ErrorCodeForeignDatabase:               "Database does not belong to this tenancy.",
	ErrorCodeNoDbserviceInPool:             "Pool does not contain service for tenancy database.",
	ErrorCodeCreateTenancyDatabaseFailed:   "Failed to create tenancy database.",
}

var ErrorHttpCodes = map[string]int{
	ErrorCodeTenancyNotFound:               http.StatusNotFound,
	ErrorCodeNoDbserviceInPool:             http.StatusInternalServerError,
	ErrorCodeTenancyDbInitializationFailed: http.StatusInternalServerError,
	ErrorCodeForeignDatabase:               http.StatusInternalServerError,
	ErrorCodeCreateTenancyDatabaseFailed:   http.StatusInternalServerError,
}

type Multitenancy interface {

	// Check if multiple tenancies are enabled
	IsMultiTenancy() bool

	// Get all tenancies
	Tenancies() []Tenancy

	// Find tenancy by ID.
	Tenancy(id string) (Tenancy, error)

	// Find tenancy by path.
	TenancyByPath(path string) (Tenancy, error)

	// Load tenancy.
	LoadTenancy(ctx op_context.Context, id string) (Tenancy, error)

	// Unload tenancy.
	UnloadTenancy(id string)

	// Create tenancy
	CreateTenancy(ctx op_context.Context, data *TenancyData) (*TenancyItem, error)

	// Get tenancy controller.
	TenancyController() TenancyController

	// Close tenancies, e.g. close tenancy databases.
	Close()
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
	generic_error.ErrorsExtender
	Add(ctx op_context.Context, tenancy *TenancyData) (*TenancyItem, error)
	Find(ctx op_context.Context, id string, idIsDisplay ...bool) (*TenancyItem, error)
	List(ctx op_context.Context, filter *db.Filter) ([]*TenancyItem, int64, error)

	Exists(ctx op_context.Context, fields db.Fields) (bool, error)
	Delete(ctx op_context.Context, id string, withDb bool, idIsDisplay ...bool) error

	SetPath(ctx op_context.Context, id string, path string, idIsDisplay ...bool) error
	SetCustomer(ctx op_context.Context, id string, customerId string, idIsDisplay ...bool) error
	SetRole(ctx op_context.Context, id string, role string, idIsDisplay ...bool) error
	ChangePoolOrDb(ctx op_context.Context, id string, poolId string, dbName string, idIsDisplay ...bool) error
	Activate(ctx op_context.Context, id string, idIsDisplay ...bool) error
	Deactivate(ctx op_context.Context, id string, idIsDisplay ...bool) error
}
