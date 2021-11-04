package crud

import "log"

func (r *Repo) RecordTransaction(chatId int64, tx string) error {
	_, err := r.db.Exec(`
		INSERT INTO "bot::transaction" ("tgChatId", "value")
		VALUES ($1, $2);`, chatId, tx)
	return err
}

func (r *Repo) GetTransactions(chatId int64) (string, error) {
	log.Printf("Getting transactions for %d", chatId)
	rows, err := r.db.Query(`
		SELECT "value" FROM "bot::transaction"
		WHERE "archived" = FALSE AND "tgChatId" = $1
		ORDER BY "created" ASC
	`, chatId)
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

func (r *Repo) ArchiveTransactions(chatId int64) error {
	log.Printf("Archiving transactions for %d", chatId)
	_, err := r.db.Exec(`
		UPDATE "bot::transaction"
		SET "archived" = TRUE
		WHERE "tgChatId" = $1`, chatId)
	return err
}

func (r *Repo) DeleteTransactions(chatId int64) error {
	log.Printf("Permanently deleting transactions for %d", chatId)
	_, err := r.db.Exec(`
		DELETE FROM "bot::transaction"
		WHERE "tgChatId" = $1`, chatId)
	return err
}
