package crud

import (
	"log"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	"gopkg.in/tucnak/telebot.v2"
)

func TestGetTemplates(t *testing.T) {
	// create test dependencies
	TEST_MODE = true
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	message := &telebot.Message{Chat: &telebot.Chat{ID: 123}, Sender: &telebot.User{ID: 123}}

	r := NewRepo(db)

	mock.ExpectQuery("").
		WithArgs(123, "special%").
		WillReturnRows(sqlmock.NewRows([]string{"name", "template"}).
			AddRow("special 1", "${date} ...").
			AddRow("special 2", "${date} ..."))

	templates, err := r.GetTemplates(message, "special")
	if err != nil {
		t.Errorf("Should not fail for getting templates: %s", err.Error())
	}
	helpers.TestExpect(t, len(templates), 2, "template result length")

	mock.ExpectQuery("").
		WithArgs(123, "special%").
		WillReturnRows(sqlmock.NewRows([]string{"name", "template"}).
			AddRow("special 1", "${date} ...").
			AddRow("special", "${date} ... special").
			AddRow("special 2", "${date} ..."))

	templates, err = r.GetTemplates(message, "special")
	if err != nil {
		t.Errorf("Should not fail for getting templates: %s", err.Error())
	}
	helpers.TestExpect(t, len(templates), 1, "template result length")
	helpers.TestExpect(t, templates[0].Template, "${date} ... special", "special template name")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
