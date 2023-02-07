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
	Manager      FilterManager
	DefaultModel string
	Models       map[string]bool
	Validator    *db.FilterValidator
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
	_, ok := f.Models[modelName]
	if !ok {
		return nil, &validator.ValidationError{Message: "invalid model name", Field: modelName}
	}

	// find schema
	model, ok := f.Manager.Models[modelName]
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
	if query == nil {
		return nil, nil
	}
	_, err := f.ParseValidateField(query.SortField, "", true)
	if err != nil {
		return nil, err
	}
	filter := &db.Filter{}
	filter.SortDirection = query.SortDirection
	filter.SortField = query.SortField
	filter.Limit = query.Limit
	filter.Offset = query.Offset

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
	Models        map[string]*schema.Schema
	FilterParsers map[string]*FilterParser

	SchemaCache *sync.Map
	SchemaNamer schema.Namer
}

func (f *FilterManager) Construct() {
	f.Models = make(map[string]*schema.Schema)
	f.FilterParsers = make(map[string]*FilterParser)
	f.SchemaCache = &sync.Map{}
	f.SchemaNamer = &schema.NamingStrategy{}
}

func (f *FilterManager) PrepareFilterParser(models []interface{}, name string, validator ...*db.FilterValidator) (db.FilterParser, error) {

	parser := &FilterParser{}
	parser.Models = make(map[string]bool)

	// parse schemas
	for i, model := range models {
		s, err := schema.Parse(model, f.SchemaCache, f.SchemaNamer)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %d model's schema: %s", i, err)
		}
		f.Models[s.Name] = s
		if i == 0 {
			parser.DefaultModel = s.Name
		}
		parser.Models[s.Name] = true
	}

	// keep validator
	parser.Validator = utils.OptionalArg(nil, validator...)

	// save parser in cache
	f.FilterParsers[name] = parser

	// done
	return parser, nil
}

func (f *FilterManager) ParseFilter(query *db.Query, parserName string) (*db.Filter, error) {
	parser, ok := f.FilterParsers[parserName]
	if !ok {
		return nil, &validator.ValidationError{Message: "unknown parser"}
	}
	return parser.Parse(query)
}

func (f *FilterManager) ParseFilterDirect(query *db.Query, models []interface{}, parserName string, vld ...*db.FilterValidator) (*db.Filter, error) {

	parser, ok := f.FilterParsers[parserName]
	if ok {
		return parser.Parse(query)
	}

	p, err := f.PrepareFilterParser(models, parserName, vld...)
	if err != nil {
		return nil, err
	}

	return p.Parse(query)
}

// switch field.FieldType.Kind() {
// case reflect.String:
// 	return value, nil
// case reflect.Int, reflect.Int64:
// 	val, err := utils.StrToInt64(value)
// 	if err != nil {
// 		return "", err
// 	}
// 	return val, nil
// case reflect.Int8, reflect.Int16, reflect.Int32:
// 	val, err := utils.StrToInt32(value)
// 	if err != nil {
// 		return "", err
// 	}
// 	return val, nil
// case reflect.Bool:
// 	val, err := utils.StrToBool(value)
// 	if err != nil {
// 		return "", err
// 	}
// 	return val, nil
// case reflect.Uint, reflect.Uint64:
// 	val, err := utils.StrToUint64(value)
// 	if err != nil {
// 		return "", err
// 	}
// 	return val, nil
// case reflect.Uint8, reflect.Uint16, reflect.Uint32:
// 	val, err := utils.StrToUint32(value)
// 	if err != nil {
// 		return "", err
// 	}
// 	return val, nil
// case reflect.Float64, reflect.Float32:
// 	val, err := utils.StrToFloat(value)
// 	if err != nil {
// 		return "", err
// 	}
// 	return val, nil
// }
