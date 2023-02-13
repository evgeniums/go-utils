package utils

import (
	"reflect"
)

type StructTagHandler = func(tagValue string) error

func eachStrucTag(handler StructTagHandler, tag string, objectValue reflect.Value) error {

	objectType := objectValue.Type()
	if objectValue.Kind() == reflect.Ptr {
		objectType = objectValue.Elem().Type()
	}

	for i := 0; i < objectType.NumField(); i++ {

		field := objectType.Field(i)

		var fieldValue reflect.Value
		if objectValue.Kind() == reflect.Ptr {
			fieldValue = objectValue.Elem().Field(i)
		} else {
			fieldValue = objectValue.Field(i)
		}

		if field.Type.Kind() == reflect.Struct && field.Anonymous {
			err := eachStrucTag(handler, tag, fieldValue)
			if err != nil {
				return err
			}
		} else {
			tagValue := field.Tag.Get(tag)
			err := handler(tagValue)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func EachStructTag(handler StructTagHandler, tag string, obj interface{}) error {
	objectValue := reflect.ValueOf(obj)
	return eachStrucTag(handler, tag, objectValue)
}
