package db_gorm

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
	"gorm.io/gorm/schema"
)

type FilterParser struct {
	mutex  sync.Mutex
	models map[string]bool

	Manager      *FilterManager
	DefaultModel string
	Validator    *db.FilterValidator
}

func (f *FilterParser) FindModel(name string) bool {
	f.mutex.Lock()
	_, ok := f.models[name]
	f.mutex.Unlock()
	return ok
}

func (f *FilterParser) AddModel(name string) {
	f.mutex.Lock()
	f.models[name] = true
	f.mutex.Unlock()

}

func convertValue(field *schema.Field, value string) (interface{}, error) {

	switch field.DataType {
	case schema.String:
		return value, nil
	case schema.Int:
		val, err := utils.StrToInt64(value)
		if err != nil {
			return nil, err
		}
		return val, nil
	case schema.Bool:
		val, err := utils.StrToBool(value)
		if err != nil {
			return nil, err
		}
		return val, nil
	case schema.Uint:
		val, err := utils.StrToUint64(value)
		if err != nil {
			return nil, err
		}
		return val, nil
	case schema.Float:
		val, err := utils.StrToFloat(value)
		if err != nil {
			return nil, err
		}
		return val, nil
	case schema.Time:
		val, err := utils.ParseTime(value)
		if err != nil {
			return nil, err
		}
		return val, nil
	}

	return nil, errors.New("unsupported field type")
}

func (f *FilterParser) ParseValidateField(name string, value string, onlyName ...bool) (interface{}, error) {

	justName := utils.OptionalArg(false, onlyName...)

	// extract model name
	modelName := f.DefaultModel
	fieldName := name
	parts := strings.Split(name, ".")
	if len(parts) == 2 {
		modelName = parts[0]
		fieldName = parts[1]
	} else if len(parts) != 1 {
		return nil, &validator.ValidationError{Message: "invalid field name", Field: name}
	}
	fullName := utils.ConcatStrings(modelName, ".", fieldName)

	// check if model defined for this parser
	ok := f.FindModel(modelName)
	if !ok {
		return nil, &validator.ValidationError{Message: "invalid model name", Field: modelName}
	}

	// find schema
	model, ok := f.Manager.FindModel(modelName)
	if !ok {
		return nil, &validator.ValidationError{Message: "unknown model schema", Field: modelName}
	}

	// find field
	field, ok := model.FieldsByDBName[fieldName]
	if !ok {
		return nil, &validator.ValidationError{Message: "invalid field name", Field: fieldName}
	}

	// break if only field name validation required
	if justName {
		return value, nil
	}

	// convert string value to desired type
	result, err := convertValue(field, value)
	if err != nil {
		return nil, &validator.ValidationError{Message: "invalid value", Field: fieldName}
	}

	// validate result
	if f.Validator != nil {
		useExplicitRules := false
		fieldRules := ""
		if f.Validator.Rules != nil {
			fieldRules, useExplicitRules = f.Validator.Rules[fullName]
		}
		if !useExplicitRules {
			// use tag rules
			fieldRules = field.Tag.Get("validate")
		}
		if fieldRules != "" {
			err = f.Validator.Validator.ValidateValue(result, fieldRules)
			if err != nil {
				return nil, err
			}
		}
	}

	// done
	return result, nil
}

func (f *FilterParser) Parse(query *db.Query) (*db.Filter, error) {

	// setup
	var err error
	if query == nil {
		return nil, nil
	}

	if query.SortField != "" {
		_, err = f.ParseValidateField(query.SortField, "", true)
		if err != nil {
			return nil, err
		}
	}
	filter := &db.Filter{}
	filter.SortDirection = query.SortDirection
	filter.SortField = query.SortField
	filter.Limit = query.Limit
	filter.Offset = query.Offset
	filter.Count = query.Count

	// fill fields
	if len(query.Fields) > 0 {
		filter.Fields = make(map[string]interface{})
	}
	for key, value := range query.Fields {
		value, err := f.ParseValidateField(key, value)
		if err != nil {
			return nil, err
		}
		filter.Fields[key] = value
	}

	// fill fields_in
	if len(query.FieldsIn) > 0 {
		filter.FieldsIn = make(map[string][]interface{})
	}
	for key, values := range query.FieldsIn {
		arr := make([]interface{}, len(values))
		for i := 0; i < len(values); i++ {
			arr[i], err = f.ParseValidateField(key, values[i])
			if err != nil {
				return nil, err
			}
		}
		filter.FieldsIn[key] = arr
	}

	// fill fields_not_in
	if len(query.FieldsNotIn) > 0 {
		filter.FieldsNotIn = make(map[string][]interface{})
	}
	for key, values := range query.FieldsNotIn {
		arr := make([]interface{}, len(values))
		for i := 0; i < len(values); i++ {
			arr[i], err = f.ParseValidateField(key, values[i])
			if err != nil {
				return nil, err
			}
		}
		filter.FieldsNotIn[key] = arr
	}

	// fill intervals
	if len(query.Intervals) > 0 {
		filter.Intervals = make(map[string]*Interval)
	}
	for key, interval := range query.Intervals {
		from, err := f.ParseValidateField(key, interval.From)
		if err != nil {
			return nil, err
		}
		to, err := f.ParseValidateField(key, interval.From)
		if err != nil {
			return nil, err
		}
		filter.Intervals[key] = &Interval{From: from, To: to}
	}

	// fill betweens
	if len(query.BetweenFields) > 0 {
		filter.BetweenFields = make([]*db.BetweenFields, len(query.BetweenFields))
	}
	for i := 0; i < len(query.BetweenFields); i++ {
		betweenQ := query.BetweenFields[i]
		val, err := f.ParseValidateField(betweenQ.FromField, betweenQ.Value)
		if err != nil {
			return nil, err
		}
		_, err = f.ParseValidateField(betweenQ.ToField, betweenQ.Value)
		if err != nil {
			return nil, err
		}
		filter.BetweenFields[i] = &db.BetweenFields{FromField: betweenQ.FromField, ToField: betweenQ.ToField, Value: val}
	}

	// done
	return filter, nil
}

type FilterManager struct {
	mutex         sync.Mutex
	models        map[string]*schema.Schema
	FilterParsers map[string]*FilterParser

	SchemaCache *sync.Map
	SchemaNamer schema.Namer
}

func NewFilterManager() *FilterManager {
	f := &FilterManager{}
	f.Construct()
	return f
}

func (f *FilterManager) FindModel(name string) (*schema.Schema, bool) {
	f.mutex.Lock()
	m, ok := f.models[name]
	f.mutex.Unlock()
	return m, ok
}

func (f *FilterManager) Construct() {
	f.models = make(map[string]*schema.Schema)
	f.FilterParsers = make(map[string]*FilterParser)
	f.SchemaCache = &sync.Map{}
	f.SchemaNamer = &schema.NamingStrategy{}
}

func (f *FilterManager) PrepareFilterParser(models []interface{}, name string, validator ...*db.FilterValidator) (db.FilterParser, error) {

	f.mutex.Lock()

	parser := &FilterParser{}
	parser.Manager = f
	parser.models = make(map[string]bool)

	// parse schemas
	for i, model := range models {
		s, err := schema.Parse(model, f.SchemaCache, f.SchemaNamer)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %d model's schema: %s", i, err)
		}
		f.models[s.Table] = s
		if i == 0 {
			parser.DefaultModel = s.Table
		}
		parser.AddModel(s.Table)
	}

	// keep validator
	parser.Validator = utils.OptionalArg(nil, validator...)

	// save parser in cache
	f.FilterParsers[name] = parser

	// done
	f.mutex.Unlock()
	return parser, nil
}

func (f *FilterManager) ParseFilter(query *db.Query, parserName string) (*db.Filter, error) {
	f.mutex.Lock()
	parser, ok := f.FilterParsers[parserName]
	f.mutex.Unlock()
	if !ok {
		return nil, &validator.ValidationError{Message: "unknown parser"}
	}
	return parser.Parse(query)
}

func (f *FilterManager) ParseFilterDirect(query *db.Query, models []interface{}, parserName string, vld ...*db.FilterValidator) (*db.Filter, error) {

	f.mutex.Lock()
	parser, ok := f.FilterParsers[parserName]
	f.mutex.Unlock()
	if ok {
		return parser.Parse(query)
	}

	p, err := f.PrepareFilterParser(models, parserName, vld...)
	if err != nil {
		return nil, err
	}

	return p.Parse(query)
}
