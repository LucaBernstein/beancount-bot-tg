package helpers

import (
	"fmt"
	"log"
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"
)

type handlerFunc func(m *tb.Message, params ...string)

type SubcommandHandler struct {
	base               string
	quotedSingleParams bool
	mappings           map[string]handlerFunc
	separator          string
}

func MakeSubcommandHandler(base string, quotedSingleParams bool) *SubcommandHandler {
	return &SubcommandHandler{
		base:               base,
		mappings:           make(map[string]handlerFunc),
		quotedSingleParams: quotedSingleParams,
		separator:          " ",
	}
}

func (sh *SubcommandHandler) SetSeparator(sep string) *SubcommandHandler {
	sh.separator = sep
	return sh
}

func (sh *SubcommandHandler) Add(command string, handler handlerFunc) *SubcommandHandler {
	if strings.Contains(command, sh.separator) {
		log.Printf("Warning: subcommand '%s' contains a space. This most probably won't work with set separator '%s'",
			command, sh.separator)
	}
	_, exists := sh.mappings[command]
	if exists {
		log.Printf("Warning: subcommand '%s' already exists. Will ignore remapping.", command)
	}
	sh.mappings[command] = handler
	return sh
}

func (sh *SubcommandHandler) Handle(m *tb.Message) error {
	commandRemainder := strings.TrimSpace(strings.TrimPrefix(m.Text, sh.base))
	parameters := strings.Split(commandRemainder, sh.separator)

	var subcommand string
	if len(parameters) > 0 {
		subcommand = parameters[0]
	}
	fn, exists := sh.mappings[subcommand]
	if !exists {
		return fmt.Errorf("subcommand '%s' has not been mapped with this SubcommandHandler(%s)", subcommand, sh.base)
	}
	if len(parameters) <= 1 {
		fn(m)
		return nil
	}

	// More than just subcommand is left.
	// Split by \" and interpret place odd as unquoted and even as quoted
	// e.g. 'a b "c d" e f'
	remainingCommand := strings.TrimSpace(strings.TrimPrefix(commandRemainder, subcommand))
	parameters = []string{}
	params := strings.Split(remainingCommand, "\"")
	for i, e := range params {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}
		if (i+1)%2 == 0 { // Even: Quoted
			parameters = append(parameters, e)
		} else { // Odd: Unquoted
			parameters = append(parameters, strings.Split(e, " ")...)
		}
	}
	log.Print(parameters)
	fn(m, parameters...)
	return nil
}
