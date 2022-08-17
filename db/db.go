package db

import (
	"errors"

	"github.com/evgeniums/go-backend-helpers/orm"
)

type Transaction interface {
	FindByField(field string, value string, obj interface{}) (bool, error)
	FindByFields(fields map[string]interface{}, obj interface{}) (bool, error)
	Create(obj orm.BaseInterface) error
	DeleteByField(field string, value interface{}, obj interface{}) error
	DeleteByFields(fields map[string]interface{}, obj interface{}) error
}

type TransactionHandler = func(tx Transaction) error

type Cursor interface {
	Next() (bool, error)
	Close() error
	Scan(obj interface{}) error
}

type DB interface {
	FindByField(field string, value string, obj interface{}, tx ...Transaction) (bool, error)
	FindByFields(fields map[string]interface{}, obj interface{}, tx ...Transaction) (bool, error)
	Create(obj orm.BaseInterface, tx ...Transaction) error
	DeleteByField(field string, value interface{}, obj interface{}, tx ...Transaction) error
	DeleteByFields(fields map[string]interface{}, obj interface{}, tx ...Transaction) error

	RowsByFields(fields map[string]interface{}, obj interface{}) (Cursor, error)
	AllRows(obj interface{}) (Cursor, error)

	Transaction(handler TransactionHandler) error

	// Find(id interface{}, obj interface{}, tx ...DBTransaction) (bool, error)
	// Delete(obj orm.WithIdInterface, tx ...DBTransaction) error
	// CreateDoc(obj interface{}, tx ...DBTransaction) error
	// Update(obj interface{}, tx ...DBTransaction) error
	// UpdateField(obj interface{}, field string, tx ...DBTransaction) error
	// UpdateFields(obj interface{}, fields ...string) error
	// UpdateFieldsTx(tx DBTransaction, obj interface{}, fields ...string) error
	// FindNotIn(fields map[string]interface{}, obj interface{}, tx ...DBTransaction) error
	// FindAllByFields(fields map[string]interface{}, obj interface{}, tx ...DBTransaction) error
	// DeleteAll(obj interface{}, tx ...DBTransaction) error
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
