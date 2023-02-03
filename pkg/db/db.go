package db

import (
	"time"

	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

const (
	SORT_ASC  string = "asc"
	SORT_DESC string = "desc"
)

type Fields = map[string]interface{}

type DBConfig struct {
	DB_PROVIDER     string `gorm:"index"`
	DB_HOST         string `gorm:"index"`
	DB_PORT         uint16 `gorm:"index"`
	DB_NAME         string `gorm:"index"`
	DB_USER         string `gorm:"index"`
	DB_PASSWORD     string `mask:"true"`
	DB_EXTRA_CONFIG string
	DB_DSN          string
}

type DBHandlers interface {
	FindByField(ctx logger.WithLogger, field string, value string, obj interface{}) (found bool, err error)
	FindByFields(ctx logger.WithLogger, fields Fields, obj interface{}) (found bool, err error)
	FindWithFilter(ctx logger.WithLogger, filter *Filter, docs interface{}) error

	Create(ctx logger.WithLogger, obj interface{}) error
	DeleteByField(ctx logger.WithLogger, field string, value interface{}, obj interface{}) error
	DeleteByFields(ctx logger.WithLogger, fields Fields, obj interface{}) error

	RowsWithFilter(ctx logger.WithLogger, filter *Filter, obj interface{}) (Cursor, error)
	AllRows(ctx logger.WithLogger, obj interface{}) (Cursor, error)

	Update(ctx logger.WithLogger, obj interface{}, filter Fields, fields Fields) error
	UpdateAll(ctx logger.WithLogger, obj interface{}, newFields Fields) error
}

type Transaction interface {
	DBHandlers
}

type TransactionHandler = func(tx Transaction) error

type Cursor interface {
	Next(ctx logger.WithLogger) (bool, error)
	Close(ctx logger.WithLogger) error
	Scan(ctx logger.WithLogger, obj interface{}) error
}

type DB interface {
	NewDB() DB

	InitWithConfig(ctx logger.WithLogger, vld validator.Validator, cfg *DBConfig) error

	DBHandlers

	Transaction(handler TransactionHandler) error

	EnableDebug(bool)
	EnableVerboseErrors(bool)

	AutoMigrate(ctx logger.WithLogger, models []interface{}) error

	Close()
}

type WithDB interface {
	DB() DB
}

type WithDBBase struct {
	db DB
}

func (w *WithDBBase) DB() DB {
	return w.db
}

func (w *WithDBBase) Init(db DB) {
	w.db = db
}

type Interval struct {
	From interface{}
	To   interface{}
}

type BetweenFields struct {
	FromField string
	ToField   string
	Value     interface{}
}

func (i *Interval) IsNull() bool {
	return i.From == nil && i.To == nil
}

type Filter struct {
	Fields         Fields
	FieldsIn       map[string][]interface{}
	FieldsNotIn    map[string][]interface{}
	IntervalFields map[string]*Interval
	BetweenFields  []*BetweenFields

	SortField     string
	SortDirection string
	Offset        int
	Limit         int
}

func Update(db DBHandlers, ctx logger.WithLogger, obj common.Object, fields Fields) error {
	f := utils.CopyMap(fields)
	f["updated_at"] = time.Now()
	return db.Update(ctx, obj, Fields{"id": obj.GetID()}, f)
}
