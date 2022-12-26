package db_gorm

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"

	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	return db, err
}

func FindByField(db *gorm.DB, fieldName string, fieldValue interface{}, doc interface{}) (bool, error) {
	result := db.First(doc, fmt.Sprintf("\"%v\" = ?", fieldName), fieldValue)
	if result.Error != nil {
		notFound := errors.Is(result.Error, gorm.ErrRecordNotFound)
		return notFound, result.Error
	}

	return false, nil
}

func FindByFields(db *gorm.DB, fields map[string]interface{}, doc interface{}) (bool, error) {
	result := db.Where(fields).First(doc)
	if result.Error != nil {
		notFound := errors.Is(result.Error, gorm.ErrRecordNotFound)
		return notFound, result.Error
	}

	return false, nil
}

func RowsByFields(db *gorm.DB, fields map[string]interface{}, doc interface{}) (*sql.Rows, error) {
	return db.Model(doc).Where(fields).Rows()
}

func AllRows(db *gorm.DB, doc interface{}) (*sql.Rows, error) {
	return db.Model(doc).Rows()
}

func FindAll(db *gorm.DB, docs interface{}) error {
	result := db.Find(docs)
	return result.Error

}

type Interval = db.Interval
type Filter = db.Filter

func prepareInterval(db *gorm.DB, name string, interval *Interval) *gorm.DB {
	h := db

	if interval.From != nil && interval.To != nil {
		if interval.From == interval.To {
			h = h.Where(fmt.Sprintf("\"%v\" = ?", name), interval.From)
		} else {
			h = h.Where(fmt.Sprintf("\"%v\" >= ? AND \"%v\" <= ? ", name, name), interval.From, interval.To)
		}
	} else if interval.From != nil {
		h = h.Where(fmt.Sprintf("\"%v\" >= ? ", name), interval.From)
	} else if interval.To != nil {
		h = h.Where(fmt.Sprintf("\"%v\" <= ? ", name), interval.To)
	}
	return h
}

func prepareFilter(db *gorm.DB, filter *Filter) *gorm.DB {
	h := db

	if filter.PreconditionFields != nil {
		h = db.Where(filter.PreconditionFields)
	}

	if filter.PreconditionFieldsIn != nil {
		for field, values := range filter.PreconditionFieldsIn {
			h = h.Where(fmt.Sprintf("\"%v\" IN ? ", field), values)
		}
	}

	if filter.PreconditionFieldsNotIn != nil {
		for field, values := range filter.PreconditionFieldsNotIn {
			h = h.Where(fmt.Sprintf("\"%v\" NOT IN ? ", field), values)
		}
	}

	for name, interval := range filter.IntervalFields {
		h = prepareInterval(h, name, interval)
	}

	for _, between := range filter.Between {
		h = h.Where(fmt.Sprintf(`? >= "%v" AND ? <= "%v"`, between.FromField, between.ToField), between.Value, between.Value)
	}

	return h
}

func FindWithFilter(db *gorm.DB, filter *Filter, doc interface{}) (bool, error) {

	h := prepareFilter(db, filter)

	if filter.SortField != "" && (filter.SortDirection == "asc" || filter.SortDirection == "desc") {
		h = h.Order(fmt.Sprintf("\"%v\" %v", filter.SortField, filter.SortDirection))
	}

	if filter.Offset > 0 {
		h = h.Offset(filter.Offset)
	}

	if filter.Limit > 0 {
		h = h.Limit(filter.Limit)
	}

	result := h.Find(doc)
	if result.Error != nil {
		notFound := errors.Is(result.Error, gorm.ErrRecordNotFound)
		return notFound, result.Error
	}

	return false, nil
}

func RowsWithFilter(db *gorm.DB, filter *Filter, docs interface{}) (*sql.Rows, error) {

	h := prepareFilter(db.Model(docs), filter)

	if filter.SortField != "" && (filter.SortDirection == "asc" || filter.SortDirection == "desc") {
		h = h.Order(fmt.Sprintf("\"%v\" %v", filter.SortField, filter.SortDirection))
	}

	if filter.Offset > 0 {
		h = h.Offset(filter.Offset)
	}

	if filter.Limit > 0 {
		h = h.Limit(filter.Limit)
	}

	rows, err := h.Rows()
	if err != nil {
		return nil, err
	}
	return rows, err
}

func CountWithFilter(db *gorm.DB, filter *Filter, doc interface{}) int64 {

	m := db.Model(doc)
	h := prepareFilter(m, filter)

	var count int64
	h.Count(&count)
	return count
}

func SumWithFilter(db *gorm.DB, filter *Filter, fields map[string]string, doc interface{}, result interface{}) error {

	sums := ""
	for key, name := range fields {
		if sums != "" {
			sums += ", "
		}
		sums += fmt.Sprintf("sum(%v) as %v", key, name)
	}

	m := db.Model(doc).Select(sums)
	h := prepareFilter(m, filter)

	r := h.Take(result)
	return r.Error
}

func FindAllByFields(db *gorm.DB, fields map[string]interface{}, docs interface{}) error {
	result := db.Where(fields).Find(docs)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func FindNotIn(db *gorm.DB, fields map[string]interface{}, docs interface{}) error {
	result := db.Not(fields).Find(docs)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func FindSelectNotIn(db *gorm.DB, fields map[string]interface{}, docModel interface{}, docs interface{}) error {
	result := db.Model(docModel).Not(fields).Find(docs)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func RemoveById(db *gorm.DB, id interface{}, doc interface{}) error {
	result := db.Where("id = ?", id).Delete(doc)
	return result.Error
}

func RemoveByField(db *gorm.DB, field string, value interface{}, doc interface{}) error {
	result := db.Where(fmt.Sprintf("\"%v\" = ?", field), value).Delete(doc)
	return result.Error
}

func Create(db *gorm.DB, doc interface{}) error {
	result := db.Create(doc)
	return result.Error
}

func UpdateFields(db *gorm.DB, fields map[string]interface{}, doc interface{}) error {
	result := db.Model(doc).Updates(fields)
	return result.Error
}

func UpdateField(db *gorm.DB, field string, doc interface{}) error {
	result := db.Model(doc).Select(field).Updates(doc)
	return result.Error
}

/*
func collectFieldNames(t reflect.Type, names *[]string) {

		// Return if not struct or pointer to struct.
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		if t.Kind() != reflect.Struct {
			return
		}

		// Iterate through fields collecting names in map.
		for i := 0; i < t.NumField(); i++ {
			sf := t.Field(i)

			// Recurse into anonymous fields.
			if sf.Anonymous {
				if sf.Name != "BaseObject" {
					collectFieldNames(sf.Type, names)
				}
			} else {
				*names = append(*names, sf.Name)
			}
		}
	}
*/
type TransactionHandler func(tx *gorm.DB) error

func Transaction(db *gorm.DB, handler TransactionHandler) error {
	return db.Transaction(handler)
}

func ObjectTypeName(obj interface{}) string {
	t := reflect.TypeOf(obj)
	if reflect.ValueOf(obj).Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

func UpdateFieldMulti(db *gorm.DB, fields map[string]interface{}, doc interface{}, field string, value interface{}) error {
	result := db.Model(doc).Where(fields).Update(field, value)
	return result.Error
}

func DeleteAllByFields(db *gorm.DB, fields map[string]interface{}, docs interface{}) error {
	result := db.Where(fields).Delete(docs)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func FindAllInterval(db *gorm.DB, name string, interval *Interval, docs interface{}) error {
	h := prepareInterval(db, name, interval)
	result := h.Find(docs)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func DeleteAll(db *gorm.DB, docs interface{}) error {
	result := db.Where("1 = 1").Delete(docs)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
