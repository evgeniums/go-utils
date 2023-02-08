package api

type Query interface {
	Query() string
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
