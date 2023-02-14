package db_gorm

import (
	"fmt"
	"strings"

	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type JoinTable struct {
	db.JoinTableBase
	schema *schema.Schema
}

func (jt *JoinTable) Schema(f *FilterManager) (*schema.Schema, error) {
	if jt.schema == nil {
		var err error
		jt.schema, err = schema.Parse(jt.Model, f.SchemaCache, f.SchemaNamer)
		return jt.schema, err
	}
	return jt.schema, nil
}

type JoinPair struct {
	db.JoinPairBase
	left  *JoinTable
	right *JoinTable
}

type JoinQueryConstructor struct {
	db.JoinQueryBase
	pairs []*JoinPair
}

func constructTableSelect(f *FilterManager, table *JoinTable, destinationFields map[string]map[string]string) ([]string, error) {
	var err error
	fields := make([]string, 0)

	tableModel, err := table.Schema(f)
	if err != nil {
		return nil, fmt.Errorf("failed to parse table model: %s", err)
	}
	var fieldsModel *schema.Schema
	if table.FieldsModel() != nil {
		fieldsModel, err = schema.Parse(table.FieldsModel, f.SchemaCache, f.SchemaNamer)
	} else {
		fieldsModel = tableModel
	}
	if err != nil {
		return nil, fmt.Errorf("failed to parse table model: %s", err)
	}

	tableDestinationFields := destinationFields[tableModel.Table]
	for _, field := range fieldsModel.DBNames {
		substituted := false
		if tableDestinationFields != nil {
			substitution, ok := tableDestinationFields[field]
			if ok {
				substituted = true
				fields = append(fields, fmt.Sprintf("%s.%s AS %s", tableModel.Table, field, substitution))
			}
		}
		if !substituted {
			fields = append(fields, utils.ConcatStrings(tableModel.Table, ".", field))
		}
	}

	return fields, nil
}

func constructJoins(g *gorm.DB, f *FilterManager, q *JoinQueryConstructor) (*gorm.DB, error) {
	db := g

	for _, pair := range q.pairs {
		leftSchema, err := pair.left.Schema(f)
		if err != nil {
			return nil, err
		}
		rightSchema, err := pair.right.Schema(f)
		if err != nil {
			return nil, err
		}
		join := fmt.Sprintf("JOIN %s ON %s.%s=%s.%s", rightSchema.Table, leftSchema.Table, pair.LeftField(), rightSchema.Table, pair.RightField())
		db = db.Joins(join)
	}

	return db, nil
}

func PrepareJoin(f *FilterManager, g *gorm.DB, constructor *JoinQueryConstructor) (*gorm.DB, error) {

	q := constructor

	mainModel := q.pairs[0].left.Model
	db := g.Model(mainModel)

	tables := make(map[string]*JoinTable)
	for _, pair := range q.pairs {
		left, err := pair.left.Schema(f)
		if err != nil {
			return nil, fmt.Errorf("failed to parse table model: %s", err)
		}
		tables[left.Table] = pair.left
		right, err := pair.right.Schema(f)
		if err != nil {
			return nil, fmt.Errorf("failed to parse table model: %s", err)
		}
		tables[right.Table] = pair.right
	}

	destinationFields := make(map[string]map[string]string)
	fillDestinationFields := func(jsonTag string) error {

		parts := strings.Split(jsonTag, ".")
		if len(parts) == 2 {
			tableFields, ok := destinationFields[parts[0]]
			if !ok {
				tableFields = make(map[string]string)
			}
			tableFields[parts[1]] = strings.Join(parts, "_")
			destinationFields[parts[0]] = tableFields
		}

		return nil
	}
	err := utils.EachStructTag(fillDestinationFields, "json", q.Destination)
	if err != nil {
		return nil, fmt.Errorf("failed to extract tags from: %s", err)
	}

	selects := make([]string, 0)
	for tableName, table := range tables {
		tableSelects, err := constructTableSelect(f, table, destinationFields)
		if err != nil {
			return nil, fmt.Errorf("failed to make selects for %s: %s", tableName, err)
		}
		selects = append(selects, tableSelects...)
	}

	db = db.Select(selects)
	db, err = constructJoins(db, f, q)
	if err != nil {
		return nil, fmt.Errorf("failed to construct joins: %s", err)
	}

	return db, nil
}

type JoinQuery struct {
	PreparedSession *gorm.DB
}

func (j *JoinQuery) Join(ctx logger.WithLogger, filter *Filter, dest interface{}) (int64, error) {
	session := j.PreparedSession.Session(&gorm.Session{})
	return find(session, filter, dest)
}

type Joiner struct {
	db          *GormDB
	session     *gorm.DB
	constructor *JoinQueryConstructor
	pair        *JoinPair
}

func newJoiner(db *GormDB) *Joiner {
	j := &Joiner{}
	j.constructor = newJoinQueryConstuctor()
	j.session = db.db.Session(&gorm.Session{})
	j.db = db
	return j
}

func newJoinQueryConstuctor() *JoinQueryConstructor {
	q := &JoinQueryConstructor{}
	q.pairs = make([]*JoinPair, 0)
	return q
}

func (j *Joiner) Destination(destination interface{}) (db.JoinQuery, error) {
	if j.constructor == nil {
		j.constructor = newJoinQueryConstuctor()
	}
	j.constructor.SetDestination(destination)

	preparedSession, err := PrepareJoin(j.db.filterManager, j.session, j.constructor)
	if err != nil {
		return nil, err
	}

	q := &JoinQuery{PreparedSession: preparedSession}
	return q, nil
}

func (j *Joiner) Join(model interface{}, field string, fieldsModel ...interface{}) db.JoinBegin {
	if j.constructor == nil {
		j.constructor = newJoinQueryConstuctor()
	}
	j.pair = &JoinPair{}
	j.pair.left = &JoinTable{}
	j.pair.left.JoinTableData.Model = model
	j.pair.left.JoinTableData.FieldsModel = utils.OptionalArg(nil, fieldsModel...)
	j.pair.JoinPairData.LeftField = field
	return j
}

func (j *Joiner) On(model interface{}, field string, fieldsModel ...interface{}) db.JoinEnd {
	if j.constructor == nil || j.pair == nil {
		panic("can not call ON without calling Join first")
	}
	j.pair = &JoinPair{}
	j.pair.right = &JoinTable{}
	j.pair.right.JoinTableData.Model = model
	j.pair.right.JoinTableData.FieldsModel = utils.OptionalArg(nil, fieldsModel...)
	j.pair.JoinPairData.RightField = field
	j.constructor.pairs = append(j.constructor.pairs, j.pair)
	return j
}

func (g *GormDB) Joiner() db.Joiner {
	return newJoiner(g)
}

func (g *GormDB) Join(ctx logger.WithLogger, joinConfig *db.JoinQueryConfig, filter *Filter, dest interface{}) (int64, error) {
	q := g.joinQueries.FindOrCreate(joinConfig)
	return q.Join(ctx, filter, dest)
}
