package helpers

import (
	"bytes"
	"fmt"
	"html/template"
)

func Template(tmpl string, values map[string]interface{}) (string, error) {
	message, err := template.New("").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("creating template failed: %s", err.Error())
	}
	buf := new(bytes.Buffer)
	err = message.Execute(buf, values)
	if err != nil {
		return "", fmt.Errorf("filling template failed: %s", err.Error())
	}
	return buf.String(), nil
}
