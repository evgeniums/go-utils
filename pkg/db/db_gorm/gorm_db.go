package db_gorm

import (
	"errors"
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type gormDBConfig struct {
	PROVIDER string `default:"postgres"`
	HOST     string `default:"127.0.0.1"`
	PORT     uint16 `default:"5432"`
	USER     string
	DBNAME   string
	PASSWORD string `mask:""`

	ENABLE_DEBUG   bool
	VERBOSE_ERRORS bool
}

type DbConnector = func(provider string, dsn string) (gorm.Dialector, error)

type GormDB struct {
	db *gorm.DB
	gormDBConfig
	dbConnector DbConnector
}

func (g *GormDB) Config() interface{} {
	return &g.gormDBConfig
}

func DefaultDsnConnector(provider string, dsn string) (gorm.Dialector, error) {

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

func New(dbConnector ...DbConnector) *GormDB {
	g := &GormDB{}
	g.dbConnector = utils.OptionalArg(DefaultDsnConnector, dbConnector...)
	return g
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

func (g *GormDB) NewDB() db.DB {
	return New(g.dbConnector)
}

func (g *GormDB) Init(ctx logger.WithLogger, cfg config.Config, vld validator.Validator, configPath ...string) error {

	ctx.Logger().Info("Init GormDB")

	// load configuration
	err := object_config.LoadLogValidate(cfg, ctx.Logger(), vld, g, "psql", configPath...)
	if err != nil {
		return ctx.Logger().Fatal("failed to load GormDB configuration", err)
	}

	// connect database
	return g.Connect(ctx)
}

func (g *GormDB) InitWithConfig(ctx logger.WithLogger, vld validator.Validator, cfg *db.DBConfig) error {

	ctx.Logger().Info("Connect GormDB with DBConfig")

	// convert configuration
	g.gormDBConfig = gormDBConfig{PROVIDER: cfg.DbProvider, HOST: cfg.DbHost, PORT: cfg.DbPort,
		USER: cfg.DbLogin, PASSWORD: cfg.DbPassword}

	// validate configuration
	err := vld.Validate(g.Config())
	if err != nil {
		return ctx.Logger().Fatal("failed to validate GormDB configuration", err)
	}

	// connect database
	return g.Connect(ctx)
}

func (g *GormDB) Connect(ctx logger.WithLogger) error {

	// connect database
	dsn := fmt.Sprintf("host=%v port=%v user=%v dbname=%v password=%v sslmode=disable", g.HOST, g.PORT, g.USER, g.DBNAME, g.PASSWORD)
	dbDialector, err := g.dbConnector(g.PROVIDER, dsn)
	if err != nil {
		return ctx.Logger().Fatal("failed to connect to database", err)
	}
	g.db, err = ConnectDB(dbDialector)
	if err != nil {
		return ctx.Logger().Fatal("failed to connect to database", err)
	}

	// done
	return nil
}

func (g *GormDB) FindByField(ctx logger.WithLogger, field string, value string, obj interface{}) (bool, error) {
	notFound, err := FindByField(g.db_(), field, value, obj)
	if err != nil && g.VERBOSE_ERRORS && !notFound {
		e := fmt.Errorf("failed to FindByField %v", ObjectTypeName(obj))
		ctx.Logger().Error("GormDB", e, logger.Fields{"field": field, "value": value, "error": err})
	}
	return notFound, err
}

func (g *GormDB) FindByFields(ctx logger.WithLogger, fields map[string]interface{}, obj interface{}) (bool, error) {
	notFound, err := FindByFields(g.db_(), fields, obj)
	if err != nil && g.VERBOSE_ERRORS && !notFound {
		e := fmt.Errorf("failed to FindByFields %v", ObjectTypeName(obj))
		ctx.Logger().Error("GormDB", e, logger.Fields{"fields": fields, "error": err})
	}
	return notFound, err
}

func (g *GormDB) RowsByFields(ctx logger.WithLogger, fields map[string]interface{}, obj interface{}) (db.Cursor, error) {

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

func (g *GormDB) Create(ctx logger.WithLogger, obj common.Object) error {
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

func (g *GormDB) DeleteByFields(ctx logger.WithLogger, fields map[string]interface{}, obj interface{}) error {
	err := DeleteAllByFields(g.db_(), fields, obj)
	if err != nil && g.VERBOSE_ERRORS {
		e := fmt.Errorf("failed to DeleteByFields %v", ObjectTypeName(obj))
		ctx.Logger().Error("GormDB", e, logger.Fields{"fields": fields, "error": err})
	}
	return err
}

func (g *GormDB) UpdateFields(ctx logger.WithLogger, obj interface{}, fields map[string]interface{}) error {
	err := UpdateFields(g.db_(), fields, obj)
	if err != nil && g.VERBOSE_ERRORS {
		e := fmt.Errorf("failed to UpdateFields %v", ObjectTypeName(obj))
		ctx.Logger().Error("GormDB", e, logger.Fields{"error": err})
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

func (g *GormDB) FinWithFilter(ctx logger.WithLogger, filter *Filter, obj interface{}) (bool, error) {
	notFound, err := FindWithFilter(g.db_(), filter, obj)
	if err != nil && g.VERBOSE_ERRORS && !notFound {
		e := fmt.Errorf("failed to FinWithFilter %v", ObjectTypeName(obj))
		ctx.Logger().Error("GormDB", e, logger.Fields{"error": err})
	}
	return notFound, err
}
