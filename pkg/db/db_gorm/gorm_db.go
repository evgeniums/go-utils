package db_gorm

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/orm"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
	"gorm.io/gorm"
)

type GormDBConfig struct {
	host     string `config:"host" default:"127.0.0.1"`
	port     string `config:"port" default:"5432"`
	user     string `config:"user"`
	dbname   string `config:"dbname"`
	password string `config:"password"`

	debug    bool `config:"enable_debug"`
	logError bool `config:"verbose_errors"`
}

type GormDB struct {
	logger.WithLoggerBase

	db *gorm.DB

	GormDBConfig
}

func New(log logger.Logger) *GormDB {
	g := &GormDB{}
	g.SetLogger(log)
	return g
}

func (g *GormDB) EnableDebug(value bool) {
	g.debug = value
}

func (g *GormDB) EnableVerboseErrors(value bool) {
	g.logError = value
}

func (g *GormDB) db_() *gorm.DB {
	if g.debug {
		return g.db.Debug()
	}
	return g.db
}

func (g *GormDB) Init(cfg config.Config, vld validator.Validator, configPath ...string) error {

	g.Logger().Info("Init GormDB")

	// load configuration
	err := config.InitObject(cfg, g.Logger(), vld, g, "psql", configPath...)
	if err != nil {
		return g.Logger().Fatal("failed to load GormDB configuration", err)
	}

	// connect database
	dsn := fmt.Sprintf("host=%v port=%v user=%v dbname=%v password=%v sslmode=disable", g.host, g.port, g.user, g.dbname, g.password)
	g.db, err = ConnectDB(dsn)
	if err != nil {
		return g.Logger().Fatal("failed to connect to database", err)
	}

	// done
	return nil
}

func (g *GormDB) FindByField(field string, value string, obj interface{}) (bool, error) {
	notFound, err := FindByField(g.db_(), field, value, obj)
	if err != nil && g.logError && !notFound {
		e := fmt.Errorf("failed to FindByField %v", ObjectTypeName(obj))
		g.Logger().Error("GormDB", e, logger.Fields{"field": field, "value": value, "error": err})
	}
	return notFound, err
}

func (g *GormDB) FindByFields(fields map[string]interface{}, obj interface{}) (bool, error) {
	notFound, err := FindByFields(g.db_(), fields, obj)
	if err != nil && g.logError && !notFound {
		e := fmt.Errorf("failed to FindByFields %v", ObjectTypeName(obj))
		g.Logger().Error("GormDB", e, logger.Fields{"fields": fields, "error": err})
	}
	return notFound, err
}

func (g *GormDB) RowsByFields(fields map[string]interface{}, obj interface{}) (db.Cursor, error) {

	var err error
	cursor := &GormCursor{gormDB: g}
	rows, err := RowsByFields(g.db_(), fields, obj)
	if err != nil && g.logError {
		e := fmt.Errorf("failed to RowsByFields %v", ObjectTypeName(obj))
		g.Logger().Error("GormDB", e, logger.Fields{"fields": fields, "error": err})
	}
	cursor.rows = rows
	cursor.sql = rows

	return cursor, err
}

func (g *GormDB) AllRows(obj interface{}) (db.Cursor, error) {

	var err error
	cursor := &GormCursor{gormDB: g}
	rows, err := AllRows(g.db_(), obj)
	if err != nil && g.logError {
		e := fmt.Errorf("failed to AllRows %v", ObjectTypeName(obj))
		g.Logger().Error("GormDB", e, logger.Fields{"error": err})
	}
	cursor.rows = rows
	cursor.sql = rows

	return cursor, err
}

func (g *GormDB) Create(obj orm.BaseInterface) error {
	err := Create(g.db_(), obj)
	if err != nil && g.logError {
		e := fmt.Errorf("failed to Create %v", ObjectTypeName(obj))
		g.Logger().Error("GormDB", e, logger.Fields{"error": err})
	}
	return err
}

func (g *GormDB) DeleteByField(field string, value interface{}, obj interface{}) error {
	err := RemoveByField(g.db_(), field, value, obj)
	if err != nil && g.logError {
		e := fmt.Errorf("failed to DeleteByField %v", ObjectTypeName(obj))
		g.Logger().Error("GormDB", e, logger.Fields{"field": field, "value": value, "error": err})
	}
	return err
}

func (g *GormDB) DeleteByFields(fields map[string]interface{}, obj interface{}) error {
	err := DeleteAllByFields(g.db_(), fields, obj)
	if err != nil && g.logError {
		e := fmt.Errorf("failed to DeleteByFields %v", ObjectTypeName(obj))
		g.Logger().Error("GormDB", e, logger.Fields{"fields": fields, "error": err})
	}
	return err
}

func (g *GormDB) UpdateFields(obj interface{}, fields map[string]interface{}) error {
	err := UpdateFields(g.db_(), fields, obj)
	if err != nil && g.logError {
		e := fmt.Errorf("failed to UpdateFields %v", ObjectTypeName(obj))
		g.Logger().Error("GormDB", e, logger.Fields{"error": err})
	}
	return err
}

func (g *GormDB) Transaction(handler db.TransactionHandler) error {

	nativeHandler := func(nativeTx *gorm.DB) error {
		tx := &GormDB{}
		tx.WithLoggerBase = g.WithLoggerBase
		tx.db = nativeTx
		return handler(tx)
	}

	return g.db.Transaction(nativeHandler)
}

func (g *GormDB) RowsWithFilter(filter *Filter, obj interface{}) (db.Cursor, error) {
	var err error
	cursor := &GormCursor{gormDB: g}
	rows, err := RowsWithFilter(g.db_(), filter, obj)
	if err != nil && g.logError {
		e := fmt.Errorf("failed to RowsWithFilter %v", ObjectTypeName(obj))
		g.Logger().Error("GormDB", e, logger.Fields{"error": err})
	}
	cursor.rows = rows
	cursor.sql = rows

	return cursor, err
}

func (g *GormDB) FinWithFilter(filter *Filter, obj interface{}) (bool, error) {
	notFound, err := FindWithFilter(g.db_(), filter, obj)
	if err != nil && g.logError && !notFound {
		e := fmt.Errorf("failed to FinWithFilter %v", ObjectTypeName(obj))
		g.Logger().Error("GormDB", e, logger.Fields{"error": err})
	}
	return notFound, err
}
