package dynamic_table_gorm

import (
	"fmt"
	"sync"

	"github.com/evgeniums/go-backend-helpers/pkg/api/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/db/db_gorm"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"gorm.io/gorm/schema"
)

type Table struct {
	*api_server.DynamicTable
	Schema *schema.Schema
}

type DynamicTablesGorm struct {
	mutex      sync.RWMutex
	tables     map[string]*Table
	translator api_server.DynamicFieldTranslator

	schemaCache *sync.Map
	schemaNamer schema.Namer
}

func New(translator ...api_server.DynamicFieldTranslator) *DynamicTablesGorm {

	d := &DynamicTablesGorm{}
	d.tables = make(map[string]*Table)
	d.schemaCache = &sync.Map{}
	d.schemaNamer = &schema.NamingStrategy{}

	if len(translator) != 0 {
		d.translator = translator[0]
	}

	return d
}

func (d *DynamicTablesGorm) SetTranslator(translator api_server.DynamicFieldTranslator) {
	d.translator = translator
}

func (d *DynamicTablesGorm) Table(request api_server.Request, path string) (*api_server.DynamicTable, error) {

	// setup
	c := request.TraceInMethod("DynamicTable.Table", logger.Fields{"path": path})
	request.TraceOutMethod()

	// find table
	d.mutex.RLock()
	t, ok := d.tables[path]
	d.mutex.Unlock()
	if !ok {
		return nil, c.SetErrorStr("unknown table")
	}
	c.SetLoggerField("table", t.Schema.Table)

	// clone result from stored table
	result := &api_server.DynamicTable{}
	*result = *t.DynamicTable

	// process fields
	for _, field := range result.Fields {

		// translate field's display
		if d.translator != nil {
			d.translator.Tr(field, t.Schema.Table)
		}

		// fill field's enum list
		if field.EnumGetter != nil {
			enum, err := field.EnumGetter(request)
			if err != nil {
				c.Logger().Warn("failed to translate field", db.Fields{"field": field.Field})
				continue
			}
			field.Enum = enum
		}
	}

	// done
	return result, nil
}

func (d *DynamicTablesGorm) AddTable(table *api_server.DynamicTableConfig) error {

	// create table and set default fields
	t := &Table{}
	t.DynamicTable = &api_server.DynamicTable{}
	t.Fields = make([]*api_server.DynamicTableField, 0)
	t.DefaultSortDirection = table.DefaultSortDirection
	if t.DefaultSortDirection == "" {
		t.DefaultSortDirection = db.SORT_DESC
	}
	t.DefaultSortField = table.DefaultSortField
	if t.DefaultSortField == "" {
		t.DefaultSortField = "created_at"
	}

	// parse model's schema
	s, err := schema.Parse(table.Model, d.schemaCache, d.schemaNamer)
	if err != nil {
		return fmt.Errorf("invalid model: %s", err)
	}

	// process fields
	defaultOrder := make([]string, 0, len(s.Fields))
	fields := make(map[string]*api_server.DynamicTableField)
	for _, field := range s.Fields {
		tableField := &api_server.DynamicTableField{}

		// set field name
		tableField.Field = field.Tag.Get("json")
		if tableField.Field == "" {
			tableField.Field = field.DBName
		}

		// set field display
		tableField.Display = field.Tag.Get("display")
		display, ok := table.Displays[tableField.Field]
		if ok {
			tableField.Display = display
		}
		if tableField.Display == "" {
			tableField.Display = tableField.Field
		}

		// set index flag
		tableField.Index = db_gorm.IsIndexField(field)

		// set type
		tableField.Type = string(field.DataType)

		// set enum getter
		tableField.EnumGetter = table.EnumGetters[tableField.Field]

		// add to map
		fields[tableField.Field] = tableField
		defaultOrder = append(defaultOrder, tableField.Field)
	}

	// try to order fields using column order argument
	for i := 0; i < len(table.ColumnsOrder); i++ {
		fieldName := table.ColumnsOrder[i]
		field, ok := fields[fieldName]
		if !ok {
			return fmt.Errorf("unknown field %s in column order", fieldName)
		}
		t.Fields = append(t.Fields, field)
		delete(fields, fieldName)
	}

	// sort the rest fields by default order
	for i := 0; i < len(defaultOrder); i++ {
		fieldName := defaultOrder[i]
		field, ok := fields[fieldName]
		if !ok {
			continue
		}
		t.Fields = append(t.Fields, field)
		delete(fields, fieldName)
	}

	// add table to store
	t.Schema = s
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.tables[table.Operation.Resource().ServicePathPrototype()] = t

	// done
	return nil
}
