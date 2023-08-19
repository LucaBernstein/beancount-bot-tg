package helpers_test

import (
	"testing"

	"github.com/LucaBernstein/beancount-bot-tg/v2/helpers"
)

func TestTemplateParsing(t *testing.T) {
	s, err := helpers.Template("Fill in this: {{.TestValue}}", map[string]interface{}{"TestValue": "Wurzelgemüse"})
	if err != nil {
		t.Errorf("Error encountered parsing template: %s", err.Error())
	}
	if s != "Fill in this: Wurzelgemüse" {
		t.Errorf("Unexpected string after templating: %s", s)
	}
}
