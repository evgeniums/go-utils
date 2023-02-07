package db_gorm

import (
	"errors"
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type gormDBConfig struct {
	db.DBConfig

	ENABLE_DEBUG   bool
	VERBOSE_ERRORS bool
}

type DbConnector struct {
	DialectorOpener func(provider string, dsn string) (gorm.Dialector, error)
	DsnBuilder      func(config *db.DBConfig) (string, error)
}

type GormDB struct {
	FilterManager
	db *gorm.DB
	gormDBConfig
	dbConnector *DbConnector
}

func (g *GormDB) Config() interface{} {
	return &g.gormDBConfig
}

func PostgresOpener(provider string, dsn string) (gorm.Dialector, error) {

	switch provider {
	case "postgres":
		return postgres.Open(dsn), nil
		// case "mysql":
		// 	return mysql.Open(dsn), nil
		// case "sqlite":
		// 	return sqlite.Open(dsn), nil
		// case "sqlserver":
		// 	return sqlserver.Open(dsn), nil
	}

	return nil, errors.New("unknown database provider")
}

func PostgresDsnBuilder(config *db.DBConfig) (string, error) {
	dsn := fmt.Sprintf("host=%v port=%v user=%v dbname=%v password=%v sslmode=disable", config.DB_HOST, config.DB_PORT, config.DB_USER, config.DB_NAME, config.DB_PASSWORD)
	return dsn, nil
}

func postgresDbConnector() *DbConnector {
	c := &DbConnector{}
	c.DialectorOpener = PostgresOpener
	c.DsnBuilder = PostgresDsnBuilder
	return c
}

var DefaultDbConnector = postgresDbConnector

func New(dbConnector ...*DbConnector) *GormDB {
	g := &GormDB{}
	g.FilterManager.Construct()

	g.dbConnector = DefaultDbConnector()

	if len(dbConnector) != 0 {
		g.dbConnector = dbConnector[0]
	}

	return g
}

func (g *GormDB) NewDB() db.DB {
	return New(g.dbConnector)
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

func (g *GormDB) FindByField(ctx logger.WithLogger, field string, value interface{}, obj interface{}) (bool, error) {
	notFound, err := FindByField(g.db_(), field, value, obj)
	if err != nil && g.VERBOSE_ERRORS && !notFound {
		e := fmt.Errorf("failed to FindByField %v", ObjectTypeName(obj))
		ctx.Logger().Error("GormDB", e, logger.Fields{"field": field, "value": value, "error": err})
	}
	return notFound, err
}

func (g *GormDB) FindByFields(ctx logger.WithLogger, fields db.Fields, obj interface{}) (bool, error) {
	notFound, err := FindByFields(g.db_(), fields, obj)
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
	err := Create(g.db_(), obj)
	if err != nil && g.VERBOSE_ERRORS {
		e := fmt.Errorf("failed to Create %v", ObjectTypeName(obj))
		ctx.Logger().Error("GormDB", e, logger.Fields{"error": err})
	}
	return err
}

func (g *GormDB) DeleteByField(ctx logger.WithLogger, field string, value interface{}, obj interface{}) error {
	err := RemoveByField(g.db_(), field, value, obj)
	if err != nil && g.VERBOSE_ERRORS {
		e := fmt.Errorf("failed to DeleteByField %v", ObjectTypeName(obj))
		ctx.Logger().Error("GormDB", e, logger.Fields{"field": field, "value": value, "error": err})
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
	rows, err := RowsWithFilter(g.db_(), filter, obj)
	if err != nil && g.VERBOSE_ERRORS {
		e := fmt.Errorf("failed to RowsWithFilter %v", ObjectTypeName(obj))
		ctx.Logger().Error("GormDB", e, logger.Fields{"error": err})
	}
	cursor.rows = rows
	cursor.sql = rows

	return cursor, err
}

func (g *GormDB) FindWithFilter(ctx logger.WithLogger, filter *Filter, obj interface{}) error {
	err := FindWithFilter(g.db_(), filter, obj)
	if err != nil && g.VERBOSE_ERRORS {
		e := fmt.Errorf("failed to FindWithFilter %v", ObjectTypeName(obj))
		ctx.Logger().Error("GormDB", e, logger.Fields{"error": err})
	}
	return err
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
