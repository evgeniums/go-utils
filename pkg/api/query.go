package api

import "github.com/evgeniums/go-backend-helpers/pkg/db"

type Query interface {
	Query() string
	SetQuery(q string)
}

type WithDbQuery struct {
	Query string `json:"query"`
}

type DbQuery struct {
	WithDbQuery
}

func (d *DbQuery) Query() string {
	return d.WithDbQuery.Query
}

func (d *DbQuery) SetQuery(q string) {
	d.WithDbQuery.Query = q
}

func NewDbQuery(filter *db.Filter) *DbQuery {
	cmd := &DbQuery{}
	if filter != nil {
		cmd.SetQuery(filter.ToQueryString())
	}
	return cmd
}
