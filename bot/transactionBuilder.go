package bot

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/LucaBernstein/beancount-bot-tg/v2/db/crud"
	c "github.com/LucaBernstein/beancount-bot-tg/v2/helpers"
	"github.com/fatih/structs"
	tb "gopkg.in/telebot.v3"
)

type Hint struct {
	Prompt          string
	KeyboardOptions []string
}

type Input struct {
	key     string
	hint    *Hint
	handler func(m *tb.Message) (string, error)
	field   TemplateField
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
	return FORMATTER_PLACEHOLDER + ParseAmount(finalAmount) + currency, nil
}

func handleThousandsSeparators(value string) (cleanValue string, err error) {
	err = fmt.Errorf("invalid separators in value '%s'", value)
	if !(strings.Contains(value, ".") && strings.Contains(value, ",")) &&
		!(strings.Count(value, ".") > 1 || strings.Count(value, ",") > 1) {
		return strings.ReplaceAll(value, ",", "."), nil
	}

	thousandsSeparator := '.'
	decimalSeparator := ','
	if strings.ContainsRune(value, ',') && (strings.IndexRune(value, ',') < strings.IndexRune(value, '.') || !strings.ContainsRune(value, '.')) {
		thousandsSeparator = ','
		decimalSeparator = '.'
	}
	if strings.Count(value, string(decimalSeparator)) > 1 {
		return
	}
	if strings.ContainsRune(value, decimalSeparator) && strings.Contains(strings.SplitN(value, string(decimalSeparator), 2)[1], string(thousandsSeparator)) {
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
	today := time.Now().UTC()
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
	Prepare() Tx
	Input(*tb.Message) (bool, error)
	IsDone() bool
	Debug() string
	NextHint(*crud.Repo, *tb.Message) *Hint
	EnrichHint(r *crud.Repo, m *tb.Message, i *Input) *Hint
	FillTemplate(currency, tag string, tzOffset int) (string, error)
	CacheData() map[string]string

	SetDate(string) (Tx, error)
	setTimeIfEmpty(tzOffset int) bool
}

type SimpleTx struct {
	template               string
	userCurrencySuggestion string

	nextFields []*TemplateField
	data       map[string]string
}

type TemplateHintData struct {
	Raw string

	FieldName      string
	FieldSpecifier string
	FieldHint      string
	FieldDefault   string // This is not filled by field parsing logic yet. Instead values can be passed in.
}

type Type string
type HintTemplate struct {
	Text    string
	Handler func(m *tb.Message) (string, error)
}

var TEMPLATE_TYPE_HINTS = map[Type]HintTemplate{
	Type(c.FIELD_AMOUNT): {
		Text:    "Please enter the *amount* of money {{.FieldHint}} (e.g. '12.34' or '12.34 {{.FieldDefault}}')",
		Handler: HandleFloat,
	},
	Type(c.FIELD_ACCOUNT): {
		Text:    "Please enter the *account* {{.FieldHint}} (or select one from the list)",
		Handler: HandleRaw,
	},
	Type(c.FIELD_DESCRIPTION): {
		Text:    "Please enter a *description* {{.FieldHint}} (or select one from the list)",
		Handler: HandleRaw,
	},
}

const TEMPLATE_SIMPLE_DEFAULT = `${date} * "${description}"${tag}
  ${account:from:the money came *from*} ${-amount}
  ${account:to:the money went *to*}`

func CreateSimpleTx(suggestedCur, template string) (Tx, error) {
	tx := (&SimpleTx{
		data:                   make(map[string]string),
		template:               template,
		userCurrencySuggestion: suggestedCur,
	}).Prepare()
	return tx, nil
}

func (tx *SimpleTx) CacheData() (data map[string]string) {
	fieldOrder := []string{}
	fields := ParseTemplateFields(tx.template, "")
	for _, f := range fields {
		if !c.ArrayContains(c.AllowedSuggestionTypes(), c.TypeCacheKey(f.FieldIdentifierForValue())) {
			// Don't cache non-suggestible data
			continue
		}
		fieldOrder = append(fieldOrder, f.FieldIdentifierForValue())
	}
	cleanedData := make(map[string]string)
	for k, d := range tx.data {
		if !c.ArrayContains(fieldOrder, k) {
			continue
		}
		cleanedData[k] = strings.ReplaceAll(d, FORMATTER_PLACEHOLDER, "")
	}
	log.Print(cleanedData)
	return cleanedData
}

func (tx *SimpleTx) Prepare() Tx {
	tx.nextFields = ParseTemplateFields(tx.template, tx.userCurrencySuggestion)
	tx.cleanNextFields()
	return tx
}

func (tx *SimpleTx) SetDate(d string) (Tx, error) {
	date, err := ParseDate(d)
	if err != nil {
		return nil, err
	}
	tx.data[c.FqCacheKey(c.FIELD_DATE)] = date
	return tx, nil
}

func (tx *SimpleTx) setTimeIfEmpty(tzOffset int) bool {
	if tx.data[c.FqCacheKey(c.FIELD_DATE)] == "" {
		// set today as fallback/default date
		timezoneOff := time.Duration(tzOffset) * time.Hour
		tx.data[c.FqCacheKey(c.FIELD_DATE)] = time.Now().UTC().Add(timezoneOff).Format(c.BEANCOUNT_DATE_FORMAT)
		return true
	}
	return false
}

func (tx *SimpleTx) setTagIfEmpty(tag string) bool {
	if tx.data[c.FqCacheKey(c.FIELD_TAG)] == "" {
		tagS := ""
		if tag != "" {
			tagS += " #" + tag
		}
		tx.data[c.FqCacheKey(c.FIELD_TAG)] = tagS
		return true
	}
	return false
}

func SortTemplateFields(unsortedFields []*TemplateField) []*TemplateField {
	sortMapping := map[string]int{
		c.FIELD_AMOUNT:      1,
		c.FIELD_DESCRIPTION: 2,
		c.FIELD_ACCOUNT:     3,
	}
	sort.Slice(unsortedFields, func(i, j int) bool {
		if unsortedFields[i].FieldName == unsortedFields[j].FieldName {
			sortedSpecifiers := []string{unsortedFields[i].FieldSpecifier, unsortedFields[j].FieldSpecifier}
			sort.Strings(sortedSpecifiers)
			return unsortedFields[i].FieldSpecifier == sortedSpecifiers[0]
		}
		a, exists := sortMapping[unsortedFields[i].FieldName]
		if !exists {
			a = len(sortMapping) + 1
		}
		b, exists := sortMapping[unsortedFields[j].FieldName]
		if !exists {
			b = len(sortMapping) + 1
		}
		return a < b
	})
	return unsortedFields
}

func ParseTemplateFields(template, currencySuggestion string) []*TemplateField {
	varBegins := strings.Split(template, "${")
	if len(varBegins) > 1 {
		varBegins = varBegins[1:]
	}
	unsortedFields := []*TemplateField{}
	for _, v := range varBegins {
		field := ParseTemplateField(strings.Split(v, "}")[0], currencySuggestion)
		unsortedFields = append(unsortedFields, field)
	}
	return SortTemplateFields(unsortedFields)
}

type NumberConfig struct {
	Fraction   int
	IsNegative bool
}

type TemplateField struct {
	TemplateHintData
	NumberConfig
}

func (tf *TemplateField) FieldIdentifierForValue() string {
	return tf.FieldName + ":" + tf.FieldSpecifier
}

func ParseTemplateField(rawField, currencySuggestion string) *TemplateField {
	rawField = strings.TrimSpace(rawField)
	field := &TemplateField{
		TemplateHintData{
			Raw: rawField,
		},
		NumberConfig{},
	}

	splitFieldByColon := strings.Split(rawField, ":")
	field.FieldName = strings.TrimSpace(splitFieldByColon[0])
	if len(splitFieldByColon) >= 2 {
		field.FieldSpecifier = strings.TrimSpace(splitFieldByColon[1])
	}
	if len(splitFieldByColon) >= 3 {
		field.FieldHint = strings.TrimSpace(splitFieldByColon[2])
	}
	if field.FieldHint == "" && field.FieldSpecifier != "" {
		field.FieldHint = fmt.Sprintf("*%s*", field.FieldSpecifier)
	}

	field.IsNegative = strings.HasPrefix(field.FieldName, "-")
	field.FieldName = strings.TrimLeft(field.FieldName, "-")

	fractionSplits := strings.Split(field.FieldName, "/")
	field.FieldName = fractionSplits[0]
	field.Fraction = 1
	if len(fractionSplits) == 2 {
		field.FieldName = fractionSplits[0]
		var err error
		field.Fraction, err = strconv.Atoi(fractionSplits[1])
		if err != nil {
			c.LogLocalf(WARN, nil, "converting fraction for template failed: '%s' -> %s", rawField, err.Error())
			field.Fraction = 1
		}
		if field.Fraction == 0 {
			c.LogLocalf(WARN, nil, "fraction was 0. Setting to 1: '%s' -> %s", rawField, err.Error())
			field.Fraction = 1
		}
	}
	field.FieldName = fractionSplits[0]

	if field.FieldName == c.FIELD_AMOUNT {
		field.FieldDefault = currencySuggestion
	}

	return field
}

func (tx *SimpleTx) Input(m *tb.Message) (isDone bool, err error) {
	nextField := tx.nextFields[0]
	hint := TEMPLATE_TYPE_HINTS[Type(nextField.FieldName)]
	res, err := hint.Handler(m)
	if err != nil {
		return tx.IsDone(), err
	}
	tx.data[nextField.FieldIdentifierForValue()] = res
	return tx.IsDone(), nil
}

func (tx *SimpleTx) cleanNextFields() {
	if len(tx.nextFields) > 0 {
		nextField := tx.nextFields[0]
		_, isDataFilled := tx.data[nextField.FieldIdentifierForValue()]
		_, isFieldAutoFilled := TEMPLATE_TYPE_HINTS[Type(nextField.FieldName)]
		if isDataFilled || !isFieldAutoFilled {
			tx.nextFields = tx.nextFields[1:]
			tx.cleanNextFields()
			return
		}
	}
}

func (tx *SimpleTx) NextHint(r *crud.Repo, m *tb.Message) *Hint {
	if len(tx.nextFields) == 0 {
		crud.LogDbf(r, TRACE, m, "During extraction of next hint an error ocurred: step exceeds max index.")
		return nil
	}
	nextField := tx.nextFields[0]
	hint := TEMPLATE_TYPE_HINTS[Type(nextField.FieldName)]
	message, err := c.Template(hint.Text, structs.Map(nextField.TemplateHintData))
	if err != nil {
		crud.LogDbf(r, TRACE, m, "During message building an error ocurred: "+err.Error())
		return nil
	}
	return tx.EnrichHint(r, m, &Input{
		key: nextField.FieldName,
		hint: &Hint{
			Prompt: message,
		},
		handler: hint.Handler,
		field:   *nextField,
	})
}

func (tx *SimpleTx) EnrichHint(r *crud.Repo, m *tb.Message, i *Input) *Hint {
	crud.LogDbf(r, TRACE, m, "Enriching hint (%s).", i.key)
	if i.key == c.FIELD_DESCRIPTION {
		return tx.hintDescription(r, m, i)
	}
	if i.key == c.FIELD_ACCOUNT {
		return tx.hintAccount(r, m, i)
	}
	return i.hint
}

func (tx *SimpleTx) hintAccount(r *crud.Repo, m *tb.Message, i *Input) *Hint {
	accountFQSpecifier := i.field.FieldIdentifierForValue()
	crud.LogDbf(r, TRACE, m, "Enriching hint: '%s'", accountFQSpecifier)
	var (
		res []string = nil
		err error    = nil
	)
	res, err = r.GetCacheHints(m, accountFQSpecifier)
	if err != nil {
		crud.LogDbf(r, ERROR, m, "Error occurred getting cached hint (%s): %s", accountFQSpecifier, err.Error())
		return i.hint
	}
	i.hint.KeyboardOptions = res
	return i.hint
}

func (tx *SimpleTx) hintDescription(r *crud.Repo, m *tb.Message, i *Input) *Hint {
	accountFQSpecifier := i.field.FieldIdentifierForValue()
	res, err := r.GetCacheHints(m, accountFQSpecifier)
	if err != nil {
		crud.LogDbf(r, ERROR, m, "Error occurred getting cached hint (hintDescription): %s", err.Error())
	}
	i.hint.KeyboardOptions = res
	return i.hint
}

func (tx *SimpleTx) IsDone() bool {
	tx.cleanNextFields()
	return len(tx.nextFields) == 0
}

const FORMATTER_PLACEHOLDER = "${SPACE_FORMAT}"

func formatAllLinesWithFormatterPlaceholder(s string, dotIndentation int, currency string) string {
	rebuiltString := ""
	for _, line := range strings.Split(s, "\n") {
		if strings.Contains(line, FORMATTER_PLACEHOLDER) {
			splits := strings.SplitN(line, FORMATTER_PLACEHOLDER, 2)
			firstPart, secondPart := strings.TrimSpace(splits[0]), strings.TrimSpace(splits[1])
			// in case there are multiple occurrences in one line:
			secondPart = strings.TrimSpace(strings.ReplaceAll(secondPart, FORMATTER_PLACEHOLDER, ""))
			firstPart = fmt.Sprintf("  %s ", firstPart) // Two leading spaces and one trailing for separation
			if !strings.Contains(secondPart, " ") {
				secondPart += " " + currency
			}
			if strings.Contains(secondPart, ".") {
				dotSplits := strings.SplitN(secondPart, ".", 2)
				runeCount := 0
				runeCount += utf8.RuneCountInString(firstPart)
				runeCount += utf8.RuneCountInString(dotSplits[0])
				spacesNeeded := dotIndentation - runeCount + 1 // one dot char
				if spacesNeeded < 0 {
					spacesNeeded = 0
				}
				rebuiltString += firstPart + " " + strings.Repeat(" ", spacesNeeded) + secondPart + "\n"
			} else {
				rebuiltString += firstPart + " " + secondPart + "\n"
			}
		} else if line == "" {
			rebuiltString += line
		} else {
			rebuiltString += line + "\n"
		}
	}
	return rebuiltString
}

func (tx *SimpleTx) FillTemplate(currency, tag string, tzOffset int) (string, error) {
	if !tx.IsDone() {
		return "", fmt.Errorf("not all data for this tx has been gathered")
	}
	// If still empty, set time and correct for timezone
	tx.setTimeIfEmpty(tzOffset)
	tx.setTagIfEmpty(tag)

	template := tx.template
	fields := ParseTemplateFields(tx.template, "")
	for _, f := range fields {
		value, exists := tx.data[f.FieldIdentifierForValue()]
		if exists {
			value, err := applyFieldOptionsForNumbersIfApplicable(value, f)
			if err != nil {
				return "", err
			}
			template = strings.ReplaceAll(template, fmt.Sprintf("${%s}", f.Raw), value)
		}
	}
	template = formatAllLinesWithFormatterPlaceholder(template, c.DOT_INDENT, currency)
	return strings.TrimSpace(template) + "\n", nil
}

func applyFieldOptionsForNumbersIfApplicable(value string, f *TemplateField) (string, error) {
	splits := strings.SplitN(value, FORMATTER_PLACEHOLDER, 2)
	if len(splits) > 1 {
		leftSide, rightSide := splits[0], splits[1]
		if f.IsNegative {
			if strings.HasPrefix(rightSide, "-") {
				rightSide = rightSide[1:]
			} else {
				rightSide = "-" + rightSide
			}
		}
		if f.Fraction > 1 {
			amountSplits := strings.SplitN(rightSide, " ", 2)
			amountLeft := amountSplits[0]
			currency := ""
			if len(amountSplits) > 1 {
				currency = amountSplits[1]
			}
			amountParsed, err := strconv.ParseFloat(amountLeft, 64)
			if err != nil {
				return "", err
			}
			amountParsed /= float64(f.Fraction)
			rightSide = ParseAmount(amountParsed) + " " + currency
		}
		return leftSide + FORMATTER_PLACEHOLDER + rightSide, nil
	}
	return value, nil
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
	return fmt.Sprintf("SimpleTx{remainingFields=%v, data=%v}", len(tx.nextFields), tx.data)
}
