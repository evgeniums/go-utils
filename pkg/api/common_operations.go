package api

import "github.com/evgeniums/go-backend-helpers/pkg/access_control"

func Add() Operation {
	return NewOperation("add", access_control.Create)
}

func Find() Operation {
	return NewOperation("find", access_control.Get)
}

func List() Operation {
	return NewOperation("list", access_control.Read)
}

func Delete() Operation {
	return NewOperation("delete", access_control.Delete)
}

func Update() Operation {
	return NewOperation("update", access_control.Put)
}

func UpdatePartial() Operation {
	return NewOperation("update_partial", access_control.Patch)
}

type UpdateCmd struct {
	Fields map[string]interface{} `json:"field"`
}
