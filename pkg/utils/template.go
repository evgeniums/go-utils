package utils

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

func ReadTemplate(path string, name string, vals interface{}, language ...string) (string, error) {
	var fileName string
	if len(language) != 0 {
		fileName = fmt.Sprintf("%s/%s/%s", path, language[0], name)
		if !FileExists(fileName) {
			fileName = ""
		}
	}
	if fileName == "" {
		fileName = fmt.Sprintf("%s/%s", path, name)
	}
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
