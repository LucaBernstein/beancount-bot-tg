package admin

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/LucaBernstein/beancount-bot-tg/v2/db"
	"github.com/gin-gonic/gin"
)

type Log struct {
	Created string `json:"createdOn"`
	Chat    string `json:"chat"`
	Level   int    `json:"level"`
	Message string `json:"message"`
}

func GetLogs(from, to string, minLevel int) ([]Log, error) {
	rows, err := db.Connection().Query(`
		SELECT "created", "chat", "level", "message"
		FROM "app::log"
		WHERE "created" >= $1 AND "created" < $2 AND "level" >= $3
		ORDER BY "created" DESC
	`, from, to, minLevel)
	if err != nil {
		return nil, err
	}
	logs := []Log{}
	for rows.Next() {
		logEntry := &Log{}
		chatDb := sql.NullString{}
		err := rows.Scan(&logEntry.Created, &chatDb, &logEntry.Level, &logEntry.Message)
		if err != nil {
			return nil, err
		}
		logEntry.Chat = chatDb.String
		logs = append(logs, *logEntry)
	}
	return logs, nil
}

func (r *Router) Logs(c *gin.Context) {
	from := c.Query("from")
	if from == "" {
		from = time.Now().Add(-24 * time.Hour).Format(time.DateTime)
	}
	to := c.Query("to")
	if to == "" {
		to = time.Now().Format(time.DateTime)
	}
	minLevelQ := c.Query("minLevel")
	if minLevelQ == "" {
		minLevelQ = "0"
	}
	minLevel, err := strconv.Atoi(minLevelQ)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	logs, err := GetLogs(from, to, minLevel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, logs)
}
