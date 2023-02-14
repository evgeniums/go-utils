package db_test

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/db/db_gorm"
	"github.com/evgeniums/go-backend-helpers/pkg/test_utils"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var _, testBasePath, _, _ = runtime.Caller(0)
var testDir = filepath.Dir(testBasePath)

func TestInitDb(t *testing.T) {
	app := test_utils.InitAppContext(t, testDir, "maindb.json")
	app.Close()
}

type SampleModel1 struct {
	common.ObjectBase
	Field1 string `gorm:"uniqueIndex"`
	Field2 string `gorm:"index"`
}

type SampleModel2 struct {
	common.ObjectBase
	Field1 string `gorm:"index"`
	Field2 int    `gorm:"index"`
}

func createDb(t *testing.T, app app_context.Context) {
	test_utils.CreateDbModel(t, app, &SampleModel1{}, &SampleModel2{})
}

func TestCreateDatabase(t *testing.T) {
	app := test_utils.InitAppContext(t, testDir, "maindb.json")
	defer app.Close()
	createDb(t, app)
}

func TestMainDbOperations(t *testing.T) {
	app := test_utils.InitAppContext(t, testDir, "maindb.json")
	defer app.Close()
	createDb(t, app)

	doc1 := &SampleModel1{}
	doc1.InitObject()
	doc1.Field1 = "value1"
	doc1.Field2 = "value2"
	require.NoError(t, app.Db().Create(app, doc1), "failed to create doc1 in database")

	docDb1 := &SampleModel1{}
	found, err := app.Db().FindByFields(app, db.Fields{"field1": "value1"}, docDb1)
	require.NoError(t, err, "failed to find doc1 in database")
	assert.Equal(t, found, true)
	assert.Equal(t, doc1, docDb1)

	docDb1NotFound := &SampleModel1{}
	found, err = app.Db().FindByFields(app, db.Fields{"field1": "value11"}, docDb1NotFound)
	require.NoError(t, err, "failed to find docDb1NotFound in database")
	assert.Equal(t, found, false)

	filter := &db.Filter{}
	filter.Fields = db.Fields{"field2": "value2"}
	filter.SortField = "field1"
	filter.SortDirection = db.SORT_ASC
	docsDb1 := make([]*SampleModel1, 0)
	require.NoError(t, app.Db().FindWithFilter(app, filter, &docsDb1), "failed to find docs with filter in database")
	require.Len(t, docsDb1, 1)
	assert.Equal(t, doc1, docsDb1[0])

	doc2 := &SampleModel1{}
	doc2.InitObject()
	doc2.Field1 = "value1"
	doc2.Field2 = "value2"
	assert.Error(t, app.Db().Create(app, doc1), "doc with field1=valu1e must be unique in database")
	docsDb2 := make([]*SampleModel1, 0)
	require.NoError(t, app.Db().FindWithFilter(app, filter, &docsDb2), "failed to find docs with filter in database")
	require.Len(t, docsDb2, 1)
	assert.Equal(t, doc1, docsDb2[0])

	doc3 := &SampleModel1{}
	doc3.InitObject()
	doc3.Field1 = "value3"
	doc3.Field2 = "value2"
	assert.NoError(t, app.Db().Create(app, doc3), "failed to create doc3 in database")

	docsDb3 := make([]*SampleModel1, 0)
	require.NoError(t, app.Db().FindWithFilter(app, filter, &docsDb3), "failed to find docs with filter in database")
	require.Len(t, docsDb3, 2)
	assert.Equal(t, doc1, docsDb3[0])
	assert.Equal(t, doc3, docsDb3[1])

	require.NoError(t, app.Db().Update(app, doc3, db.Fields{"field1": "value3"}, db.Fields{"field2": "value33"}), "failed to update doc3 in database")

	docsDb4 := make([]*SampleModel1, 0)
	require.NoError(t, app.Db().FindWithFilter(app, filter, &docsDb4), "failed to find docsDb4 with filter in database")
	require.Len(t, docsDb4, 1)
	assert.Equal(t, doc1, docsDb4[0])

	docDb33 := &SampleModel1{}
	found, err = app.Db().FindByFields(app, db.Fields{"field2": "value33"}, docDb33)
	require.NoError(t, err, "failed to find docDb33 in database")
	assert.Equal(t, found, true)
	assert.Equal(t, docDb33.Field1, doc3.Field1)
}

type ParentEssentials struct {
	ParentField1 string `json:"parent_field1"`
	ParentField2 string `json:"parent_field2"`
	ParentField3 string `json:"parent_field3"`
}

type Parent struct {
	common.ObjectBase
	common.WithNameBase
	ParentEssentials
}

type ParentDest struct {
	common.IDBase
	common.WithNameBase
	ParentEssentials
}

type ChildEssentials struct {
	ChildField1 string `json:"child_field1"`
	ChildField2 string `json:"child_field2"`
	ChildField3 string `json:"child_field3"`
}

type Child struct {
	common.ObjectBase
	ChildEssentials
	common.WithNameBase
	ParentId string `json:"parent_id"`
}

type ChildWithParentEssentials struct {
	ChildEssentials
	ParentEssentials
}

type ChildWithParent struct {
	ChildWithParentEssentials
	ParentId   string `json:"parents.id" gorm:"->;column:parents_id"`
	ParentName string `json:"parents.name" gorm:"->;column:parents_name"`
	ChildId    string `json:"children.id" gorm:"->;column:children_id"`
	ChildName  string `json:"children.name" gorm:"->;column:children_name"`
}

func TestDottedJsonFields(t *testing.T) {

	obj := &ChildWithParent{}
	obj.ParentId = "p1"
	obj.ChildId = "c1"
	obj.ParentName = "parent name 1"
	obj.ChildName = "child name 1"
	obj.ParentField1 = "parent value 1"
	obj.ChildField1 = "child value 1"

	b, err := json.MarshalIndent(obj, "", "  ")
	assert.NoError(t, err)
	t.Logf("Object: \n %s", string(b))
}

type JoinTable struct {
	Model       interface{}
	FieldsModel interface{}

	schema *schema.Schema
}

func (jt *JoinTable) Schema(f *db_gorm.FilterManager) (*schema.Schema, error) {
	if jt.schema == nil {
		var err error
		jt.schema, err = schema.Parse(jt.Model, f.SchemaCache, f.SchemaNamer)
		return jt.schema, err
	}
	return jt.schema, nil
}

type JoinPair struct {
	Left       *JoinTable
	LeftField  string
	Right      *JoinTable
	RightField string
}

type JoinQuery struct {
	Pairs       []*JoinPair
	Destination interface{}
}

func constructTableSelect(f *db_gorm.FilterManager, table *JoinTable, destinationFields map[string]map[string]string) ([]string, error) {
	var err error
	fields := make([]string, 0)

	tableModel, err := table.Schema(f)
	if err != nil {
		return nil, fmt.Errorf("failed to parse table model: %s", err)
	}
	var fieldsModel *schema.Schema
	if table.FieldsModel != nil {
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

func constructJoins(g *gorm.DB, f *db_gorm.FilterManager, q *JoinQuery) (*gorm.DB, error) {
	db := g

	for _, pair := range q.Pairs {
		leftSchema, err := pair.Left.Schema(f)
		if err != nil {
			return nil, err
		}
		rightSchema, err := pair.Right.Schema(f)
		if err != nil {
			return nil, err
		}
		join := fmt.Sprintf("JOIN %s ON %s.%s=%s.%s", rightSchema.Table, leftSchema.Table, pair.LeftField, rightSchema.Table, pair.RightField)
		db = db.Joins(join)
	}

	return db, nil
}

func ConstructJoin(f *db_gorm.FilterManager, g *gorm.DB, q *JoinQuery) (*gorm.DB, error) {

	mainModel := q.Pairs[0].Left.Model
	db := g.Model(mainModel)

	tables := make(map[string]*JoinTable)
	for _, pair := range q.Pairs {
		left, err := pair.Left.Schema(f)
		if err != nil {
			return nil, fmt.Errorf("failed to parse table model: %s", err)
		}
		tables[left.Table] = pair.Left
		right, err := pair.Right.Schema(f)
		if err != nil {
			return nil, fmt.Errorf("failed to parse table model: %s", err)
		}
		tables[right.Table] = pair.Right
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

func initOneToMany(t *testing.T) (app_context.Context, *gorm.DB) {
	app := test_utils.InitAppContext(t, testDir, "maindb.json")
	test_utils.CreateDbModel(t, app, &Parent{}, &Child{})

	g, ok := app.Db().NativeHandler().(*gorm.DB)
	require.True(t, ok)

	return app, g.Debug()
}

func TestJoin(t *testing.T) {

	app := test_utils.InitAppContext(t, testDir, "maindb.json")
	test_utils.CreateDbModel(t, app, &Parent{}, &Child{})

}

func createDocs(t *testing.T, g *gorm.DB) (*Parent, *Child) {
	parent := &Parent{}
	parent.InitObject()
	parent.NAME = "parent name"
	parent.ParentField1 = "parent field1"
	parent.ParentField2 = "parent field2"
	parent.ParentField3 = "parent field3"
	res := g.Create(parent)
	require.NoError(t, res.Error)

	child := &Child{}
	child.InitObject()
	child.NAME = "child name"
	child.ChildField1 = "child field1"
	child.ChildField2 = "child field2"
	child.ChildField3 = "child field3"
	child.ParentId = parent.GetID()
	res = g.Create(child)
	require.NoError(t, res.Error)

	return parent, child
}

func TestRawOneToMany(t *testing.T) {
	_, g := initOneToMany(t)

	parent, child := createDocs(t, g)

	dst1 := make(map[string]interface{})
	res := g.Model(&Child{}).Joins("JOIN parents on children.parent=parents.id").Find(&dst1)
	require.NoError(t, res.Error)
	t.Logf("Map result: %+v", dst1)

	dst2 := make(map[string]interface{})
	res = g.Model(&Child{}).Select("children.*", "parents.*", "parents.id as parent_id").Joins("JOIN parents on children.parent=parents.id").Find(&dst2)
	require.NoError(t, res.Error)
	t.Logf("Map result: %+v", dst2)

	childWithParent := &ChildWithParent{}
	res = g.Model(&Child{}).Select("children.*", "parents.*", "parents.id as parent_id", "parents.name as parent_name",
		"children.id as child_id", "children.name as child_name").Joins("JOIN parents on children.parent=parents.id").Find(childWithParent)
	require.NoError(t, res.Error)
	b1, err := json.MarshalIndent(childWithParent, "", "  ")
	assert.NoError(t, err)
	t.Logf("Destination object 1: \n %s", string(b1))

	b2, err := json.Marshal(childWithParent.ChildWithParentEssentials)
	assert.NoError(t, err)
	assert.Equal(t, `{"child_field1":"child field1","child_field2":"child field2","child_field3":"child field3","parent_field1":"parent field1","parent_field2":"parent field2","parent_field3":"parent field3"}`,
		string(b2))
	assert.Equal(t, parent.GetID(), childWithParent.ParentId)
	assert.Equal(t, parent.Name(), childWithParent.ParentName)
	assert.Equal(t, child.GetID(), childWithParent.ChildId)
	assert.Equal(t, child.Name(), childWithParent.ChildName)
}

func TestOneToMany(t *testing.T) {
	_, g := initOneToMany(t)
	f := &db_gorm.FilterManager{}
	f.Construct()

	createDocs(t, g)

	q := &JoinQuery{}
	q.Destination = &ChildWithParent{}
	q.Pairs = make([]*JoinPair, 1)
	left := &JoinTable{Model: &Child{}}
	right := &JoinTable{Model: &Parent{}, FieldsModel: &ParentDest{}}
	q.Pairs[0] = &JoinPair{Left: left, LeftField: "parent_id", Right: right, RightField: "id"}

	db, err := ConstructJoin(f, g, q)
	require.NoError(t, err)

	childWithParent := &ChildWithParent{}

	// db = db.Session(&gorm.Session{DryRun: true})
	// stmt := db.Find(childWithParent).Statement
	// t.Logf("Query: \n, %s", stmt.SQL.String())

	res := db.Find(childWithParent)
	require.NoError(t, res.Error)
	b1, err := json.MarshalIndent(childWithParent, "", "  ")
	assert.NoError(t, err)
	t.Logf("Destination object 1: \n %s", string(b1))
}
