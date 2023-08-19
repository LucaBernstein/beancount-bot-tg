package crud

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/LucaBernstein/beancount-bot-tg/v2/helpers"
	"github.com/google/uuid"
)

func (r *Repo) CreateApiVerification(userId int64) (nonce string, err error) {
	err = EnsureApiEnabled(r, userId)
	if err != nil {
		return
	}

	challengeTimeout := 15 * time.Minute
	// Get already open, but unconfirmed verifications
	createdRow := r.db.QueryRow(`SELECT "createdOn" FROM "app::apiToken" WHERE "tgChatId" = $1 AND "nonce" IS NOT NULL;`, userId)
	var createdOn time.Time
	err = createdRow.Scan(&createdOn)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error getting open api verifications from db: %e", err)
		return "", err
	}
	// If within rate limit, don't renew, notify existing and wait until timeout before retrying
	if err == nil && createdOn.Add(challengeTimeout).After(time.Now()) {
		log.Printf("There is still an open api token challenge until %v (timeout: %v minutes).", createdOn, challengeTimeout)
		return "", helpers.ErrApiTokenChallengeInProgress
	}
	// If none or after timeout, remove existing for user and recreate.
	if err == nil {
		_, err := r.db.Exec(`DELETE FROM "app::apiToken" WHERE "tgChatId" = $1 AND "nonce" IS NOT NULL`, userId)
		if err != nil {
			log.Printf("Unexpected error while pruning open and to-be-replaced api token verification challenge: %e", err)
			return "", err
		}
	}
	sessId := uuid.NewString()
	nonce = GenNonce(8)
	_, err = r.db.Exec(`INSERT INTO "app::apiToken" ("tgChatId", "nonce", "token") VALUES ($1, $2, $3)`, userId, nonce, sessId)
	if err != nil {
		return "", err
	}
	return nonce, nil
}

func (r *Repo) VerifyApiToken(userId int64, nonce string) (token string, err error) {
	err = EnsureApiEnabled(r, userId)
	if err != nil {
		return
	}

	verifiedRow := r.db.QueryRow(`UPDATE "app::apiToken" SET NONCE = NULL WHERE "tgChatId" = $1 AND "nonce" = $2 RETURNING "token";`, userId, nonce)
	err = verifiedRow.Scan(&token)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	if err == sql.ErrNoRows {
		return "", helpers.ErrApiInvalidTokenVerification
	}
	return token, nil
}

func (r *Repo) RevokeApiToken(token string) (count int64, err error) {
	res, err := r.db.Exec(`DELETE FROM "app::apiToken" WHERE "token" = $1`, token)
	if err != nil {
		return
	}
	count, err = res.RowsAffected()
	return
}

func (r *Repo) GetTokenChatId(token string) (chatId int64, err error) {
	res := r.db.QueryRow(`SELECT "tgChatId" FROM "app::apiToken" WHERE "token" = $1`, token)
	err = res.Scan(&chatId)
	return
}

func EnsureApiEnabled(r *Repo, userId int64) error {
	if _, val, err := r.GetUserSetting(helpers.USERSET_ENABLEAPI, userId); err != nil || strings.ToUpper(val) != "TRUE" {
		return helpers.ErrApiDisabled
	}
	return nil
}

func GenNonce(digits int) string {
	n := ""
	for ; digits > 0; digits-- {
		n += fmt.Sprintf("%d", rand.Intn(10))
	}
	return n
}
