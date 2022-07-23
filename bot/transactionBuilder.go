package bot

import (
	"fmt"
	"math"
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
	split := strings.Split(input, " ")
	var (
		value    = split[0]
		currency = ""
	)
	if len(split) > 2 {
		return "", fmt.Errorf("input '%s' contained too many spaces. It should only contain the value and an optional currency", input)
	}
	if len(split) == 2 {
		currency = " " + split[1]
	}
	// Should fail if tx is left open (with trailing '+' operator) and currency is given
	if strings.HasSuffix(value, "+") && currency != "" {
		return "", fmt.Errorf("for transactions being kept open with trailing '+' operator, no additionally specified currency is allowed")
	}
	operator := ""
	amounts := []string{value}
	if strings.Contains(value, "+") {
		amounts = strings.Split(value, "+")
		operator = "+"
	} else if strings.Contains(value, "*") {
		amounts = strings.Split(value, "*")
		operator = "*"
		if len(amounts) != 2 {
			return "", fmt.Errorf("expected exactly two multiplicators ('a*b')")
		}
	}
	values := []float64{}
	for _, amount := range amounts {
		value, err := handleThousandsSeparators(amount)
		if err != nil {
			return "", err
		}
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return "", fmt.Errorf("parsing failed at value '%s': %s", value, err.Error())
		}
		if v < 0 {
			c.LogLocalf(INFO, nil, "Got negative value. Inverting.")
			v *= -1
		}
		c.LogLocalf(TRACE, nil, "Handled float: '%s' -> %f", amount, v)
		values = append(values, v)
	}
	finalAmount := 0.0
	if operator == "+" {
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		finalAmount = sum
	} else if operator == "*" {
		product := 1.0
		for _, v := range values {
			product *= v
		}
		finalAmount = product
	} else {
		finalAmount = values[0]
	}
	return ParseAmount(finalAmount) + currency, nil
}

func handleThousandsSeparators(value string) (cleanValue string, err error) {
	err = fmt.Errorf("invalid separators in value '%s'", value)
	if !(strings.Contains(value, ".") && strings.Contains(value, ",")) {
		return strings.ReplaceAll(value, ",", "."), nil
	}

	thousandsSeparator := '.'
	decimalSeparator := ','
	if strings.IndexRune(value, ',') < strings.IndexRune(value, '.') {
		thousandsSeparator = ','
		decimalSeparator = '.'
	}
	if strings.Count(value, string(decimalSeparator)) > 1 {
		return
	}
	if strings.Contains(strings.SplitN(value, string(decimalSeparator), 2)[1], string(thousandsSeparator)) {
		return
	}
	const DIGITS_PER_BLOCK = 3
	for idx, block := range strings.Split(strings.Split(value, string(decimalSeparator))[0], string(thousandsSeparator)) {
		if len(block) < DIGITS_PER_BLOCK && idx != 0 {
			return
		}
	}
	newValue := value
	newValue = strings.ReplaceAll(newValue, string(thousandsSeparator), "")
	newValue = strings.ReplaceAll(newValue, ",", ".")
	return newValue, nil
}

func HandleRaw(m *tb.Message) (string, error) {
	return m.Text, nil
}

func ParseDate(m string) (string, error) {
	// TODO: Handle tz offset
	today := time.Now()
	patterns := []string{
		"2006-01-02",
		"20060102",
		"01-02",
		"0102",
		"02",
	}
	for _, p := range patterns {
		t, err := time.Parse(p, m)
		if err == nil {
			if len(m) < len("20060102") {
				t, _ = time.Parse(c.BEANCOUNT_DATE_FORMAT, fmt.Sprintf("%s-%s", today.Format("2006"), t.Format("01-02")))
			}
			if len(m) < len("0102") {
				t, _ = time.Parse(c.BEANCOUNT_DATE_FORMAT, fmt.Sprintf("%s-%s", today.Format("2006-01"), t.Format("02")))
			}
			return t.Format(c.BEANCOUNT_DATE_FORMAT), nil
		}
	}
	return "", fmt.Errorf("Input could not be parsed to a specific date. Multiple date formats are allowed, e.g. YYYY-MM-DD, MM-DD or DD")
}

type Tx interface {
	Input(*tb.Message) error
	IsDone() bool
	Debug() string
	NextHint(*crud.Repo, *tb.Message) *Hint
	EnrichHint(r *crud.Repo, m *tb.Message, i Input) *Hint
	FillTemplate(currency, tag string, tzOffset int) (string, error)
	DataKeys() map[string]string

	addStep(command command, hint string, handler func(m *tb.Message) (string, error)) Tx
	SetDate(string) (Tx, error)
	setTimeIfEmpty(tzOffset int) bool
}

type SimpleTx struct {
	template    string
	steps       []command
	stepDetails map[command]Input
	data        []data
	date_upd    string
	step        int
}

const TEMPLATE_SIMPLE_DEFAULT = `${date} * "${description}"${tag}
  ${from} ${-amount}
  ${to}`

func CreateSimpleTx(suggestedCur, template string) (Tx, error) {
	tx := (&SimpleTx{
		stepDetails: make(map[command]Input),
		template:    template,
	}).
		addStep("amount", fmt.Sprintf("Please enter the amount of money (e.g. '12.34' or '12.34 %s')", suggestedCur), HandleFloat).
		addStep("from", "Please enter the account the money came from (or select one from the list)", HandleRaw).
		addStep("to", "Please enter the account the money went to (or select one from the list)", HandleRaw).
		addStep("description", "Please enter a description (or select one from the list)", HandleRaw)
	return tx, nil
}

func (tx *SimpleTx) SetDate(d string) (Tx, error) {
	date, err := ParseDate(d)
	if err != nil {
		return nil, err
	}
	tx.date_upd = date
	return tx, nil
}

func ParseTemplateFields(template string) (fields map[string]*TemplateField) {
	fields = make(map[string]*TemplateField)
	varBegins := strings.SplitAfter(template, "${")
	for _, v := range varBegins {
		field := ParseTemplateField(strings.Split(v, "}")[0])
		fields[field.Raw] = field
	}
	return
}

type TemplateField struct {
	Name string

	Fraction int

	IsNegative bool

	Raw string
}

func ParseTemplateField(rawField string) *TemplateField {
	field := &TemplateField{
		Raw:      rawField,
		Name:     rawField,
		Fraction: 1,
	}

	field.IsNegative = strings.HasPrefix(field.Name, "-")
	field.Name = strings.TrimLeft(field.Name, "-")

	fractionSplits := strings.Split(field.Name, "/")
	if len(fractionSplits) == 2 {
		field.Name = fractionSplits[0]
		var err error
		field.Fraction, err = strconv.Atoi(fractionSplits[1])
		if err != nil {
			c.LogLocalf(ERROR, nil, "converting fraction for template failed: '%s' -> %s", rawField, err.Error())
			field.Fraction = 1
		}
	} else {
		field.Name = fractionSplits[0]
	}

	return field
}

func (tx *SimpleTx) addStep(command command, hint string, handler func(m *tb.Message) (string, error)) Tx {
	templateFields := ParseTemplateFields(tx.template)
	exists := false
	for _, f := range templateFields {
		if f.Name == string(command) {
			exists = true
		}
	}
	if !exists {
		return tx
	}
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
	varMap := make(map[string]string)
	varMap["date"] = tx.date_upd
	for i, v := range tx.steps {
		varMap[string(v)] = string(tx.data[i])
	}
	return varMap
}

func (tx *SimpleTx) IsDone() bool {
	return tx.step >= len(tx.steps)
}

func (tx *SimpleTx) setTimeIfEmpty(tzOffset int) bool {
	if tx.date_upd == "" {
		// set today as fallback/default date
		timezoneOff := time.Duration(tzOffset) * time.Hour
		tx.date_upd = time.Now().UTC().Add(timezoneOff).Format(c.BEANCOUNT_DATE_FORMAT)
		return true
	}
	return false
}

func (tx *SimpleTx) FillTemplate(currency, tag string, tzOffset int) (string, error) {
	if !tx.IsDone() {
		return "", fmt.Errorf("not all data for this tx has been gathered")
	}
	// If still empty, set time and correct for timezone
	tx.setTimeIfEmpty(tzOffset)

	varMap := tx.DataKeys()

	template := tx.template
	fields := ParseTemplateFields(tx.template)
	var amountFields []*TemplateField
	for _, f := range fields {
		// TODO: Refactor!
		value := f.Raw
		if f.Name == "amount" {
			amountFields = append(amountFields, f)
			continue // Only replace last for formatting
		} else if f.Name == "description" {
			if v, exists := varMap["description"]; exists {
				value = v
			}
		} else if f.Name == "tag" {
			tagS := ""
			if tag != "" {
				tagS += " #" + tag
			}
			value = tagS
		} else if f.Name == "date" {
			if v, exists := varMap["date"]; exists {
				value = v
			}
		} else if f.Name == "from" {
			if v, exists := varMap["from"]; exists {
				value = v
			}
		} else if f.Name == "to" {
			if v, exists := varMap["to"]; exists {
				value = v
			}
		} else {
			continue
		}
		template = strings.ReplaceAll(template, fmt.Sprintf("${%s}", f.Raw), value)
	}
	for _, amountField := range amountFields {
		if v, exists := varMap["amount"]; exists {
			amount := strings.Split(v, " ")
			if len(amount) >= 2 {
				// amount input contains currency
				currency = amount[1]
			}
			f, err := strconv.ParseFloat(amount[0], 64)
			if err != nil {
				return "", err
			}

			oldTemplate := template
			template = ""
			for _, line := range strings.Split(oldTemplate, "\n") {
				if strings.Contains(line, fmt.Sprintf("${%s}", amountField.Raw)) {
					before := strings.Split(line, fmt.Sprintf("${%s}", amountField.Raw))[0]
					spacesNeeded := c.DOT_INDENT - utf8.RuneCountInString(before)
					fractionedAmount := f / float64(amountField.Fraction)
					spacesNeeded -= CountLeadingDigits(fractionedAmount) // float length before point
					spacesNeeded += 2                                    // indentation
					negSign := ""
					if amountField.IsNegative {
						negSign = "-"
						spacesNeeded -= 1
					}
					if spacesNeeded < 0 {
						spacesNeeded = 0
					}
					addSpacesFrom := strings.Repeat(" ", spacesNeeded) // DOT_INDENT: 47 chars from account start to dot
					template += strings.ReplaceAll(line, fmt.Sprintf("${%s}", amountField.Raw), addSpacesFrom+negSign+ParseAmount(fractionedAmount)+" "+currency) + "\n"
				} else {
					template += line + "\n"
				}
			}
		}
	}
	return strings.TrimSpace(template) + "\n", nil
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
