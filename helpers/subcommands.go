package helpers

import (
	"fmt"
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"
)

type handlerFunc func(m *tb.Message, params ...string)

type SubcommandHandler struct {
	base               string
	quotedSingleParams bool
	mappings           map[string]handlerFunc
}

func MakeSubcommandHandler(base string, quotedSingleParams bool) *SubcommandHandler {
	return &SubcommandHandler{
		base:               base,
		mappings:           make(map[string]handlerFunc),
		quotedSingleParams: quotedSingleParams,
	}
}

func (sh *SubcommandHandler) Add(command string, handler handlerFunc) *SubcommandHandler {
	if strings.Contains(command, " ") {
		LogLocalf(WARN, nil, "Subcommand '%s' contains a space. This most probably won't work with space (' ') separator", command)
	}
	_, exists := sh.mappings[command]
	if exists {
		LogLocalf(WARN, nil, "Subcommand '%s' already exists. Performing remapping! Please check whether this is desired behavior.", command)
	}
	sh.mappings[command] = handler
	return sh
}

func (sh *SubcommandHandler) Handle(m *tb.Message) error {
	commandRemainder := strings.TrimSpace(strings.TrimPrefix(m.Text, sh.base))
	parameters := strings.Split(commandRemainder, " ")

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

	remainingCommand := strings.TrimSpace(strings.TrimPrefix(commandRemainder, subcommand))
	parameters = SplitQuotedCommand(remainingCommand)
	if ArraysEqual(parameters, []string{}) {
		return fmt.Errorf("this remainingCommand could not be split: '%s'", remainingCommand)
	}
	fn(m, parameters...)
	return nil
}

type TV struct {
	T     string
	Value string
}

func ExtractTypeValue(params ...string) (*TV, error) {
	e := &TV{}
	if len(params) < 1 || len(params) > 2 {
		return nil, fmt.Errorf("unexpected count of parameters")
	}
	if len(params) >= 1 {
		e.T = params[0]
	}
	if len(params) >= 2 {
		e.Value = params[1]
	}
	return e, nil
}

func SplitQuotedCommand(s string) (res []string) {
	isEscaped := false
	isQuoted := false
	split := ""
	for _, c := range s {
		if isEscaped {
			isEscaped = false
			split += string(c)
			continue
		}
		if c == '\\' {
			isEscaped = true
			continue
		}
		if c == '"' {
			isQuoted = !isQuoted
			continue
		}
		if c == ' ' && !isQuoted {
			if split == "" {
				continue
			}
			res = append(res, split)
			split = ""
			continue
		}
		split += string(c)
	}
	if split != "" {
		res = append(res, split)
	}
	if isEscaped || isQuoted {
		return []string{}
	}
	return
}
