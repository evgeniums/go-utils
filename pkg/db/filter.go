package db

import (
	"encoding/json"

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
}

type QueryInterval struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type QueryBetweenFields struct {
	FromField string `json:"from_field"`
	ToField   string `json:"to_field"`
	Value     string `json:"value"`
}

type Query struct {
	Fields        map[string]string        `json:"fields"`
	FieldsIn      map[string][]string      `json:"fields_in"`
	FieldsNotIn   map[string][]string      `json:"fields_not_in"`
	Intervals     map[string]QueryInterval `json:"intervals"`
	BetweenFields []QueryBetweenFields     `json:"betwees_fields"`

	SortField     string `json:"sort_field"`
	SortDirection string `json:"sort_direction" validate:"omitempty,oneof=asc desc"`
	Offset        int    `json:"offset" validate:"gte=0"`
	Limit         int    `json:"limit" validate:"gte=0"`
}

type WithFilterParser interface {
	PrepareFilterParser(models []interface{}, name string, validator ...*FilterValidator) (FilterParser, error)
	ParseFilter(query *Query, parserName string) (*Filter, error)
	ParseFilterDirect(query *Query, models []interface{}, name string, validator ...*FilterValidator) (*Filter, error)
}

type FilterParser interface {
	Parse(query *Query) (*Filter, error)
}

type FilterValidator struct {
	Validator validator.Validator
	Rules     map[string]string
}

func ParseQuery(db DB, query string, models []interface{}, parserName string, validator ...*FilterValidator) (*Filter, error) {

	q := &Query{}
	err := json.Unmarshal([]byte(query), q)
	if err != nil {
		return nil, err
	}

	return db.ParseFilterDirect(q, models, parserName, validator...)
}

func EmptyFilterValidator(vld validator.Validator) *FilterValidator {
	return &FilterValidator{Validator: vld}
}
