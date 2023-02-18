package db

import (
	"encoding/json"
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

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
	Fields        Fields
	FieldsIn      map[string][]interface{}
	FieldsNotIn   map[string][]interface{}
	Intervals     map[string]*Interval
	BetweenFields []*BetweenFields

	SortField     string
	SortDirection string
	Offset        int
	Limit         int

	Count bool
}

func NewFilter() *Filter {
	return &Filter{}
}

func (f *Filter) AddFields(fields Fields) {
	for key, value := range fields {
		f.AddField(key, value)
	}
}

func (f *Filter) AddField(name string, value interface{}) {
	if f.Fields == nil {
		f.Fields = Fields{}
	}
	f.Fields[name] = value
}

func (f *Filter) AddFieldIn(name string, values ...interface{}) {
	if f.FieldsIn == nil {
		f.FieldsIn = make(map[string][]interface{})
	}
	f.Fields[name] = append([]interface{}{}, values...)
}

func (f *Filter) AddFieldNotIn(name string, values ...interface{}) {
	if f.FieldsNotIn == nil {
		f.FieldsNotIn = make(map[string][]interface{})
	}
	f.FieldsNotIn[name] = append([]interface{}{}, values...)
}

func (f *Filter) AddInterval(name string, from interface{}, to interface{}) {
	if f.Intervals == nil {
		f.Intervals = make(map[string]*Interval)
	}
	f.Intervals[name] = &Interval{From: from, To: to}
}

func (f *Filter) AddBetweenField(fromField string, toField string, value interface{}) {
	if f.BetweenFields == nil {
		f.BetweenFields = make([]*BetweenFields, 0, 1)
	}
	f.BetweenFields = append(f.BetweenFields, &BetweenFields{Value: value, FromField: fromField, ToField: toField})
}

func filterValueToString(value interface{}) string {
	return fmt.Sprintf("%v", value)
}

func (f *Filter) ToQuery() *Query {
	q := &Query{}

	q.SortField = f.SortField
	q.SortDirection = f.SortDirection
	q.Limit = f.Limit
	q.Offset = f.Offset
	q.Count = f.Count

	// fill fields
	if len(f.Fields) > 0 {
		q.Fields = make(map[string]string)
	}
	for key, value := range f.Fields {
		q.Fields[key] = filterValueToString(value)
	}

	// fill fields_in
	if len(f.FieldsIn) > 0 {
		q.FieldsIn = make(map[string][]string)
	}
	for key, values := range f.FieldsIn {
		arr := make([]string, len(values))
		for i := 0; i < len(values); i++ {
			arr[i] = filterValueToString(values[i])
		}
		q.FieldsIn[key] = arr
	}

	// fill fields_not_in
	if len(f.FieldsNotIn) > 0 {
		q.FieldsNotIn = make(map[string][]string)
	}
	for key, values := range f.FieldsNotIn {
		arr := make([]string, len(values))
		for i := 0; i < len(values); i++ {
			arr[i] = filterValueToString(values[i])
		}
		q.FieldsNotIn[key] = arr
	}

	// fill intervals
	if len(f.Intervals) > 0 {
		q.Intervals = make(map[string]QueryInterval)
	}
	for key, interval := range f.Intervals {
		q.Intervals[key] = QueryInterval{From: filterValueToString(interval.From), To: filterValueToString(interval.To)}
	}

	// fill betweens
	if len(f.BetweenFields) > 0 {
		q.BetweenFields = make([]QueryBetweenFields, len(f.BetweenFields))
	}
	for i := 0; i < len(f.BetweenFields); i++ {
		betweenQ := f.BetweenFields[i]
		q.BetweenFields[i] = QueryBetweenFields{FromField: betweenQ.FromField, ToField: betweenQ.ToField, Value: filterValueToString(betweenQ.Value)}
	}

	return q
}

func (f *Filter) ToQueryString() string {
	q := f.ToQuery()
	b, _ := json.Marshal(q)
	return string(b)
}

type QueryInterval struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type QueryBetweenFields struct {
	FromField string `json:"from_field,omitempty"`
	ToField   string `json:"to_field,omitempty"`
	Value     string `json:"value"`
}

type Query struct {
	Fields        map[string]string        `json:"fields,omitempty"`
	FieldsIn      map[string][]string      `json:"fields_in,omitempty"`
	FieldsNotIn   map[string][]string      `json:"fields_not_in,omitempty"`
	Intervals     map[string]QueryInterval `json:"intervals,omitempty"`
	BetweenFields []QueryBetweenFields     `json:"betwees_fields,omitempty"`

	SortField     string `json:"sort_field,omitempty"`
	SortDirection string `json:"sort_direction,omitempty" validate:"omitempty,oneof=asc desc"`
	Offset        int    `json:"offset,omitempty" validate:"gte=0"`
	Limit         int    `json:"limit,omitempty" validate:"gte=0"`
	Count         bool   `json:"count,omitempty"`
}

type WithFilterParser interface {
	PrepareFilterParser(model interface{}, name string, validator ...*FilterValidator) (FilterParser, error)
	ParseFilter(query *Query, parserName string) (*Filter, error)
	ParseFilterDirect(query *Query, model interface{}, name string, validator ...*FilterValidator) (*Filter, error)
}

type FilterParser interface {
	Parse(query *Query) (*Filter, error)
}

type FilterValidator struct {
	Validator validator.Validator
	Rules     map[string]string
}

func ParseQuery(db DB, query string, model interface{}, parserName string, validator ...*FilterValidator) (*Filter, error) {

	if query == "" {
		return nil, nil
	}

	q := &Query{}
	err := json.Unmarshal([]byte(query), q)
	if err != nil {
		return nil, err
	}

	return db.ParseFilterDirect(q, model, parserName, validator...)
}

func EmptyFilterValidator(vld validator.Validator) *FilterValidator {
	return &FilterValidator{Validator: vld}
}
