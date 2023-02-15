package db_gorm

import (
	"errors"
	"fmt"
	"sync"

	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
	"gorm.io/gorm/schema"
)

type FilterParser struct {
	Manager     *FilterManager
	Destination *ModelDescriptor
	Validator   *db.FilterValidator
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

func (f *FilterParser) ParseValidateField(name string, value string, onlyName ...bool) (string, interface{}, error) {

	// TODO allow only indexed fields

	// find field by json name
	field, err := f.Destination.FindJsonField(name)
	if err != nil {
		return "", nil, &validator.ValidationError{Message: "Invalid field name", Field: name}
	}

	// break if only field name validation required
	if utils.OptionalArg(false, onlyName...) {
		return "", value, nil
	}

	// convert string value to desired type
	result, err := convertValue(field.Schema, value)
	if err != nil {
		return "", nil, &validator.ValidationError{Message: "Invalid field value", Field: name}
	}

	// validate result
	if f.Validator != nil {
		useExplicitRules := false
		fieldRules := ""
		if f.Validator.Rules != nil {
			fieldRules, useExplicitRules = f.Validator.Rules[name]
		}
		if !useExplicitRules {
			// use tag rules
			fieldRules = field.Schema.Tag.Get("validate")
		}
		if fieldRules != "" {
			err = f.Validator.Validator.ValidateValue(result, fieldRules)
			if err != nil {
				return "", nil, err
			}
		}
	}

	// done
	return field.FullDbName, result, nil
}

func (f *FilterParser) Parse(query *db.Query) (*db.Filter, error) {

	// setup
	var err error
	if query == nil {
		return nil, nil
	}
	filter := &db.Filter{}
	filter.SortDirection = query.SortDirection
	filter.Limit = query.Limit
	filter.Offset = query.Offset
	filter.Count = query.Count

	// sort field
	if query.SortField != "" {
		field, _, err := f.ParseValidateField(query.SortField, "", true)
		if err != nil {
			return nil, err
		}
		filter.SortField = field
	}

	// fill fields
	if len(query.Fields) > 0 {
		filter.Fields = make(map[string]interface{})
	}
	for key, value := range query.Fields {
		field, value, err := f.ParseValidateField(key, value)
		if err != nil {
			return nil, err
		}
		filter.Fields[field] = value
	}

	// fill fields_in
	if len(query.FieldsIn) > 0 {
		filter.FieldsIn = make(map[string][]interface{})
	}
	for key, values := range query.FieldsIn {
		arr := make([]interface{}, len(values))
		field := key
		for i := 0; i < len(values); i++ {
			field, arr[i], err = f.ParseValidateField(key, values[i])
			if err != nil {
				return nil, err
			}
		}
		filter.FieldsIn[field] = arr
	}

	// fill fields_not_in
	if len(query.FieldsNotIn) > 0 {
		filter.FieldsNotIn = make(map[string][]interface{})
	}
	for key, values := range query.FieldsNotIn {
		arr := make([]interface{}, len(values))
		field := key
		for i := 0; i < len(values); i++ {
			field, arr[i], err = f.ParseValidateField(key, values[i])
			if err != nil {
				return nil, err
			}
		}
		filter.FieldsNotIn[field] = arr
	}

	// fill intervals
	if len(query.Intervals) > 0 {
		filter.Intervals = make(map[string]*Interval)
	}
	for key, interval := range query.Intervals {
		field, from, err := f.ParseValidateField(key, interval.From)
		if err != nil {
			return nil, err
		}
		_, to, err := f.ParseValidateField(key, interval.From)
		if err != nil {
			return nil, err
		}
		filter.Intervals[field] = &Interval{From: from, To: to}
	}

	// fill betweens
	if len(query.BetweenFields) > 0 {
		filter.BetweenFields = make([]*db.BetweenFields, len(query.BetweenFields))
	}
	for i := 0; i < len(query.BetweenFields); i++ {
		betweenQ := query.BetweenFields[i]
		fromField, val, err := f.ParseValidateField(betweenQ.FromField, betweenQ.Value)
		if err != nil {
			return nil, err
		}
		toField, _, err := f.ParseValidateField(betweenQ.ToField, betweenQ.Value)
		if err != nil {
			return nil, err
		}
		filter.BetweenFields[i] = &db.BetweenFields{FromField: fromField, ToField: toField, Value: val}
	}

	// done
	return filter, nil
}

type FilterManager struct {
	mutex      sync.Mutex
	modelStore *ModelStore
	parsers    map[string]*FilterParser
}

func NewFilterManager(modelStore ...*ModelStore) *FilterManager {
	f := &FilterManager{}
	f.modelStore = utils.OptionalArg(GlobalModelStore, modelStore...)
	f.parsers = make(map[string]*FilterParser)
	return f
}

func (f *FilterManager) PrepareFilterParser(model interface{}, name string, validator ...*db.FilterValidator) (db.FilterParser, error) {

	parser := &FilterParser{}
	parser.Manager = f

	// parse schema
	s, err := schema.Parse(model, f.modelStore.schemaCache, f.modelStore.schemaNamer)
	if err != nil {
		return nil, fmt.Errorf("failed to parse model schema: %s", err)
	}
	descriptor := f.modelStore.FindDescriptor(s.Table)
	if descriptor == nil {
		return nil, fmt.Errorf("model %s (table %s) not registered", s.Name, s.Table)
	}
	if !descriptor.FieldsReady() {
		err = f.modelStore.ParseModelFields(descriptor)
		if err != nil {
			return nil, fmt.Errorf("failed to parse fields for model %s (table %s): %s", s.Name, s.Table, err)
		}
	}
	parser.Destination = descriptor

	// keep validator
	parser.Validator = utils.OptionalArg(nil, validator...)

	// save parser in cache
	f.mutex.Lock()
	f.parsers[name] = parser
	f.mutex.Unlock()

	// done
	return parser, nil
}

func (f *FilterManager) ParseFilter(query *db.Query, parserName string) (*db.Filter, error) {
	f.mutex.Lock()
	parser, ok := f.parsers[parserName]
	f.mutex.Unlock()
	if !ok {
		return nil, &validator.ValidationError{Message: "unknown parser"}
	}
	return parser.Parse(query)
}

func (f *FilterManager) ParseFilterDirect(query *db.Query, model interface{}, parserName string, vld ...*db.FilterValidator) (*db.Filter, error) {

	f.mutex.Lock()
	parser, ok := f.parsers[parserName]
	f.mutex.Unlock()
	if ok {
		return parser.Parse(query)
	}

	p, err := f.PrepareFilterParser(model, parserName, vld...)
	if err != nil {
		return nil, err
	}

	return p.Parse(query)
}
