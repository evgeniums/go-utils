package app_context

type SingleParameter struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

type GeneralSettings struct {
	Name       string            `json:"name"`
	Parameters []SingleParameter `json:"parameters"`
}
