package crud_test

import (
	"log"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/LucaBernstein/beancount-bot-tg/db/crud"
	"github.com/LucaBernstein/beancount-bot-tg/helpers"
)

func TestHealthGetLogs(t *testing.T) {
	// create test dependencies
	crud.TEST_MODE = true
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := crud.NewRepo(db)

	// TODO: Don't get query in here to parse (regex issues)
	mock.ExpectQuery("").
		WithArgs(helpers.ERROR, helpers.WARN, 24).
		WillReturnRows(sqlmock.NewRows([]string{"level", "c"}).
			AddRow(helpers.ERROR, 17))

	errors, warnings, err := r.HealthGetLogs(24)
	if err != nil {
		t.Errorf("Should not fail for getting health logs count: %s", err.Error())
	}
	if errors != 17 || warnings != 0 {
		t.Errorf("Unexpected logs count: errors: %d != %d or warnings: %d != %d", errors, 17, warnings, 0)
	}

	// TODO: Don't get query in here to parse (regex issues)
	mock.ExpectQuery("").
		WithArgs(helpers.ERROR, helpers.WARN, 2).
		WillReturnRows(sqlmock.NewRows([]string{"level", "c"}).
			AddRow(helpers.ERROR, 17).
			AddRow(helpers.WARN, 35))

	errors, warnings, err = r.HealthGetLogs(2)
	if err != nil {
		t.Errorf("Should not fail for getting health logs count: %s", err.Error())
	}
	if errors != 17 || warnings != 35 {
		t.Errorf("Unexpected logs count: errors: %d != %d or warnings: %d != %d", errors, 17, warnings, 35)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
