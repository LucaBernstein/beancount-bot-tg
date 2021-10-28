package bot

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"
)

type command string
type hint string
type data string

type Input struct {
	hint    hint
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
	NextHint() hint
	IsDone() bool
	Debug() string
	addStep(command command, hint hint, handler func(m *tb.Message) (string, error)) Tx
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
		addStep("date", "Please enter the transaction data in the format YYYY-MM-DD (or select one from the list)", HandleDate)
}

func (tx *SimpleTx) addStep(command command, hint hint, handler func(m *tb.Message) (string, error)) Tx {
	tx.steps = append(tx.steps, command)
	tx.stepDetails[command] = Input{hint: hint, handler: handler}
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

func (tx *SimpleTx) NextHint() hint {
	if tx.step > len(tx.steps)-1 {
		log.Printf("During extraction of next hint an error ocurred: step exceeds max index.")
		return ""
	}
	return (hint)(tx.stepDetails[tx.steps[tx.step]].hint)
}

func (tx *SimpleTx) IsDone() bool {
	return tx.step >= len(tx.steps)
}

func (tx *SimpleTx) Debug() string {
	return fmt.Sprintf("SimpleTx{step=%d, totalSteps=%d, data=%v}", tx.step, len(tx.steps), tx.data)
}
