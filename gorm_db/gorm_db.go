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
	notFound, err := FindByField(g.DB, field, value, obj)
	if err != nil && !notFound {
		err = fmt.Errorf("failed to FindByField %v", ObjectTypeName(obj))
		g.Logger().Error("GormDB", err, logger.Fields{"field": field, "value": value, "error": err})
	}
	return notFound, err
}

func (g *GormDB) FindByFields(fields map[string]interface{}, obj interface{}) (bool, error) {
	notFound, err := FindByFields(g.DB, fields, obj)
	if err != nil && !notFound {
		err = fmt.Errorf("failed to FindByFields %v", ObjectTypeName(obj))
		g.Logger().Error("GormDB", err, logger.Fields{"fields": fields, "error": err})
	}
	return notFound, err
}

func (g *GormDB) RowsByFields(fields map[string]interface{}, obj interface{}) (db.Cursor, error) {

	var err error
	cursor := &GormCursor{GormDB: g}
	rows, err := RowsByFields(g.DB, fields, obj)
	if err != nil {
		err = fmt.Errorf("failed to RowsByFields %v", ObjectTypeName(obj))
		g.Logger().Error("GormDB", err)
	}
	cursor.Rows = rows
	cursor.Sql = rows

	return cursor, err
}

func (g *GormDB) AllRows(obj interface{}) (db.Cursor, error) {

	var err error
	cursor := &GormCursor{GormDB: g}
	rows, err := AllRows(g.DB, obj)
	if err != nil {
		err = fmt.Errorf("failed to AllRows %v", ObjectTypeName(obj))
		g.Logger().Error("GormDB", err)
	}
	cursor.Rows = rows
	cursor.Sql = rows

	return cursor, err
}

func (g *GormDB) Create(obj orm.BaseInterface) error {
	err := Create(g.DB, obj)
	if err != nil {
		err = fmt.Errorf("failed to Create %v", ObjectTypeName(obj))
		g.Logger().Error("GormDB", err)
	}
	return err
}

func (g *GormDB) DeleteByField(field string, value interface{}, obj interface{}) error {
	err := RemoveByField(g.DB, field, value, obj)
	if err != nil {
		err = fmt.Errorf("failed to DeleteByField %v", ObjectTypeName(obj))
		g.Logger().Error("GormDB", err, logger.Fields{"field": field, "value": value})
	}
	return err
}

func (g *GormDB) DeleteByFields(fields map[string]interface{}, obj interface{}) error {
	err := DeleteAllByFields(g.DB, fields, obj)
	if err != nil {
		err = fmt.Errorf("failed to DeleteByFields %v", ObjectTypeName(obj))
		g.Logger().Error("GormDB", err, fields)
	}
	return err
}

func (g *GormDB) UpdateFields(obj interface{}, fields map[string]interface{}) error {
	err := UpdateFields(g.DB, fields, obj)
	if err != nil {
		err = fmt.Errorf("failed to UpdateFields %v", ObjectTypeName(obj))
		g.Logger().Error("GormDB", err, fields)
	}
	return err
}

func (g *GormDB) Transaction(handler db.TransactionHandler) error {

	nativeHandler := func(nativeTx *gorm.DB) error {
		tx := g
		tx.DB = nativeTx
		return handler(tx)
	}

	return g.DB.Transaction(nativeHandler)
}
