package gorm_db

import (
	"github.com/evgeniums/go-backend-helpers/orm"
)

type GormTransaction struct {
	Tx *GormDB
}

func (g *GormTransaction) Create(obj orm.BaseInterface) error {
	return g.Tx.Create(obj)
}

func (g *GormTransaction) FindByField(field string, value string, obj interface{}) (bool, error) {
	return g.Tx.FindByField(field, value, obj)
}

func (g *GormTransaction) FindByFields(fields map[string]interface{}, obj interface{}) (bool, error) {
	return g.Tx.FindByFields(fields, obj)
}

func (g *GormTransaction) DeleteByField(field string, value interface{}, obj interface{}) error {
	return g.Tx.DeleteByField(field, value, obj)
}

func (g *GormTransaction) DeleteByFields(fields map[string]interface{}, obj interface{}) error {
	return g.Tx.DeleteByFields(fields, obj)
}
