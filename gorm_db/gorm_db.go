package gorm_db

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/config"
	"github.com/evgeniums/go-backend-helpers/db"
	"github.com/evgeniums/go-backend-helpers/logger"
	"github.com/evgeniums/go-backend-helpers/orm"
	"gorm.io/gorm"
)

type GormDBTransaction struct {
}

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
		msg := fmt.Errorf("failed to connect to database: %s", err)
		g.Logger().Fatal(msg.Error())
		return msg
	}
	return nil
}

func (g *GormDB) FindByField(field string, value string, obj interface{}, tx ...db.Transaction) (bool, error) {
	if len(tx) != 0 {
		return tx[0].FindByField(field, value, obj)
	}
	notFound, err := FindByField(g.DB, field, value, obj)
	if err != nil && !notFound {
		err = fmt.Errorf("failed to FindByField %v", ObjectTypeName(obj))
		g.Logger().Error("GormDB", err, logger.Fields{"field": field, "value": value, "error": err})
	}
	return notFound, err
}

func (g *GormDB) FindByFields(fields map[string]interface{}, obj interface{}, tx ...db.Transaction) (bool, error) {
	if len(tx) != 0 {
		return tx[0].FindByFields(fields, obj)
	}
	notFound, err := FindByFields(g.DB, fields, obj)
	if err != nil && !notFound {
		err = fmt.Errorf("failed to FindByFields %v", ObjectTypeName(obj))
		g.Logger().Error("GormDB", err, logger.Fields{"fields": fields, "error": err})
	}
	return notFound, err
}

func (g *GormDB) Create(obj orm.BaseInterface, tx ...db.Transaction) error {
	if len(tx) != 0 {
		return tx[0].Create(obj)
	}
	err := Create(g.DB, obj)
	if err != nil {
		err = fmt.Errorf("failed to Create %v", ObjectTypeName(obj))
		g.Logger().Error("GormDB", err)
	}
	return err
}
