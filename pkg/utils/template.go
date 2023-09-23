package utils

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

func ReadTemplate(path string, name string, vals interface{}) (string, error) {
	fileName := fmt.Sprintf("%v/%v", path, name)
	return ReadTemplateFile(fileName, vals)
}

func ReadTemplateFile(fileName string, vals interface{}) (string, error) {
	t, err := template.ParseFiles(fileName)
	if err != nil {
		return "", err
	}

	var resultBuf bytes.Buffer
	err = t.Execute(&resultBuf, vals)
	result := resultBuf.String()
	if err != nil {
		return "", err
	}
	result = strings.ReplaceAll(result, "<no value>", "")
	return result, nil
}
