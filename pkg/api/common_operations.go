package api

import "github.com/evgeniums/go-backend-helpers/pkg/access_control"

func NamedResourceOperation(sampleResource Resource, resourceName string, resourceId string, op Operation) Operation {
	opResource := NewResource(resourceName)
	opResource.AddOperation(op)
	namedResource := sampleResource.CloneChain(false)
	namedResource.SetId(resourceId)
	namedResource.AddChild(opResource)
	return op
}

func Add(name string) Operation {
	return NewOperation(name, access_control.Create)
}

func Find(name string) Operation {
	return NewOperation(name, access_control.Get)
}

func Exists(name string) Operation {
	return NewOperation(name, access_control.Get)
}

func List(name string) Operation {
	return NewOperation(name, access_control.Read)
}

func Delete(name string) Operation {
	return NewOperation(name, access_control.Delete)
}

func Update(name string) Operation {
	return NewOperation(name, access_control.Put)
}

func UpdatePartial(name string) Operation {
	return NewOperation(name, access_control.Patch)
}

type UpdateCmd struct {
	Fields map[string]interface{} `json:"field"`
}

func Bind(name string) Operation {
	return NewOperation(name, access_control.Create)
}

func Unbind(name string) Operation {
	return NewOperation(name, access_control.Delete)
}
