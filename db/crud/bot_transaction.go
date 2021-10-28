package crud

func (r *Repo) RecordTransaction(chatId int64, tx string) error {
	_, err := r.db.Exec(`INSERT INTO "bot::transactions" ("tgChatId", "value")
		VALUES ($1, $2);`, chatId, tx)
	return err
}
