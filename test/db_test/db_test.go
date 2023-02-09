package db_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/test_utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, app.DB().Create(app, doc1), "failed to create doc1 in database")

	docDb1 := &SampleModel1{}
	found, err := app.DB().FindByFields(app, db.Fields{"field1": "value1"}, docDb1)
	require.NoError(t, err, "failed to find doc1 in database")
	assert.Equal(t, found, true)
	assert.Equal(t, doc1, docDb1)

	docDb1NotFound := &SampleModel1{}
	found, err = app.DB().FindByFields(app, db.Fields{"field1": "value11"}, docDb1NotFound)
	require.NoError(t, err, "failed to find docDb1NotFound in database")
	assert.Equal(t, found, false)

	filter := &db.Filter{}
	filter.Fields = db.Fields{"field2": "value2"}
	filter.SortField = "field1"
	filter.SortDirection = db.SORT_ASC
	docsDb1 := make([]*SampleModel1, 0)
	require.NoError(t, app.DB().FindWithFilter(app, filter, &docsDb1), "failed to find docs with filter in database")
	require.Len(t, docsDb1, 1)
	assert.Equal(t, doc1, docsDb1[0])

	doc2 := &SampleModel1{}
	doc2.InitObject()
	doc2.Field1 = "value1"
	doc2.Field2 = "value2"
	assert.Error(t, app.DB().Create(app, doc1), "doc with field1=valu1e must be unique in database")
	docsDb2 := make([]*SampleModel1, 0)
	require.NoError(t, app.DB().FindWithFilter(app, filter, &docsDb2), "failed to find docs with filter in database")
	require.Len(t, docsDb2, 1)
	assert.Equal(t, doc1, docsDb2[0])

	doc3 := &SampleModel1{}
	doc3.InitObject()
	doc3.Field1 = "value3"
	doc3.Field2 = "value2"
	assert.NoError(t, app.DB().Create(app, doc3), "failed to create doc3 in database")

	docsDb3 := make([]*SampleModel1, 0)
	require.NoError(t, app.DB().FindWithFilter(app, filter, &docsDb3), "failed to find docs with filter in database")
	require.Len(t, docsDb3, 2)
	assert.Equal(t, doc1, docsDb3[0])
	assert.Equal(t, doc3, docsDb3[1])

	require.NoError(t, app.DB().Update(app, doc3, db.Fields{"field1": "value3"}, db.Fields{"field2": "value33"}), "failed to update doc3 in database")

	docsDb4 := make([]*SampleModel1, 0)
	require.NoError(t, app.DB().FindWithFilter(app, filter, &docsDb4), "failed to find docsDb4 with filter in database")
	require.Len(t, docsDb4, 1)
	assert.Equal(t, doc1, docsDb4[0])

	docDb33 := &SampleModel1{}
	found, err = app.DB().FindByFields(app, db.Fields{"field2": "value33"}, docDb33)
	require.NoError(t, err, "failed to find docDb33 in database")
	assert.Equal(t, found, true)
	assert.Equal(t, docDb33.Field1, doc3.Field1)
}
