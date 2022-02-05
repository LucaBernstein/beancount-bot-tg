package bot

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (bc *BotController) configHandler(m *tb.Message) {
	sc := helpers.MakeSubcommandHandler("/"+CMD_CONFIG, true)
	sc.
		Add("currency", bc.configHandleCurrency).
		Add("tag", bc.configHandleTag).
		Add("notify", bc.configHandleNotification).
		Add("limit", bc.configHandleLimit).
		Add("about", bc.configHandleAbout).
		Add("tz_offset", bc.configHandleTimezoneOffset)
	err := sc.Handle(m)
	if err != nil {
		bc.configHelp(m, nil)
	}
}

func (bc *BotController) configHelp(m *tb.Message, err error) {
	errorMsg := ""
	if err != nil {
		errorMsg += fmt.Sprintf("Error executing your command: %s\n\n", err.Error())
	}

	tz, _ := time.Now().Zone()
	filledTemplate, err := helpers.Template(`Usage help for /{{.CONFIG_COMMAND}}:

/{{.CONFIG_COMMAND}} currency <c> - Change default currency
/{{.CONFIG_COMMAND}} about - Display the version this bot is running on

Tags will be added to each new transaction with a '#':

/{{.CONFIG_COMMAND}} tag - Get currently set tag
/{{.CONFIG_COMMAND}} tag off - Turn off tag
/{{.CONFIG_COMMAND}} tag <name> - Set tag to apply to new transactions, e.g. when on vacation

Create a schedule to be notified of open transactions (i.e. not archived or deleted):

/{{.CONFIG_COMMAND}} notify - Get current notification status
/{{.CONFIG_COMMAND}} notify off - Disable reminder notifications
/{{.CONFIG_COMMAND}} notify <delay> <hour> - Notify of open transaction after <delay> days at <hour> of the day ({{.TZ}})",

Set suggestion cache limits (i.e. only cache new values until limit is reached, then old ones get dismissed if new ones are added):

/{{.CONFIG_COMMAND}} limit - Get currently set cache limits
/{{.CONFIG_COMMAND}} limit <suggestionType> <amount>|off - Set or disable suggestion limit for a type

Set timezone offset from UTC for transactions where date is added automatically:

/{{.CONFIG_COMMAND}} tz_offset - Get current timezone offset from UTC
/{{.CONFIG_COMMAND}} tz_offset <hours> - Set timezone offset from UTC, default 0
`, map[string]interface{}{
		"CONFIG_COMMAND": CMD_CONFIG,
		"TZ":             tz,
	})
	if err != nil {
		bc.Logf(ERROR, m, "Parsing configHelp template failed: %s", err.Error())
	}

	_, err = bc.Bot.Send(m.Sender, errorMsg+filledTemplate)
	if err != nil {
		bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
	}
}

func (bc *BotController) configHandleCurrency(m *tb.Message, params ...string) {
	currency := bc.Repo.UserGetCurrency(m)
	if len(params) == 0 { // 0 params: GET currency
		// Return currently set currency
		_, err := bc.Bot.Send(m.Sender, fmt.Sprintf("Your current currency is set to '%s'. To change it add the new currency to use to the command like this: '/%s currency EUR'.", currency, CMD_CONFIG))
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	} else if len(params) > 1 { // 2 or more params: too many
		bc.configHelp(m, fmt.Errorf("invalid amount of parameters specified"))
		return
	}
	// Set new currency
	newCurrency := params[0]
	err := bc.Repo.UserSetCurrency(m, newCurrency)
	if err != nil {
		_, err = bc.Bot.Send(m.Sender, "An error ocurred saving your currency preference: "+err.Error())
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}
	_, err = bc.Bot.Send(m.Sender, fmt.Sprintf("Changed default currency for all future transactions from '%s' to '%s'.", currency, newCurrency))
	if err != nil {
		bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
	}
}

func (bc *BotController) configHandleTag(m *tb.Message, params ...string) {
	if len(params) == 0 {
		// GET tag
		tag := bc.Repo.UserGetTag(m)
		if tag != "" {
			_, err := bc.Bot.Send(m.Sender, fmt.Sprintf("All new transactions automatically get the tag #%s added (vacation mode enabled)", tag))
			if err != nil {
				bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
			}
		} else {
			_, err := bc.Bot.Send(m.Sender, "No tags are currently added to new transactions (vacation mode disabled).")
			if err != nil {
				bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
			}
		}
		return
	} else if len(params) > 1 { // Only 0 or 1 allowed
		bc.configHelp(m, fmt.Errorf("invalid amount of parameters specified"))
		return
	}
	if params[0] == "off" {
		// DELETE tag
		bc.Repo.UserSetTag(m, "")
		_, err := bc.Bot.Send(m.Sender, "Disabled automatically set tags on new transactions")
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}
	// SET tag
	tag := strings.TrimPrefix(params[0], "#")
	err := bc.Repo.UserSetTag(m, tag)
	if err != nil {
		_, err = bc.Bot.Send(m.Sender, "An error ocurred saving the tag: "+err.Error())
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}
	_, err = bc.Bot.Send(m.Sender, fmt.Sprintf("From now on all new transactions automatically get the tag #%s added (vacation mode enabled)", tag))
	if err != nil {
		bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
	}
}

func (bc *BotController) configHandleNotification(m *tb.Message, params ...string) {
	var tz, _ = time.Now().Zone()
	if len(params) == 0 {
		// GET schedule
		daysDelay, hour, err := bc.Repo.UserGetNotificationSetting(m)
		if err != nil {
			bc.configHelp(m, fmt.Errorf("an application error occurred while retrieving user information from database"))
			return
		}
		if daysDelay < 0 {
			_, err = bc.Bot.Send(m.Sender, "Notifications are disabled for open transactions.")
			if err != nil {
				bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
			}
			return
		}
		plural_s := "s"
		if daysDelay == 1 {
			plural_s = ""
		}
		_, err = bc.Bot.Send(m.Sender, fmt.Sprintf("The bot will notify you daily at hour %d (%s) if transactions are open for more than %d day%s", hour, tz, daysDelay, plural_s))
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	} else if len(params) == 1 {
		// DELETE schedule
		if params[0] == "off" {
			err := bc.Repo.UserSetNotificationSetting(m, -1, -1)
			if err != nil {
				bc.configHelp(m, fmt.Errorf("error setting notification schedule: %s", err.Error()))
				return
			}
			_, err = bc.Bot.Send(m.Sender, "Successfully disabled notifications for open transactions.")
			if err != nil {
				bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
			}
			return
		}
		bc.configHelp(m, fmt.Errorf("invalid parameters"))
		return
	} else if len(params) == 2 {
		// SET schedule
		daysDelay, err := strconv.Atoi(params[0])
		if err != nil {
			bc.configHelp(m, fmt.Errorf("error converting daysDelay to number: %s: %s", params[0], err.Error()))
			return
		}
		hour, err := strconv.Atoi(params[1])
		if err != nil {
			bc.configHelp(m, fmt.Errorf("error converting hour to number: %s: %s", params[1], err.Error()))
			return
		}
		if hour > 23 || hour < 0 {
			bc.configHelp(m, fmt.Errorf("invalid hour (%d is out of valid range 1-23)", hour))
			return
		}
		err = bc.Repo.UserSetNotificationSetting(m, daysDelay, hour)
		if err != nil {
			bc.configHelp(m, fmt.Errorf("error setting notification schedule: %s", err.Error()))
		}
		bc.configHandleNotification(m) // Recursively call with zero params --> GET
		return
	}
	bc.configHelp(m, fmt.Errorf("invalid amount of parameters specified"))
}

func (bc *BotController) configHandleLimit(m *tb.Message, params ...string) {
	if len(params) == 0 {
		// GET limits for all types
		cacheLimits, err := bc.Repo.CacheUserSettingGetLimits(m)
		if err != nil {
			bc.Logf(WARN, m, "CacheUserSettingGetLimits failed: %s", err.Error())
			bc.configHelp(m, fmt.Errorf("could not get your cache limits"))
			return
		}
		message := "You have the following cache limits configured:\n"
		for limit, value := range cacheLimits {
			message += fmt.Sprintf("\n%s: %d", limit, value)
		}
		message += "\n\n"
		message += "If new cache entries are created for the given types, old ones are automatically deleted.\nPlease note: If suggestions were added using /suggestions add, they will not be deleted automatically by this mechanism. '-1' means no limit."
		_, err = bc.Bot.Send(m.Sender, message)
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	} else if len(params) == 1 {
		// GET limit for type
		cacheType := params[0]
		cacheLimits, err := bc.Repo.CacheUserSettingGetLimits(m)
		if err != nil {
			bc.Logf(WARN, m, "CacheUserSettingGetLimits failed: %s", err.Error())
			bc.configHelp(m, fmt.Errorf("an application error occurred while retrieving cache limits from database"))
			return
		}
		message := "You have the following cache limit configured for '" + cacheType + "': " + strconv.Itoa(cacheLimits[cacheType])
		message += "\n\n"
		message += "If new cache entries are created for the given types, old ones are automatically deleted.\nPlease note: If suggestions were added using /suggestions add, they will not be deleted automatically by this mechanism. '-1' means no limit."
		_, err = bc.Bot.Send(m.Sender, message)
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	} else if len(params) == 2 {
		// SET limit for type
		cacheType := params[0]
		limitValue := params[1]

		if !helpers.ArrayContains(helpers.AllowedSuggestionTypes(), cacheType) {
			bc.configHelp(m, fmt.Errorf("unknown suggestion type: %s. Must be one from: %v", cacheType, helpers.AllowedSuggestionTypes()))
			return
		}

		limitValueParsed, errParsing := strconv.Atoi(limitValue)
		if errParsing != nil {
			bc.configHelp(m, fmt.Errorf("an application error occurred while interpreting your amount as a number to set the limit to"))
			return
		}

		if limitValue == "off" || (errParsing == nil && limitValueParsed < 0) {
			err := bc.Repo.CacheUserSettingSetLimit(m, cacheType, -1)
			if err != nil {
				bc.Logf(ERROR, m, "Error disabling suggestions cache limit: %s", err.Error())
				bc.configHelp(m, fmt.Errorf("an application error occurred while disabling suggestions cache: %s", err.Error()))
				return
			}
			_, err = bc.Bot.Send(m.Sender, fmt.Sprintf("Successfully disabled suggestions cache limits for type %s", cacheType))
			if err != nil {
				bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
			}
		} else {
			err := bc.Repo.CacheUserSettingSetLimit(m, cacheType, limitValueParsed)
			if err != nil {
				bc.Logf(ERROR, m, "Error setting suggestions cache limit: %s", err.Error())
				bc.configHelp(m, fmt.Errorf("an application error occurred while setting a new suggestions cache limit: %s", err.Error()))
				return
			}
			_, err = bc.Bot.Send(m.Sender, fmt.Sprintf("Successfully set suggestions cache limits for type %s to %d", cacheType, limitValueParsed))
			if err != nil {
				bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
			}
		}
		err := bc.Repo.PruneUserCachedSuggestions(m)
		if err != nil {
			bc.Logf(WARN, m, "Could not prune suggestions: %s", err.Error())
		}
		return
	}

	bc.configHelp(m, fmt.Errorf("invalid amount of parameters specified"))
}

func (bc *BotController) configHandleAbout(m *tb.Message, params ...string) {
	if len(params) > 0 {
		bc.configHelp(m, fmt.Errorf("no parameters expected"))
		return
	}
	version := os.Getenv("VERSION")
	versionLink := "https://github.com/LucaBernstein/beancount-bot-tg/releases/"
	if strings.HasPrefix(version, "v") {
		versionLink += "tag/" + version
	}
	if version == "" {
		version = "not specified"
	}
	_, err := bc.Bot.Send(m.Sender, escapeCharacters(fmt.Sprintf(`Version information about [LucaBernstein/beancount-bot-tg](https://github.com/LucaBernstein/beancount-bot-tg)

Version: [%s](%s)`,
		version,
		versionLink,
	), ".", "-"), tb.ModeMarkdownV2)
	if err != nil {
		bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
	}
}

func escapeCharacters(s string, c ...string) string {
	for _, char := range c {
		s = strings.ReplaceAll(s, char, "\\"+char)
	}
	return s
}

func (bc *BotController) configHandleTimezoneOffset(m *tb.Message, params ...string) {
	tz_offset := bc.Repo.UserGetTzOffset(m)
	if len(params) == 0 { // 0 params: GET
		_, err := bc.Bot.Send(m.Sender, fmt.Sprintf("Your current timezone offset is set to 'UTC%s'.", prettyTzOffset(tz_offset)))
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	} else if len(params) > 1 { // 2 or more params: too many
		bc.configHelp(m, fmt.Errorf("invalid amount of parameters specified"))
		return
	}
	// Set new tz_offset
	newTzOffset := params[0]
	newTzParsed, err := strconv.Atoi(newTzOffset)
	if err != nil {
		_, err = bc.Bot.Send(m.Sender, "An error ocurred saving your timezone offset preference: "+err.Error())
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}
	err = bc.Repo.UserSetTzOffset(m, newTzParsed)
	if err != nil {
		_, err = bc.Bot.Send(m.Sender, "An error ocurred saving your timezone offset preference: "+err.Error())
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}
	_, err = bc.Bot.Send(m.Sender, fmt.Sprintf("Changed timezone offset for default dates for all future transactions from 'UTC%s' to 'UTC%s'.", prettyTzOffset(tz_offset), prettyTzOffset(newTzParsed)))
	if err != nil {
		bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
	}
}

func prettyTzOffset(tzOffset int) string {
	if tzOffset < 0 {
		return strconv.Itoa(tzOffset)
	}
	return "+" + strconv.Itoa(tzOffset)
}
