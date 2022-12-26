package validator_playground

import (
	"reflect"
	"strings"

	"github.com/evgeniums/go-backend-helpers/pkg/validator"
	playground "github.com/go-playground/validator"
)

type PlaygroundValdator struct {
	validator *playground.Validate
}

func New() *PlaygroundValdator {
	return &PlaygroundValdator{validator: playground.New()}
}

func (v *PlaygroundValdator) Validate(s interface{}) error {
	err := v.validator.Struct(s)
	if err != nil {
		field, msg, err := v.doValidation(s)
		return &validator.ValidationError{Field: field, Message: msg, Err: err}
	}

	return nil
}

func (v *PlaygroundValdator) validationSubfield(structField reflect.StructField, typenames []string) (reflect.StructField, bool) {

	first := ""
	if len(typenames) > 0 {
		first = typenames[0]
	}

	t := structField.Type
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	field, ok := t.FieldByName(first)
	if !ok {
		return reflect.StructField{}, false
	}

	if len(typenames) == 1 {
		return field, true
	}

	return v.validationSubfield(field, typenames[1:])
}

func (v *PlaygroundValdator) doValidation(s interface{}) (string, string, error) {
	err := v.validator.Struct(s)
	if err != nil {
		var name, message string
		errs := err.(playground.ValidationErrors)
		if len(errs) > 0 {
			fieldErr := errs[0]
			t := reflect.TypeOf(s)
			if reflect.ValueOf(s).Kind() == reflect.Ptr {
				t = t.Elem()
			}

			names := strings.Split(fieldErr.StructNamespace(), ".")
			f1, found := t.FieldByName(names[1])
			if !found {
				return fieldErr.Field(), "", err
			}
			var f reflect.StructField
			if len(names) > 2 {
				f, found = v.validationSubfield(f1, names[2:])
			} else {
				f = f1
			}
			if !found {
				return fieldErr.Field(), "", err
			}

			name, _ = f.Tag.Lookup("json")
			if name == "" {
				name, _ = f.Tag.Lookup("config")
			}
			message, _ = f.Tag.Lookup("vmessage")
		}
		return name, message, err
	}
	return "", "", nil
}
