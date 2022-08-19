package gorm_db

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/config"
	"github.com/evgeniums/go-backend-helpers/db"
	"github.com/evgeniums/go-backend-helpers/logger"
	"github.com/evgeniums/go-backend-helpers/orm"
	"gorm.io/gorm"
)

type GormDB struct {
	logger.WithLoggerBase
	config.WithConfigBase

	DB *gorm.DB

	debug bool
}

func (g *GormDB) EnableDebug(value bool) {
	g.debug = value
}

func (g *GormDB) db_() *gorm.DB {
	if g.debug {
		return g.DB.Debug()
	}
	return g.DB
}

func (g *GormDB) Init(configPath ...string) error {

	g.Logger().Info("Init GormDB")

	path := "psql"
	if len(configPath) == 1 {
		path = configPath[0]
	}

	config.Key(path, "host")

	dsn := fmt.Sprintf("host=%v port=%v user=%v dbname=%v sslmode=disable",
		g.Config().GetString(config.Key(path, "host")),
		g.Config().GetUint(config.Key(path, "port")),
		g.Config().GetString(config.Key(path, "user")),
		g.Config().GetString(config.Key(path, "dbname")))

	g.LogConfigParameters(g.Logger(), "psql")

	dsn = fmt.Sprintf("%v password=%v", dsn, g.Config().GetString(config.Key(path, "password")))
	var err error

	g.DB, err = ConnectDB(dsn)
	if err != nil {
		return g.Logger().Fatal("failed to connect to database", err)
	}
	return nil
}

func (g *GormDB) FindByField(field string, value string, obj interface{}) (bool, error) {
	notFound, err := FindByField(g.db_(), field, value, obj)
	if err != nil && !notFound {
		e := fmt.Errorf("failed to FindByField %v", ObjectTypeName(obj))
		g.Logger().Error("GormDB", e, logger.Fields{"field": field, "value": value, "error": err})
	}
	return notFound, err
}

func (g *GormDB) FindByFields(fields map[string]interface{}, obj interface{}) (bool, error) {
	notFound, err := FindByFields(g.db_(), fields, obj)
	if err != nil && !notFound {
		e := fmt.Errorf("failed to FindByFields %v", ObjectTypeName(obj))
		g.Logger().Error("GormDB", e, logger.Fields{"fields": fields, "error": err})
	}
	return notFound, err
}

func (g *GormDB) RowsByFields(fields map[string]interface{}, obj interface{}) (db.Cursor, error) {

	var err error
	cursor := &GormCursor{GormDB: g}
	rows, err := RowsByFields(g.db_(), fields, obj)
	if err != nil {
		e := fmt.Errorf("failed to RowsByFields %v", ObjectTypeName(obj))
		g.Logger().Error("GormDB", e, logger.Fields{"fields": fields, "error": err})
	}
	cursor.Rows = rows
	cursor.Sql = rows

	return cursor, err
}

func (g *GormDB) AllRows(obj interface{}) (db.Cursor, error) {

	var err error
	cursor := &GormCursor{GormDB: g}
	rows, err := AllRows(g.db_(), obj)
	if err != nil {
		e := fmt.Errorf("failed to AllRows %v", ObjectTypeName(obj))
		g.Logger().Error("GormDB", e, logger.Fields{"error": err})
	}
	cursor.Rows = rows
	cursor.Sql = rows

	return cursor, err
}

func (g *GormDB) Create(obj orm.BaseInterface) error {
	err := Create(g.db_(), obj)
	if err != nil {
		e := fmt.Errorf("failed to Create %v", ObjectTypeName(obj))
		g.Logger().Error("GormDB", e, logger.Fields{"error": err})
	}
	return err
}

func (g *GormDB) DeleteByField(field string, value interface{}, obj interface{}) error {
	err := RemoveByField(g.db_(), field, value, obj)
	if err != nil {
		e := fmt.Errorf("failed to DeleteByField %v", ObjectTypeName(obj))
		g.Logger().Error("GormDB", e, logger.Fields{"field": field, "value": value, "error": err})
	}
	return err
}

func (g *GormDB) DeleteByFields(fields map[string]interface{}, obj interface{}) error {
	err := DeleteAllByFields(g.db_(), fields, obj)
	if err != nil {
		e := fmt.Errorf("failed to DeleteByFields %v", ObjectTypeName(obj))
		g.Logger().Error("GormDB", e, logger.Fields{"fields": fields, "error": err})
	}
	return err
}

func (g *GormDB) UpdateFields(obj interface{}, fields map[string]interface{}) error {
	err := UpdateFields(g.db_(), fields, obj)
	if err != nil {
		e := fmt.Errorf("failed to UpdateFields %v", ObjectTypeName(obj))
		g.Logger().Error("GormDB", e, logger.Fields{"error": err})
	}
	return err
}

func (g *GormDB) Transaction(handler db.TransactionHandler) error {

	nativeHandler := func(nativeTx *gorm.DB) error {
		tx := &GormDB{}
		tx.WithConfigBase = g.WithConfigBase
		tx.WithLoggerBase = g.WithLoggerBase
		tx.DB = nativeTx
		return handler(tx)
	}

	return g.DB.Transaction(nativeHandler)
}
