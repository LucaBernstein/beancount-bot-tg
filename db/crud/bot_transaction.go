package crud

import (
	"fmt"

	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (r *Repo) RecordTransaction(chatId int64, tx string) error {
	if tx == "" {
		return fmt.Errorf("a transaction inserted into the database must not be empty")
	}
	_, err := r.db.Exec(`
		INSERT INTO "bot::transaction" ("tgChatId", "value")
		VALUES ($1, $2);`, chatId, tx)
	return err
}

type TransactionResult struct {
	Id   int
	Tx   string
	Date string
}

func (r *Repo) GetTransactions(m *tb.Message, isArchived bool) ([]*TransactionResult, error) {
	LogDbf(r, helpers.TRACE, m, "Getting transactions")
	rows, err := r.db.Query(`
		SELECT "id", "value", "created" FROM "bot::transaction"
		WHERE "tgChatId" = $1 AND "archived" = $2
		ORDER BY "created" ASC
	`, m.Chat.ID, isArchived)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	allTransactions := []*TransactionResult{}
	var id int
	var transactionString string
	var created string
	for rows.Next() {
		err = rows.Scan(&id, &transactionString, &created)
		if err != nil {
			return nil, err
		}
		allTransactions = append(allTransactions, &TransactionResult{
			Id:   id,
			Tx:   transactionString,
			Date: created,
		})
	}
	return allTransactions, nil
}

func (r *Repo) ArchiveTransactions(m *tb.Message) error {
	LogDbf(r, helpers.TRACE, m, "Archiving transactions")
	_, err := r.db.Exec(`
		UPDATE "bot::transaction"
		SET "archived" = TRUE
		WHERE "tgChatId" = $1`, m.Chat.ID)
	return err
}

func (r *Repo) DeleteTransactions(m *tb.Message) error {
	LogDbf(r, helpers.TRACE, m, "Permanently deleting transactions")
	_, err := r.db.Exec(`
		DELETE FROM "bot::transaction"
		WHERE "tgChatId" = $1`, m.Chat.ID)
	return err
}

func (r *Repo) DeleteTemplates(m *tb.Message) error {
	LogDbf(r, helpers.TRACE, m, "Permanently deleting templates")
	_, err := r.db.Exec(`
		DELETE FROM "bot::template"
		WHERE "tgChatId" = $1`, m.Chat.ID)
	return err
}

func (r *Repo) DeleteTransaction(m *tb.Message, isArchived bool, elementId int) error {
	LogDbf(r, helpers.TRACE, m, "Deleting single transaction")
	_, err := r.db.Exec(`
		DELETE FROM "bot::transaction"
		WHERE "tgChatId" = $1 AND "archived" = $2 AND "id" = $3`, m.Chat.ID, isArchived, elementId)
	return err
}
