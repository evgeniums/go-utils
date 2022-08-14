package utils

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
)

func FileExists(fileName string) bool {
	_, err := os.Stat(fileName)
	return !os.IsNotExist(err)
}

func TemplateExists(path string, name string) bool {
	fileName := fmt.Sprintf("%v/%v", path, name)
	return FileExists(fileName)
}

func ReadTemplate(path string, name string, vals interface{}) (string, error) {
	fileName := fmt.Sprintf("%v/%v", path, name)
	return ReadTemplateFile(fileName, vals)
}

func ReadTemplateFile(fileName string, vals interface{}) (string, error) {
	t, err := template.ParseFiles(fileName) //.Option("missingkey=zero")
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
