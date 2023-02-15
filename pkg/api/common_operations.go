package api

import "github.com/evgeniums/go-backend-helpers/pkg/access_control"

func Add(name string) Operation {
	return NewOperation("add", access_control.Create)
}

func Find(name string) Operation {
	return NewOperation("find", access_control.Get)
}

func List(name string) Operation {
	return NewOperation("list", access_control.Read)
}

func Delete(name string) Operation {
	return NewOperation("delete", access_control.Delete)
}

func Update(name string) Operation {
	return NewOperation("update", access_control.Put)
}

func UpdatePartial(name string) Operation {
	return NewOperation("update_partial", access_control.Patch)
}

type UpdateCmd struct {
	Fields map[string]interface{} `json:"field"`
}

func Bind(name string) Operation {
	return NewOperation("bind", access_control.Create)
}

func Unbind(name string) Operation {
	return NewOperation("unbind", access_control.Delete)
}
