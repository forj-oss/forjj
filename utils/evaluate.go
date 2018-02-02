package utils

import (
	"bytes"
	"strings"
	"text/template"
)

func Evaluate(value string, tmpl *template.Template, data interface{}, funcs template.FuncMap) (_ string, _ error){
	var doc bytes.Buffer

	if ! strings.Contains(value, "{{") {
		return value, nil
	}
	value = strings.Replace(value, "\\\n", "", -1)
	if _, err := tmpl.Funcs(funcs).Parse(value) ; err != nil {
		return "", err
	}
	if err := tmpl.Execute(&doc, data) ; err != nil {
		return "", err
	}
	return doc.String(), nil
}
