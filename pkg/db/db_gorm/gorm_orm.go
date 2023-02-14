package db_gorm

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectDB(dialector gorm.Dialector) (*gorm.DB, error) {
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	return db, err
}

func FindByField(db *gorm.DB, fieldName string, fieldValue interface{}, doc interface{}, dest ...interface{}) (bool, error) {
	dst := utils.OptionalArg(doc, dest...)
	result := db.Model(doc).First(dst, fmt.Sprintf("\"%v\" = ?", fieldName), fieldValue)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, result.Error
	}

	return true, nil
}

func FindByFields(db *gorm.DB, fields db.Fields, doc interface{}, dest ...interface{}) (bool, error) {
	dst := utils.OptionalArg(doc, dest...)
	result := db.Model(doc).Where(fields).First(dst)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, result.Error
	}

	return true, nil
}

func RowsByFields(db *gorm.DB, fields db.Fields, doc interface{}) (*sql.Rows, error) {
	return db.Model(doc).Where(fields).Rows()
}

func AllRows(db *gorm.DB, doc interface{}) (*sql.Rows, error) {
	return db.Model(doc).Rows()
}

func FindAll(db *gorm.DB, docs interface{}, dest ...interface{}) error {
	dst := utils.OptionalArg(docs, dest...)
	result := db.Model(docs).Find(dst)
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

	if filter.Fields != nil {
		h = db.Where(filter.Fields)
	}

	for field, values := range filter.FieldsIn {
		h = h.Where(fmt.Sprintf("\"%v\" IN ? ", field), values)
	}

	for field, values := range filter.FieldsNotIn {
		h = h.Where(fmt.Sprintf("\"%v\" NOT IN ? ", field), values)
	}

	for name, interval := range filter.Intervals {
		h = prepareInterval(h, name, interval)
	}

	for _, between := range filter.BetweenFields {
		h = h.Where(fmt.Sprintf(`? >= "%v" AND ? <= "%v"`, between.FromField, between.ToField), between.Value, between.Value)
	}

	return h
}

func SetFilter(g *gorm.DB, filter *Filter, docs interface{}, paginate ...bool) *gorm.DB {

	if filter == nil {
		return g
	}

	h := prepareFilter(g.Model(docs), filter)

	if filter.SortField != "" && (filter.SortDirection == db.SORT_ASC || filter.SortDirection == db.SORT_DESC) {
		h = h.Order(fmt.Sprintf("\"%v\" %v", filter.SortField, filter.SortDirection))
	}

	h = Paginate(h, filter, paginate...)

	return h
}

func Paginate(g *gorm.DB, filter *Filter, paginate ...bool) *gorm.DB {
	h := g
	if utils.OptionalArg(true, paginate...) {
		if filter.Offset > 0 {
			h = h.Offset(filter.Offset)
		}

		if filter.Limit > 0 {
			h = h.Limit(filter.Limit)
		}
	}
	return h
}

func find(g *gorm.DB, filter *Filter, dest interface{}) (int64, error) {

	var count int64

	h := SetFilter(g, filter, dest, !filter.Count)
	if filter.Count {
		counter := g.Session(&gorm.Session{})
		result := counter.Count(&count)
		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return 0, result.Error
		}
		h = Paginate(g, filter)
	}

	result := h.Find(dest)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return 0, result.Error
	}

	return count, nil
}

func FindWithFilter(db *gorm.DB, filter *Filter, docs interface{}, dest ...interface{}) (int64, error) {
	g := db.Model(docs)
	dst := utils.OptionalArg(docs, dest...)
	return find(g, filter, dst)
}

func RowsWithFilter(db *gorm.DB, filter *Filter, docs interface{}) (*sql.Rows, error) {

	h := db.Model(docs)
	h = SetFilter(h, filter, docs)

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

func FindAllByFields(db *gorm.DB, fields db.Fields, docs interface{}) error {
	result := db.Where(fields).Find(docs)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func FindNotIn(db *gorm.DB, fields db.Fields, docs interface{}) error {
	result := db.Not(fields).Find(docs)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func FindSelectNotIn(db *gorm.DB, fields db.Fields, docModel interface{}, docs interface{}) error {
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

func UpdateFields(db *gorm.DB, fields db.Fields, doc interface{}) error {
	result := db.Model(doc).Updates(fields)
	return result.Error
}

func UpdateField(db *gorm.DB, field string, doc interface{}) error {
	result := db.Model(doc).Select(field).Updates(doc)
	return result.Error
}

type TransactionHandler func(tx *gorm.DB) error

func Transaction(db *gorm.DB, handler TransactionHandler) error {
	return db.Transaction(handler)
}

var ObjectTypeName = utils.ObjectTypeName

func UpdateFieldMulti(db *gorm.DB, fields db.Fields, doc interface{}, field string, value interface{}) error {
	result := db.Model(doc).Where(fields).Update(field, value)
	return result.Error
}

func UpdateFielsdMulti(db *gorm.DB, filter db.Fields, doc interface{}, newFields db.Fields) error {
	result := db.Model(doc).Where(filter).Updates(newFields)
	return result.Error
}

func UpdateFieldsAll(db *gorm.DB, doc interface{}, newFields db.Fields) error {
	result := db.Model(doc).Where("1 = 1").Updates(newFields)
	return result.Error
}

func Exists(g *gorm.DB, filter *Filter, obj interface{}) (bool, error) {

	h := g
	if filter != nil {
		h = prepareFilter(g.Model(obj), filter)
	}
	h.Limit(1)

	result := h.First(obj)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return false, result.Error
	}

	return result.RowsAffected > 0, nil
}

func DeleteAllByFields(db *gorm.DB, fields db.Fields, docs interface{}) error {
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
