package bot

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/LucaBernstein/beancount-bot-tg/db/crud"
	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
)

const CUR = "EUR"
const DOT_INDENT = 47
const (
	BEANCOUNT_DATE_FORMAT = "2006-01-02"
)

type Hint struct {
	Prompt          string
	KeyboardOptions []string
}

type command string
type data string

type Input struct {
	key     string
	hint    *Hint
	handler func(m *tb.Message) (string, error)
}

func HandleFloat(m *tb.Message) (string, error) {
	input := strings.TrimSpace(m.Text)
	input = strings.ReplaceAll(input, ",", ".")
	v, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return "", err
	}
	log.Printf("Handled float: '%s' -> %f", m.Text, v)
	return input, nil
}

func HandleRaw(m *tb.Message) (string, error) {
	return m.Text, nil
}

func HandleDate(m *tb.Message) (string, error) {
	// Handle "now" string
	if helpers.ArrayContains([]string{"now", "today"}, strings.TrimSpace(strings.ToLower(m.Text))) {
		return time.Now().Format(BEANCOUNT_DATE_FORMAT), nil
	}
	// Handle YYYY-MM-DD
	matched, err := regexp.MatchString("\\d{4}-\\d{2}-\\d{2}", m.Text)
	if err != nil {
		return "", err
	}
	if !matched {
		return "", fmt.Errorf("Input did not match pattern 'YYYY-MM-DD'")
	}
	// TODO: Try to parse date and check if valid
	return m.Text, nil
}

type Tx interface {
	Input(m *tb.Message) error
	NextHint(r *crud.Repo) *Hint
	IsDone() bool
	Debug() string
	FillTemplate() (string, error)
	DataKeys() map[string]string
	GeneralCache() *crud.GeneralCacheEntry

	addStep(command command, hint string, handler func(m *tb.Message) (string, error)) Tx
}

type SimpleTx struct {
	steps       []command
	stepDetails map[command]Input
	data        []data
	step        int
}

func CreateSimpleTx() Tx {
	return (&SimpleTx{
		stepDetails: make(map[command]Input),
	}).
		addStep("amount", "Please enter the amount of money", HandleFloat).
		addStep("from", "Please enter the account the money came from (or select one from the list)", HandleRaw).
		addStep("to", "Please enter the account the money went to (or select one from the list)", HandleRaw).
		addStep("description", "Please enter a description (or select one from the list)", HandleRaw).
		addStep("date", "Please enter the transaction data in the format YYYY-MM-DD (or select one from the list, e.g. 'now' or 'today')", HandleDate)
}

func (tx *SimpleTx) addStep(command command, hint string, handler func(m *tb.Message) (string, error)) Tx {
	tx.steps = append(tx.steps, command)
	tx.stepDetails[command] = Input{key: string(command), hint: &Hint{Prompt: hint}, handler: handler}
	tx.data = make([]data, len(tx.steps))
	return tx
}

func (tx *SimpleTx) Input(m *tb.Message) (err error) {
	res, err := tx.stepDetails[tx.steps[tx.step]].handler(m)
	if err != nil {
		return err
	}
	tx.data[tx.step] = (data)(res)
	tx.step++
	return
}

func (tx *SimpleTx) NextHint(r *crud.Repo) *Hint {
	if tx.step > len(tx.steps)-1 {
		log.Printf("During extraction of next hint an error ocurred: step exceeds max index.")
		return nil
	}
	return EnrichHint(r, tx.stepDetails[tx.steps[tx.step]])
}

func (tx *SimpleTx) IsDone() bool {
	return tx.step >= len(tx.steps)
}

func (tx *SimpleTx) FillTemplate() (string, error) {
	if !tx.IsDone() {
		return "", fmt.Errorf("not all data for this tx has been gathered")
	}
	// Variables
	txRaw := tx.DataKeys()
	var (
		today = time.Now().Format(BEANCOUNT_DATE_FORMAT)
	)
	f, err := strconv.ParseFloat(string(txRaw["amountFrom"]), 64)
	if err != nil {
		return "", err
	}
	// Add spaces
	spacesNeeded := DOT_INDENT - (utf8.RuneCountInString(string(txRaw["accFrom"]))) // accFrom
	spacesNeeded -= CountLeadingDigits(f)                                           // float length before point
	spacesNeeded -= 2                                                               // additional space in template + negative sign
	if spacesNeeded < 0 {
		spacesNeeded = 0
	}
	addSpacesFrom := strings.Repeat(" ", spacesNeeded) // DOT_INDENT: 47 chars from account start to dot
	// Template
	tpl := `; Created by beancount-bot-tg on %s
%s * "%s"
  %s%s -%s %s
  %s
`
	return fmt.Sprintf(tpl, today, txRaw["txDate"], txRaw["txDesc"], txRaw["accFrom"], addSpacesFrom, txRaw["amountFrom"], CUR, txRaw["accTo"]), nil
}

func (tx *SimpleTx) DataKeys() map[string]string {
	return map[string]string{
		"txDate":     string(tx.data[4]),
		"txDesc":     string(tx.data[3]),
		"accFrom":    string(tx.data[1]),
		"amountFrom": string(tx.data[0]),
		"accTo":      string(tx.data[2]),
	}
}

func (tx *SimpleTx) GeneralCache() *crud.GeneralCacheEntry {
	// TODO: Not implemented yet
	return nil
}

func (tx *SimpleTx) Debug() string {
	return fmt.Sprintf("SimpleTx{step=%d, totalSteps=%d, data=%v}", tx.step, len(tx.steps), tx.data)
}

func CountLeadingDigits(f float64) int {
	count := 1
	for f >= 10 {
		f /= 10
		count++
	}
	return count
}
