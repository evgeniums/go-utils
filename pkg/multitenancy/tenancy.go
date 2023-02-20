package multitenancy

import (
	"errors"
	"strings"

	"github.com/evgeniums/go-backend-helpers/pkg/cache"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/pool"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

type Tenancy interface {
	common.Object
	common.WithActive
	common.WithDescription

	Path() string
	CustomerId() string
	CustomerDisplay() string
	Role() string
	DbName() string
	PoolId() string

	Db() db.DB
	Pool() pool.Pool
	Cache() cache.Cache
}

type WithTenancy interface {
	GetTenancy() Tenancy
}

type WithPath struct {
	PATH string `json:"path" gorm:"uniqueIndex" validate:"omitempty,alphanum_" vmessage:"Path must be alhanumeric"`
}

type WithRole struct {
	ROLE string `json:"role" gorm:"index;index:,unique,composite:u" validate:"required,alphanum_" vmessage:"Role must be alphanumeric not empty"`
}

type WithCustomerId struct {
	CUSTOMER_ID string `json:"customer_id" gorm:"index;index:,unique,composite:u" validate:"required,alphanum_|email" vmessage:"Invalid customer ID"`
}

type WithPoolAndDb struct {
	POOL_ID string `json:"pool_id" gorm:"index" validate:"omitempty,alphanum" vmessage:"Pool ID must be alhanumeric"`
	DBNAME  string `json:"dbname" gorm:"index;column:dbname" validate:"omitempty,alphanum_" vmessage:"Database name must be alhanumeric"`
}

type TenancyData struct {
	common.WithDescriptionBase
	WithPath
	WithRole
	WithCustomerId
	WithPoolAndDb
}

func (t *WithCustomerId) CustomerId() string {
	return t.CUSTOMER_ID
}

func (t *WithRole) Role() string {
	return t.ROLE
}

func (t *WithPath) Path() string {
	return t.PATH
}

func (t *WithPoolAndDb) PoolId() string {
	return t.POOL_ID
}

func (t *WithPoolAndDb) DbName() string {
	return t.DBNAME
}

func TenancyDisplay(t Tenancy) string {
	return utils.ConcatStrings(t.CustomerDisplay(), "/", t.Role())
}

func ParseTenancyDisplay(display string) (string, string, *validator.ValidationError) {
	parts := strings.Split(display, "/")
	if len(parts) != 2 {
		err := &validator.ValidationError{Message: "invalid format of tenancy ID"}
		return "", "", err
	}
	return parts[0], parts[1], nil
}

type TenancyDb struct {
	common.ObjectBase
	common.WithActiveBase
	TenancyData
}

func (TenancyDb) TableName() string {
	return "tenancies"
}

type TenancyItem struct {
	TenancyDb     `source:"tenancies"`
	CustomerLogin string `json:"customer_login" source:"customers.login"`
	PoolName      string `json:"pool_name" source:"pools.name"`
}

func (t *TenancyItem) CustomerDisplay() string {
	return t.CustomerLogin
}

func CheckTenancyDatabase(ctx op_context.Context, database db.DB, tenancyId string) error {

	c := ctx.TraceInMethod("CheckTenancyDatabase", logger.Fields{"tenancy_id": tenancyId})
	defer ctx.TraceOutMethod()

	filter := db.NewFilter()
	filter.AddField("id", tenancyId)
	exists, err := database.Exists(ctx, filter, &TenancyMeta{})
	if err != nil {
		return c.SetError(err)
	}
	if !exists {
		ctx.SetGenericErrorCode(ErrorCodeForeignDatabase)
		return c.SetError(errors.New("database does not belong to this tenancy"))
	}
	return nil
}
