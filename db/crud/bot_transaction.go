package crud

import "log"

func (r *Repo) RecordTransaction(chatId int64, tx string) error {
	_, err := r.db.Exec(`
		INSERT INTO "bot::transaction" ("tgChatId", "value")
		VALUES ($1, $2);`, chatId, tx)
	return err
}

func (r *Repo) GetTransactions(chatId int64, isArchived bool) (string, error) {
	log.Printf("Getting transactions for %d", chatId)
	rows, err := r.db.Query(`
		SELECT "value" FROM "bot::transaction"
		WHERE "tgChatId" = $1 AND "archived" = $2
		ORDER BY "created" ASC
	`, chatId, isArchived)
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
