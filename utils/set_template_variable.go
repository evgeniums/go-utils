package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
)

func SetTemplateVariable(text string, name string, value interface{}) (string, error) {
	tmplDefinition := fmt.Sprintf("{{define \"%v\"}}{{.}}{{end}}", name)
	tmpl, err := template.New("").Option("missingkey=zero").Parse(text)
	if err != nil {
		return "", err
	}
	var resultBuf bytes.Buffer
	err = tmpl.ExecuteTemplate(&resultBuf, name, value)
	if err != nil {
		return "", err
	}

	result := strings.ReplaceAll(text, tmplDefinition, resultBuf.String())
	return result, nil
}
