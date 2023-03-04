package db_gorm

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
	"gorm.io/gorm"
)

type baseDBConfig struct {
	ENABLE_DEBUG     bool
	VERBOSE_ERRORS   bool
	MAX_FILTER_LIMIT int `validate:"gte=0" vmessage:"Invalid max filter limit" default:"100"`
}

type gormDBConfig struct {
	db.DBConfig
	baseDBConfig
}

type DbConnector struct {
	DialectorOpener        func(provider string, dsn string) (gorm.Dialector, error)
	DsnBuilder             func(config *db.DBConfig) (string, error)
	DbCreator              func(provider string, db *gorm.DB, dbName string) error
	CheckDuplicateKeyError func(provider string, result *gorm.DB) (bool, error)
}

type GormDB struct {
	db *gorm.DB
	gormDBConfig
	dbConnector *DbConnector

	joinQueries   *db.JoinQueries
	filterManager *FilterManager
	paginator     *Paginator
}

func (g *GormDB) Config() interface{} {
	return &g.gormDBConfig
}

var DefaultDbConnector = PostgresDbConnector

func New(dbConnector ...*DbConnector) *GormDB {
	g := &GormDB{}

	g.filterManager = NewFilterManager()
	g.joinQueries = db.NewJoinQueries()

	g.dbConnector = DefaultDbConnector()
	g.paginator = &Paginator{}

	if len(dbConnector) != 0 {
		g.dbConnector = dbConnector[0]
	}

	return g
}

func (g *GormDB) ParseFilter(query *db.Query, parserName string) (*db.Filter, error) {
	return g.filterManager.ParseFilter(query, parserName)
}

func (g *GormDB) ParseFilterDirect(query *db.Query, model interface{}, parserName string, vld ...*db.FilterValidator) (*db.Filter, error) {
	return g.filterManager.ParseFilterDirect(query, model, parserName, vld...)
}

func (g *GormDB) PrepareFilterParser(model interface{}, name string, validator ...*db.FilterValidator) (db.FilterParser, error) {
	return g.filterManager.PrepareFilterParser(model, name, validator...)
}

func (g *GormDB) NativeHandler() interface{} {
	return g.db
}

func (g *GormDB) Clone() db.DB {
	d := New(g.dbConnector)
	d.baseDBConfig = g.baseDBConfig
	d.paginator.MaxLimit = g.MAX_FILTER_LIMIT
	return d
}

func (g *GormDB) EnableDebug(value bool) {
	g.ENABLE_DEBUG = value
}

func (g *GormDB) EnableVerboseErrors(value bool) {
	g.VERBOSE_ERRORS = value
}

func (g *GormDB) db_() *gorm.DB {
	if g.ENABLE_DEBUG {
		return g.db.Debug()
	}
	return g.db
}

func (g *GormDB) Init(ctx logger.WithLogger, cfg config.Config, vld validator.Validator, configPath ...string) error {

	ctx.Logger().Info("Init GormDB")

	// load configuration
	err := object_config.LoadLogValidate(cfg, ctx.Logger(), vld, g, "db", configPath...)
	if err != nil {
		return ctx.Logger().PushFatalStack("failed to load GormDB configuration", err)
	}
	g.paginator.MaxLimit = g.MAX_FILTER_LIMIT

	// connect database
	return g.Connect(ctx)
}

func (g *GormDB) InitWithConfig(ctx logger.WithLogger, vld validator.Validator, cfg *db.DBConfig) error {

	ctx.Logger().Info("Connect GormDB with DBConfig")

	// convert configuration
	g.gormDBConfig.DBConfig = *cfg

	// validate configuration
	err := vld.Validate(g.Config())
	if err != nil {
		return ctx.Logger().PushFatalStack("failed to validate GormDB configuration", err)
	}

	// connect database
	return g.Connect(ctx)
}

func (g *GormDB) Connect(ctx logger.WithLogger) error {

	var err error

	// connect database
	var dsn string
	if g.DB_DSN != "" {
		dsn = g.DB_DSN
	} else {
		dsn, err = g.dbConnector.DsnBuilder(&g.DBConfig)
		if err != nil {
			return ctx.Logger().PushFatalStack("failed to build DSN to connect to database", err)
		}
	}

	dbDialector, err := g.dbConnector.DialectorOpener(g.DB_PROVIDER, dsn)
	if err != nil {
		return ctx.Logger().PushFatalStack("failed open dialector to connect to database", err, logger.Fields{"db_provider": g.DB_PROVIDER})
	}

	g.db, err = ConnectDB(dbDialector)
	if err != nil {
		return ctx.Logger().PushFatalStack("failed to connect to database", err)
	}

	// done
	return nil
}

func (g *GormDB) Close() {
	if g.db != nil {
		db, err := g.db.DB()
		if err == nil && db != nil {
			db.Close()
		}
	}
}

func (g *GormDB) AutoMigrate(ctx logger.WithLogger, models []interface{}) error {
	err := g.db_().AutoMigrate(models...)
	if err != nil {
		return ctx.Logger().PushFatalStack("failed to migrate database", err)
	}
	return nil
}

func (g *GormDB) FindByField(ctx logger.WithLogger, field string, value interface{}, obj interface{}, dest ...interface{}) (bool, error) {
	notFound, err := FindByField(g.db_(), field, value, obj, dest...)
	if err != nil && g.VERBOSE_ERRORS && !notFound {
		e := fmt.Errorf("failed to FindByField %v", ObjectTypeName(obj))
		ctx.Logger().Error("GormDB", e, logger.Fields{"field": field, "value": value, "error": err})
	}
	return notFound, err
}

func (g *GormDB) FindByFields(ctx logger.WithLogger, fields db.Fields, obj interface{}, dest ...interface{}) (bool, error) {
	notFound, err := FindByFields(g.db_(), fields, obj, dest...)
	if err != nil && g.VERBOSE_ERRORS && !notFound {
		e := fmt.Errorf("failed to FindByFields %v", ObjectTypeName(obj))
		ctx.Logger().Error("GormDB", e, logger.Fields{"fields": fields, "error": err})
	}
	return notFound, err
}

func (g *GormDB) RowsByFields(ctx logger.WithLogger, fields db.Fields, obj interface{}) (db.Cursor, error) {

	var err error
	cursor := &GormCursor{gormDB: g}
	rows, err := RowsByFields(g.db_(), fields, obj)
	if err != nil && g.VERBOSE_ERRORS {
		e := fmt.Errorf("failed to RowsByFields %v", ObjectTypeName(obj))
		ctx.Logger().Error("GormDB", e, logger.Fields{"fields": fields, "error": err})
	}
	cursor.rows = rows
	cursor.sql = rows

	return cursor, err
}

func (g *GormDB) AllRows(ctx logger.WithLogger, obj interface{}) (db.Cursor, error) {

	var err error
	cursor := &GormCursor{gormDB: g}
	rows, err := AllRows(g.db_(), obj)
	if err != nil && g.VERBOSE_ERRORS {
		e := fmt.Errorf("failed to AllRows %v", ObjectTypeName(obj))
		ctx.Logger().Error("GormDB", e, logger.Fields{"error": err})
	}
	cursor.rows = rows
	cursor.sql = rows

	return cursor, err
}

func (g *GormDB) Create(ctx logger.WithLogger, obj interface{}) error {
	result := Create(g.db_(), obj)
	if result.Error != nil && g.VERBOSE_ERRORS {
		e := fmt.Errorf("failed to Create %v", ObjectTypeName(obj))
		ctx.Logger().Error("GormDB", e, logger.Fields{"error": result.Error})
	}
	return result.Error
}

func (g *GormDB) CreateDup(ctx logger.WithLogger, obj interface{}) (bool, error) {
	result := Create(g.db_(), obj)
	duplicate, err := g.dbConnector.CheckDuplicateKeyError(g.DB_PROVIDER, result)
	if err != nil && g.VERBOSE_ERRORS {
		e := fmt.Errorf("failed to Create %v", ObjectTypeName(obj))
		ctx.Logger().Error("GormDB", e, logger.Fields{"error": err})
	}
	return duplicate, err
}

func (g *GormDB) DeleteByField(ctx logger.WithLogger, field string, value interface{}, model interface{}) error {
	err := DeleteByField(g.db_(), field, value, model)
	if err != nil && g.VERBOSE_ERRORS {
		e := fmt.Errorf("failed to DeleteByField %v", ObjectTypeName(model))
		ctx.Logger().Error("GormDB", e, logger.Fields{"field": field, "value": value, "error": err})
	}
	return err
}

func (g *GormDB) Delete(ctx logger.WithLogger, obj common.Object) error {
	err := Delete(g.db_(), obj)
	if err != nil && g.VERBOSE_ERRORS {
		e := fmt.Errorf("failed to Delete %v", ObjectTypeName(obj))
		ctx.Logger().Error("GormDB", e, logger.Fields{"id": obj.GetID(), "error": err})
	}
	return err
}

func (g *GormDB) DeleteByFields(ctx logger.WithLogger, fields db.Fields, obj interface{}) error {
	err := DeleteAllByFields(g.db_(), fields, obj)
	if err != nil && g.VERBOSE_ERRORS {
		e := fmt.Errorf("failed to DeleteByFields %v", ObjectTypeName(obj))
		ctx.Logger().Error("GormDB", e, logger.Fields{"fields": fields, "error": err})
	}
	return err
}

func (g *GormDB) Transaction(handler db.TransactionHandler) error {

	nativeHandler := func(nativeTx *gorm.DB) error {
		tx := &GormDB{}
		tx.db = nativeTx
		return handler(tx)
	}

	return g.db.Transaction(nativeHandler)
}

func (g *GormDB) RowsWithFilter(ctx logger.WithLogger, filter *Filter, obj interface{}) (db.Cursor, error) {
	var err error
	cursor := &GormCursor{gormDB: g}
	rows, err := RowsWithFilter(g.db_(), filter, g.paginator, obj)
	if err != nil && g.VERBOSE_ERRORS {
		e := fmt.Errorf("failed to RowsWithFilter %v", ObjectTypeName(obj))
		ctx.Logger().Error("GormDB", e, logger.Fields{"error": err})
	}
	cursor.rows = rows
	cursor.sql = rows

	return cursor, err
}

func (g *GormDB) FindWithFilter(ctx logger.WithLogger, filter *Filter, obj interface{}, dest ...interface{}) (int64, error) {
	count, err := FindWithFilter(g.db_(), filter, g.paginator, obj, dest...)
	if err != nil && g.VERBOSE_ERRORS {
		e := fmt.Errorf("failed to FindWithFilter %v", ObjectTypeName(obj))
		ctx.Logger().Error("GormDB", e, logger.Fields{"error": err})
	}
	return count, err
}

func (g *GormDB) Update(ctx logger.WithLogger, obj interface{}, filter db.Fields, newFields db.Fields) error {
	err := UpdateFielsdMulti(g.db_(), filter, obj, newFields)
	if err != nil && g.VERBOSE_ERRORS {
		e := fmt.Errorf("failed to UpdateFieldsWithFilter %v", ObjectTypeName(obj))
		ctx.Logger().Error("GormDB", e, logger.Fields{"error": err})
	}
	return err
}

func (g *GormDB) UpdateAll(ctx logger.WithLogger, obj interface{}, newFields db.Fields) error {
	err := UpdateFieldsAll(g.db_(), obj, newFields)
	if err != nil && g.VERBOSE_ERRORS {
		e := fmt.Errorf("failed to UpdateAll %v", ObjectTypeName(obj))
		ctx.Logger().Error("GormDB", e, logger.Fields{"error": err})
	}
	return err
}

func (g *GormDB) Exists(ctx logger.WithLogger, filter *Filter, obj interface{}) (bool, error) {
	exists, err := Exists(g.db_(), filter, obj)
	if err != nil && g.VERBOSE_ERRORS {
		e := fmt.Errorf("failed to Exists %v", ObjectTypeName(obj))
		ctx.Logger().Error("GormDB", e, logger.Fields{"error": err})
	}
	return exists, err
}

func (g *GormDB) CreateDatabase(ctx logger.WithLogger, dbName string) error {
	err := g.dbConnector.DbCreator(g.DB_PROVIDER, g.db_(), dbName)
	if err != nil && g.VERBOSE_ERRORS {
		e := fmt.Errorf("failed to CreateDatabase %v", dbName)
		ctx.Logger().Error("GormDB", e, logger.Fields{"error": err})
	}
	return err
}
