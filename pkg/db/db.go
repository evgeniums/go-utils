package db

import (
	"time"

	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

const (
	SORT_ASC  string = "ASC"
	SORT_DESC string = "DESC"
)

type Fields = map[string]interface{}

func IsFieldSet(f Fields, key string) bool {
	_, ok := f[key]
	return ok
}

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
	FindByField(ctx logger.WithLogger, field string, value interface{}, obj interface{}, dest ...interface{}) (found bool, err error)
	FindByFields(ctx logger.WithLogger, fields Fields, obj interface{}, dest ...interface{}) (found bool, err error)
	FindWithFilter(ctx logger.WithLogger, filter *Filter, docs interface{}, dest ...interface{}) (int64, error)
	Exists(ctx logger.WithLogger, filter *Filter, doc interface{}) (bool, error)

	Create(ctx logger.WithLogger, obj interface{}) error
	Delete(ctx logger.WithLogger, obj common.Object) error
	DeleteByField(ctx logger.WithLogger, field string, value interface{}, model interface{}) error
	DeleteByFields(ctx logger.WithLogger, fields Fields, obj interface{}) error

	RowsWithFilter(ctx logger.WithLogger, filter *Filter, obj interface{}) (Cursor, error)
	AllRows(ctx logger.WithLogger, obj interface{}) (Cursor, error)

	Update(ctx logger.WithLogger, obj interface{}, filter Fields, fields Fields) error
	UpdateAll(ctx logger.WithLogger, obj interface{}, newFields Fields) error

	Join(ctx logger.WithLogger, joinConfig *JoinQueryConfig, filter *Filter, dest interface{}) (int64, error)

	Joiner() Joiner

	CreateDatabase(ctx logger.WithLogger, dbName string) error
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
	WithFilterParser

	Clone() DB

	InitWithConfig(ctx logger.WithLogger, vld validator.Validator, cfg *DBConfig) error

	DBHandlers

	Transaction(handler TransactionHandler) error

	EnableDebug(bool)
	EnableVerboseErrors(bool)

	AutoMigrate(ctx logger.WithLogger, models []interface{}) error

	NativeHandler() interface{}

	Close()
}

type WithDB interface {
	Db() DB
}

type WithDBBase struct {
	db DB
}

func (w *WithDBBase) Db() DB {
	return w.db
}

func (w *WithDBBase) Init(db DB) {
	w.db = db
}

func Update(db DBHandlers, ctx logger.WithLogger, obj common.Object, fields Fields) error {
	f := utils.CopyMap(fields)
	f["updated_at"] = time.Now()
	return db.Update(ctx, obj, Fields{"id": obj.GetID()}, f)
}

func UpdateMulti(db DBHandlers, ctx logger.WithLogger, obj interface{}, filter Fields, fields Fields) error {
	f := utils.CopyMap(fields)
	f["updated_at"] = time.Now()
	return db.Update(ctx, obj, filter, f)
}

func UpdateAll(db DBHandlers, ctx logger.WithLogger, obj interface{}, fields Fields) error {
	f := utils.CopyMap(fields)
	f["updated_at"] = time.Now()
	return db.UpdateAll(ctx, obj, f)
}
