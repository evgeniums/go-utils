package db

import (
	"errors"
	"time"

	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

type Fields = map[string]interface{}

type DBConfig struct {
	DbProvider    string `gorm:"index"`
	DbHost        string `gorm:"index"`
	DbPort        uint16 `gorm:"index"`
	DbLogin       string `gorm:"index"`
	DbPassword    string
	DbExtraConfig string
}

type DBHandlers interface {
	FindByField(ctx logger.WithLogger, field string, value string, obj interface{}) (bool, error)
	FindByFields(ctx logger.WithLogger, fields Fields, obj interface{}) (bool, error)
	FinWithFilter(ctx logger.WithLogger, filter *Filter, obj interface{}) (bool, error)

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

func CheckFound(notfound bool, err *error) bool {
	ok := *err == nil && !notfound
	if notfound {
		*err = errors.New("not found")
	}
	return ok
}

func CheckFoundNoError(notfound bool, err *error) bool {
	ok := *err == nil && !notfound
	if notfound {
		*err = nil
	}
	return ok
}

func CheckFoundDbError(notfound bool, err error) error {
	if err != nil && !notfound {
		return err
	}
	return nil
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
	PreconditionFields      map[string]interface{}
	IntervalFields          map[string]*Interval
	PreconditionFieldsIn    map[string][]interface{}
	PreconditionFieldsNotIn map[string][]interface{}

	SortField     string
	SortDirection string
	Offset        int
	Limit         int
	In            []string
	Between       []*BetweenFields
}

func Update(db DBHandlers, ctx logger.WithLogger, obj common.Object, fields Fields) error {
	f := utils.CopyMap(fields)
	f["updated_at"] = time.Now()
	return db.Update(ctx, obj, Fields{"id": obj.GetID()}, f)
}
