package crud

import (
	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (r *Repo) RecordTransaction(chatId int64, tx string) error {
	_, err := r.db.Exec(`
		INSERT INTO "bot::transaction" ("tgChatId", "value")
		VALUES ($1, $2);`, chatId, tx)
	return err
}

func (r *Repo) GetTransactions(m *tb.Message, isArchived bool) (string, error) {
	LogDbf(r, helpers.TRACE, m, "Getting transactions")
	rows, err := r.db.Query(`
		SELECT "value" FROM "bot::transaction"
		WHERE "tgChatId" = $1 AND "archived" = $2
		ORDER BY "created" ASC
	`, m.Chat.ID, isArchived)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	SEP := "\n"
	allTransactionsMessage := ""
	var transactionString string
	for rows.Next() {
		err = rows.Scan(&transactionString)
		if err != nil {
			return "", err
		}
		allTransactionsMessage += transactionString + SEP
	}
	return allTransactionsMessage, nil
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
