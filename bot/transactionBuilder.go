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
	. "github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
)

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
	split := strings.Split(input, " ")
	var (
		value = split[0]
	)
	if len(split) >= 3 {
		return "", fmt.Errorf("Input '%s' contained too many spaces. It should only contain the value and an optional currency", input)
	}
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return "", err
	}
	if v < 0 {
		log.Print("Got negative value. Inverting.")
		v *= -1
	}
	log.Printf("Handled float: '%s' -> %f", m.Text, v)
	return input, nil
}

func HandleRaw(m *tb.Message) (string, error) {
	return m.Text, nil
}

func HandleDate(m *tb.Message) (string, error) {
	// Handle "today" string
	if strings.HasPrefix("today", strings.TrimSpace(strings.ToLower(m.Text))) {
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
	Input(*tb.Message) error
	IsDone() bool
	Debug() string
	NextHint(*crud.Repo, *tb.Message) *Hint
	EnrichHint(r *crud.Repo, m *tb.Message, i Input) *Hint
	FillTemplate(currency string) (string, error)
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
		addStep("amount", "Please enter the amount of money (e.g. '12.34' or '12.34 USD')", HandleFloat).
		addStep("from", "Please enter the account the money came from (or select one from the list)", HandleRaw).
		addStep("to", "Please enter the account the money went to (or select one from the list)", HandleRaw).
		addStep("description", "Please enter a description (or select one from the list)", HandleRaw).
		addStep("date", "Please enter the transaction data in the format YYYY-MM-DD (or type 't' / 'today')", HandleDate)
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

func (tx *SimpleTx) NextHint(r *crud.Repo, m *tb.Message) *Hint {
	if tx.step > len(tx.steps)-1 {
		log.Printf("During extraction of next hint an error ocurred: step exceeds max index.")
		return nil
	}
	return tx.EnrichHint(r, m, tx.stepDetails[tx.steps[tx.step]])
}

func (tx *SimpleTx) EnrichHint(r *crud.Repo, m *tb.Message, i Input) *Hint {
	log.Printf("Enriching hint (%s).", i.key)
	if i.key == "description" {
		return tx.hintDescription(r, m, i.hint)
	}
	if i.key == "date" {
		return tx.hintDate(i.hint)
	}
	if ArrayContains([]string{"from", "to"}, i.key) {
		return tx.hintAccount(r, m, i)
	}
	return i.hint
}

func (tx *SimpleTx) hintAccount(r *crud.Repo, m *tb.Message, i Input) *Hint {
	log.Printf("Enriching hint: account (key=%s)", i.key)
	var (
		res []string = nil
		err error    = nil
	)
	if i.key == "from" {
		res, err = r.GetCacheHints(m, STX_ACCF)
	} else if i.key == "to" {
		res, err = r.GetCacheHints(m, STX_ACCT)
	}
	if err != nil {
		log.Printf("Error occurred getting cached hint (hintAccount): %s", err.Error())
		return i.hint
	}
	i.hint.KeyboardOptions = res
	return i.hint
}

func (tx *SimpleTx) hintDescription(r *crud.Repo, m *tb.Message, h *Hint) *Hint {
	res, err := r.GetCacheHints(m, STX_DESC)
	if err != nil {
		log.Printf("Error occurred getting cached hint (hintDescription): %s", err.Error())
	}
	h.KeyboardOptions = res
	return h
}

func (tx *SimpleTx) hintDate(h *Hint) *Hint {
	h.KeyboardOptions = []string{"today"}
	return h
}

func (tx *SimpleTx) DataKeys() map[string]string {
	return map[string]string{
		STX_DATE: string(tx.data[4]),
		STX_DESC: string(tx.data[3]),
		STX_ACCF: string(tx.data[1]),
		STX_AMTF: string(tx.data[0]),
		STX_ACCT: string(tx.data[2]),
	}
}

func (tx *SimpleTx) IsDone() bool {
	return tx.step >= len(tx.steps)
}

func (tx *SimpleTx) FillTemplate(currency string) (string, error) {
	if !tx.IsDone() {
		return "", fmt.Errorf("not all data for this tx has been gathered")
	}
	// Variables
	txRaw := tx.DataKeys()
	var (
		today = time.Now().Format(BEANCOUNT_DATE_FORMAT)
	)
	f, err := strconv.ParseFloat(strings.Split(string(txRaw[STX_AMTF]), " ")[0], 64)
	if err != nil {
		return "", err
	}
	// Add spaces
	spacesNeeded := DOT_INDENT - (utf8.RuneCountInString(string(txRaw[STX_ACCF]))) // accFrom
	spacesNeeded -= CountLeadingDigits(f)                                          // float length before point
	spacesNeeded -= 2                                                              // additional space in template + negative sign
	if spacesNeeded < 0 {
		spacesNeeded = 0
	}
	addSpacesFrom := strings.Repeat(" ", spacesNeeded) // DOT_INDENT: 47 chars from account start to dot
	// Template
	tpl := `; Created by beancount-bot-tg on %s
%s * "%s"
  %s%s -%s
  %s
`
	amount := txRaw[STX_AMTF]
	if len(strings.Split(amount, " ")) == 1 {
		// no currency in input yet
		amount += " " + currency
	}
	return fmt.Sprintf(tpl, today, txRaw[STX_DATE], txRaw[STX_DESC], txRaw[STX_ACCF], addSpacesFrom, amount, txRaw[STX_ACCT]), nil
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
