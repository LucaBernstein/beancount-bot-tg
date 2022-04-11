package crud

import (
	"fmt"

	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
)

type TemplateResult struct {
	Name     string
	Template string
}

const DB_TABLE_TEMPLATES = "bot::template"

func (r *Repo) GetTemplates(m *tb.Message, name string) ([]*TemplateResult, error) {
	LogDbf(r, helpers.TRACE, m, "Getting template(s), '%s'", name)
	additionalCondition := ""
	params := []interface{}{m.Chat.ID}
	if name != "" {
		additionalCondition = `AND "name" = $2`
		params = append(params, name)
	}
	rows, err := r.db.Query(fmt.Sprintf(`
		SELECT "name", "template" FROM "%s"
		WHERE "tgChatId" = $1 %s
	`, DB_TABLE_TEMPLATES, additionalCondition), params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []*TemplateResult{}
	var templateName string
	var templateValue string
	for rows.Next() {
		err = rows.Scan(&templateName, &templateValue)
		if err != nil {
			return nil, err
		}
		results = append(results, &TemplateResult{
			Name:     templateName,
			Template: templateValue,
		})
	}
	return results, nil
}

func (r *Repo) AddTemplate(chatId int64, name, template string) error {
	_, err := r.db.Exec(fmt.Sprintf(`
		INSERT INTO "%s" ("tgChatId", "name", "template")
		VALUES ($1, $2, $3);`, DB_TABLE_TEMPLATES), chatId, name, template)
	return err
}

func (r *Repo) RmTemplate(chatId int64, name string) (bool, error) {
	res, err := r.db.Exec(fmt.Sprintf(`DELETE FROM "%s" WHERE "tgChatId" = $1 AND "name" = $2;`, DB_TABLE_TEMPLATES),
		chatId, name)
	rows, _ := res.RowsAffected()
	return rows > 0, err
}
