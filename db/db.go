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
	FindByField(field string, value string, obj interface{}) (bool, error)
	FindByFields(fields map[string]interface{}, obj interface{}) (bool, error)
	Create(obj orm.BaseInterface) error
	DeleteByField(field string, value interface{}, obj interface{}) error
	DeleteByFields(fields map[string]interface{}, obj interface{}) error

	RowsWithFilter(filter *Filter, obj interface{}) (Cursor, error)
	AllRows(obj interface{}) (Cursor, error)

	UpdateFields(obj interface{}, fields map[string]interface{}) error

	Transaction(handler TransactionHandler) error

	EnableDebug(bool)
	EnableVerboseErrors(bool)

	FinWithFilter(filter *Filter, obj interface{}) (bool, error)
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

type Interval struct {
	From interface{}
	To   interface{}
}

type BetweenFields struct {
	FromField string
	ToField   string
	Value     interface{}
}

func (i *Interval) IsNull() bool {
	return i.From == nil && i.To == nil
}

type Filter struct {
	PreconditionFields      map[string]interface{}
	IntervalFields          map[string]*Interval
	PreconditionFieldsIn    map[string][]interface{}
	PreconditionFieldsNotIn map[string][]interface{}

	SortField     string
	SortDirection string
	Offset        int
	Limit         int
	In            []string
	Between       []*BetweenFields
}
