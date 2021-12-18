package bot

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/LucaBernstein/beancount-bot-tg/db/crud"
	c "github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
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
		c.LogLocalf(INFO, nil, "Got negative value. Inverting.")
		v *= -1
	}
	c.LogLocalf(TRACE, nil, "Handled float: '%s' -> %f", m.Text, v)
	return input, nil
}

func HandleRaw(m *tb.Message) (string, error) {
	return m.Text, nil
}

func ParseDate(m string) (string, error) {
	// Handle YYYY-MM-DD
	matched, err := regexp.MatchString("\\d{4}-\\d{2}-\\d{2}", m)
	if err != nil {
		return "", err
	}
	if !matched {
		return "", fmt.Errorf("Input did not match pattern 'YYYY-MM-DD'")
	}
	return m, nil
}

type Tx interface {
	Input(*tb.Message) error
	IsDone() bool
	Debug() string
	NextHint(*crud.Repo, *tb.Message) *Hint
	EnrichHint(r *crud.Repo, m *tb.Message, i Input) *Hint
	FillTemplate(currency, tag string) (string, error)
	DataKeys() map[string]string

	addStep(command command, hint string, handler func(m *tb.Message) (string, error)) Tx
	setDate(*tb.Message) (Tx, error)
}

type SimpleTx struct {
	steps       []command
	stepDetails map[command]Input
	data        []data
	date        string
	step        int
}

func CreateSimpleTx(m *tb.Message) (Tx, error) {
	tx := (&SimpleTx{
		stepDetails: make(map[command]Input),
	}).
		addStep("amount", "Please enter the amount of money (e.g. '12.34' or '12.34 USD')", HandleFloat).
		addStep("from", "Please enter the account the money came from (or select one from the list)", HandleRaw).
		addStep("to", "Please enter the account the money went to (or select one from the list)", HandleRaw).
		addStep("description", "Please enter a description (or select one from the list)", HandleRaw)
	return tx.setDate(m)
}

func (tx *SimpleTx) setDate(m *tb.Message) (Tx, error) {
	command := strings.Split(m.Text, " ")
	if len(command) >= 2 {
		date, err := ParseDate(command[1])
		if err != nil {
			return nil, err
		}
		tx.date = date
	}
	return tx, nil
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
		crud.LogDbf(r, TRACE, m, "During extraction of next hint an error ocurred: step exceeds max index.")
		return nil
	}
	return tx.EnrichHint(r, m, tx.stepDetails[tx.steps[tx.step]])
}

func (tx *SimpleTx) EnrichHint(r *crud.Repo, m *tb.Message, i Input) *Hint {
	crud.LogDbf(r, TRACE, m, "Enriching hint (%s).", i.key)
	if i.key == "description" {
		return tx.hintDescription(r, m, i.hint)
	}
	if i.key == "date" {
		return tx.hintDate(i.hint)
	}
	if c.ArrayContains([]string{"from", "to"}, i.key) {
		return tx.hintAccount(r, m, i)
	}
	return i.hint
}

func (tx *SimpleTx) hintAccount(r *crud.Repo, m *tb.Message, i Input) *Hint {
	crud.LogDbf(r, TRACE, m, "Enriching hint: account (key=%s)", i.key)
	var (
		res []string = nil
		err error    = nil
	)
	if i.key == "from" {
		res, err = r.GetCacheHints(m, c.STX_ACCF)
	} else if i.key == "to" {
		res, err = r.GetCacheHints(m, c.STX_ACCT)
	}
	if err != nil {
		crud.LogDbf(r, ERROR, m, "Error occurred getting cached hint (hintAccount): %s", err.Error())
		return i.hint
	}
	i.hint.KeyboardOptions = res
	return i.hint
}

func (tx *SimpleTx) hintDescription(r *crud.Repo, m *tb.Message, h *Hint) *Hint {
	res, err := r.GetCacheHints(m, c.STX_DESC)
	if err != nil {
		crud.LogDbf(r, ERROR, m, "Error occurred getting cached hint (hintDescription): %s", err.Error())
	}
	h.KeyboardOptions = res
	return h
}

func (tx *SimpleTx) hintDate(h *Hint) *Hint {
	h.KeyboardOptions = []string{"today"}
	return h
}

func (tx *SimpleTx) DataKeys() map[string]string {
	if tx.date == "" {
		// set today as fallback/default date
		tx.date = time.Now().Format(c.BEANCOUNT_DATE_FORMAT)
	}
	return map[string]string{
		c.STX_DATE: tx.date,
		c.STX_DESC: string(tx.data[3]),
		c.STX_ACCF: string(tx.data[1]),
		c.STX_AMTF: string(tx.data[0]),
		c.STX_ACCT: string(tx.data[2]),
	}
}

func (tx *SimpleTx) IsDone() bool {
	return tx.step >= len(tx.steps)
}

func (tx *SimpleTx) FillTemplate(currency, tag string) (string, error) {
	if !tx.IsDone() {
		return "", fmt.Errorf("not all data for this tx has been gathered")
	}
	// Variables
	txRaw := tx.DataKeys()
	f, err := strconv.ParseFloat(strings.Split(string(txRaw[c.STX_AMTF]), " ")[0], 64)
	if err != nil {
		return "", err
	}
	amountF := ParseAmount(f)
	// Add spaces
	spacesNeeded := c.DOT_INDENT - (utf8.RuneCountInString(string(txRaw[c.STX_ACCF]))) // accFrom
	spacesNeeded -= CountLeadingDigits(f)                                              // float length before point
	spacesNeeded -= 2                                                                  // additional space in template + negative sign
	if spacesNeeded < 0 {
		spacesNeeded = 0
	}
	addSpacesFrom := strings.Repeat(" ", spacesNeeded) // DOT_INDENT: 47 chars from account start to dot
	// Tag
	tagS := ""
	if tag != "" {
		tagS += " #" + tag
	}
	// Template
	tpl := `%s * "%s"%s
  %s%s -%s %s
  %s
`
	amount := strings.Split(txRaw[c.STX_AMTF], " ")
	if len(amount) >= 2 {
		// amount input contains currency
		currency = amount[1]
	}
	return fmt.Sprintf(tpl, txRaw[c.STX_DATE], txRaw[c.STX_DESC], tagS, txRaw[c.STX_ACCF], addSpacesFrom, amountF, currency, txRaw[c.STX_ACCT]), nil
}

func ParseAmount(f float64) string {
	var amountF string
	if math.Abs(math.Remainder(f*100, 1.0)) >= 1e-12 {
		// float has more than 2 remainder digits (e.g. 17.234)
		amountF = strings.TrimRight(fmt.Sprintf("%f", f), "0")
	} else {
		// at max 2 digits after the dot (e.g. 17.10)
		amountF = fmt.Sprintf("%.2f", f)
	}
	return amountF
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
