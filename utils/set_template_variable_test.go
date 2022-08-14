package utils_test

import (
	"bytes"
	"fmt"
	template "html/template"
	"strings"
	"testing"

	"github.com/evgeniums/go-backend-helpers/utils"
)

type testStruct struct {
	Replace string
}

func TestTemplateStruct(t *testing.T) {

	text := "Hello {{.Replace}}!"
	tmpl, err := template.New("test").Parse(text)
	if err != nil {
		t.Fatalf("Parsing failed: %s", err)
	}
	var processedText bytes.Buffer
	err = tmpl.Execute(&processedText, &testStruct{"world"})
	if err != nil {
		t.Fatalf("Execution failed: %s", err)
	}

	result := processedText.String()
	sample := "Hello world!"
	if result != sample {
		t.Fatalf("Invalid result: expected %v, got %v", sample, result)
	}
}

func TestTemplateVariable(t *testing.T) {

	text := "Hello {{.}}!"
	tmpl, err := template.New("test").Parse(text)
	if err != nil {
		t.Fatalf("Parsing failed: %s", err)
	}
	var processedText bytes.Buffer
	err = tmpl.Execute(&processedText, "world")
	if err != nil {
		t.Fatalf("Execution failed: %s", err)
	}

	result := processedText.String()
	sample := "Hello world!"
	if result != sample {
		t.Fatalf("Invalid result: expected %v, got %v", sample, result)
	}
}

func TestNamedTemplateVariable(t *testing.T) {

	text := "Hello {{.}}!"
	tmpl, err := template.New("test").Parse(text)
	if err != nil {
		t.Fatalf("Parsing failed: %s", err)
	}
	var processedText bytes.Buffer
	err = tmpl.ExecuteTemplate(&processedText, "test", "world")
	if err != nil {
		t.Fatalf("Execution failed: %s", err)
	}

	result := processedText.String()
	sample := "Hello world!"
	if result != sample {
		t.Fatalf("Invalid result: expected %v, got %v", sample, result)
	}
}

func TestEmbeddedTemplate(t *testing.T) {

	tmplDefinition := `{{define "TemplateName"}}{{.}}{{end}}`
	text := fmt.Sprintf("Hello %v and something more!", tmplDefinition)
	tmpl, err := template.New("test").Parse(text)
	if err != nil {
		t.Fatalf("Parsing failed: %s", err)
	}
	var templateText bytes.Buffer
	err = tmpl.ExecuteTemplate(&templateText, "TemplateName", "world")
	if err != nil {
		t.Fatalf("Execution failed: %s", err)
	}
	internalResult := templateText.String()
	sample := "world"
	if internalResult != sample {
		t.Fatalf("Invalid internalResult: expected %v, got %v", sample, internalResult)
	}

	result := strings.ReplaceAll(text, tmplDefinition, internalResult)
	sample = "Hello world and something more!"
	if result != sample {
		t.Fatalf("Invalid result: expected %v, got %v", sample, result)
	}
}

func TestSetTemplateVariable(t *testing.T) {
	html := `<html><body><div id="div_id0" data-captcha-key="{{define "CaptchaSiteKey"}}{{.}}{{end}}"></div></body></html>`

	result, err := utils.SetTemplateVariable(html, "CaptchaSiteKey", "key-12345")
	if err != nil {
		t.Fatalf("Operation failed: %s", err)
	}
	sample := `<html><body><div id="div_id0" data-captcha-key="key-12345"></div></body></html>`
	if result != sample {
		t.Fatalf("Invalid result: expected %v, got %v", sample, result)
	}
}
