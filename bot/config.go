package bot

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/LucaBernstein/beancount-bot-tg/db/crud"
	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/telebot.v3"
)

func (bc *BotController) configHandler(m *tb.Message) {
	sc := helpers.MakeSubcommandHandler("/"+CMD_CONFIG, true)
	sc.
		Add("currency", bc.configHandleCurrency).
		Add("tag", bc.configHandleTag).
		Add("notify", bc.configHandleNotification).
		Add("about", bc.configHandleAbout).
		Add("tz_offset", bc.configHandleTimezoneOffset).
		Add("delete_account", bc.configHandleAccountDelete).
		Add("omit_slash", bc.configHandleOmitLeadingSlash).
		Add("enable_api", bc.configHandleEnableApi)
	_, err := sc.Handle(m)
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

Tags will be added to each new transaction with a '#':

/{{.CONFIG_COMMAND}} tag - Get currently set tag
/{{.CONFIG_COMMAND}} tag off - Turn off tag
/{{.CONFIG_COMMAND}} tag <name> - Set tag to apply to new transactions, e.g. when on vacation

Create a schedule to be notified of open transactions (i.e. not archived or deleted):

/{{.CONFIG_COMMAND}} notify - Get current notification status
/{{.CONFIG_COMMAND}} notify off - Disable reminder notifications
/{{.CONFIG_COMMAND}} notify <delay> <hour> - Notify of open transaction after <delay> days at <hour> of the day. Honors configured timezone offset (see below)

Timezone offset from {{.TZ}} to honor for notifications and current date (if set automatically) in new transactions:

/{{.CONFIG_COMMAND}} tz_offset - Get current timezone offset from {{.TZ}} (default 0)
/{{.CONFIG_COMMAND}} tz_offset <hours> - Set timezone offset from {{.TZ}}

Feature toggle: Also activate commands without leading slash if not in transaction

/{{.CONFIG_COMMAND}} omit_slash - Get current setting value
/{{.CONFIG_COMMAND}} omit_slash on|off - Enable or disable omitted leading slash support

Feature toggle: Activate API / UI usage

/{{.CONFIG_COMMAND}} enable_api - Get current setting value
/{{.CONFIG_COMMAND}} enable_api on|off

Additional information about this bot

/{{.CONFIG_COMMAND}} about - Display the version this bot is running on

Reset your data stored by the bot. WARNING: This action is permanent!

/{{.CONFIG_COMMAND}} delete_account yes - Permanently delete all account-related data
`, map[string]interface{}{
		"CONFIG_COMMAND": CMD_CONFIG,
		"TZ":             tz,
	})
	if err != nil {
		bc.Logf(ERROR, m, "Parsing configHelp template failed: %s", err.Error())
	}

	bc.Bot.SendSilent(bc, Recipient(m), errorMsg+filledTemplate)
}

func (bc *BotController) configHandleCurrency(m *tb.Message, params ...string) {
	currency := bc.Repo.UserGetCurrency(m)
	if len(params) == 0 { // 0 params: GET currency
		// Return currently set currency
		bc.Bot.SendSilent(bc, Recipient(m), fmt.Sprintf("Your current currency is set to '%s'. To change it add the new currency to use to the command like this: '/%s currency EUR'.", currency, CMD_CONFIG))
		return
	} else if len(params) > 1 { // 2 or more params: too many
		bc.configHelp(m, fmt.Errorf("invalid amount of parameters specified"))
		return
	}
	// Set new currency
	newCurrency := params[0]
	err := bc.Repo.UserSetCurrency(m, newCurrency)
	if err != nil {
		bc.Bot.SendSilent(bc, Recipient(m), "An error ocurred saving your currency preference: "+err.Error())
		return
	}
	bc.Bot.SendSilent(bc, Recipient(m), fmt.Sprintf("Changed default currency for all future transactions from '%s' to '%s'.", currency, newCurrency))
}

func (bc *BotController) configHandleTag(m *tb.Message, params ...string) {
	if len(params) == 0 {
		// GET tag
		tag := bc.Repo.UserGetTag(m)
		if tag != "" {
			bc.Bot.SendSilent(bc, Recipient(m), fmt.Sprintf("All new transactions automatically get the tag #%s added (vacation mode enabled)", tag))
		} else {
			bc.Bot.SendSilent(bc, Recipient(m), "No tags are currently added to new transactions (vacation mode disabled).")
		}
		return
	} else if len(params) > 1 { // Only 0 or 1 allowed
		bc.configHelp(m, fmt.Errorf("invalid amount of parameters specified"))
		return
	}
	if params[0] == "off" {
		// DELETE tag
		bc.Repo.UserSetTag(m, "")
		bc.Bot.SendSilent(bc, Recipient(m), "Disabled automatically set tags on new transactions")
		return
	}
	// SET tag
	tag := strings.TrimPrefix(params[0], "#")
	err := bc.Repo.UserSetTag(m, tag)
	if err != nil {
		bc.Bot.SendSilent(bc, Recipient(m), "An error ocurred saving the tag: "+err.Error())
		return
	}
	bc.Bot.SendSilent(bc, Recipient(m), fmt.Sprintf("From now on all new transactions automatically get the tag #%s added (vacation mode enabled)", tag))
}

func (bc *BotController) configHandleNotification(m *tb.Message, params ...string) {
	var tz, _ = time.Now().Zone()
	userTzOffset := bc.Repo.UserGetTzOffset(m)
	if userTzOffset < 0 {
		tz += strconv.Itoa(userTzOffset)
	} else {
		tz += "+" + strconv.Itoa(userTzOffset)
	}
	if len(params) == 0 {
		// GET schedule
		daysDelay, hour, err := bc.Repo.UserGetNotificationSetting(m)
		if err != nil {
			bc.configHelp(m, fmt.Errorf("an application error occurred while retrieving user information from database"))
			return
		}
		if daysDelay < 0 {
			bc.Bot.SendSilent(bc, Recipient(m), "Notifications are disabled for open transactions.")
			return
		}
		plural_s := "s"
		if daysDelay == 1 {
			plural_s = ""
		}
		bc.Bot.SendSilent(bc, Recipient(m), fmt.Sprintf("The bot will notify you daily at hour %d (%s) if transactions are open for more than %d day%s", hour, tz, daysDelay, plural_s))
		return
	} else if len(params) == 1 {
		// DELETE schedule
		if params[0] == "off" {
			err := bc.Repo.UserSetNotificationSetting(m, -1, -1)
			if err != nil {
				bc.configHelp(m, fmt.Errorf("error setting notification schedule: %s", err.Error()))
				return
			}
			bc.Bot.SendSilent(bc, Recipient(m), "Successfully disabled notifications for open transactions.")
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
	bc.Bot.SendSilent(bc, Recipient(m), escapeCharacters(fmt.Sprintf(`Version information about [LucaBernstein/beancount-bot-tg](https://github.com/LucaBernstein/beancount-bot-tg)

Version: [%s](%s)`,
		version,
		versionLink,
	), ".", "-"), tb.ModeMarkdownV2)
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
		bc.Bot.SendSilent(bc, Recipient(m), fmt.Sprintf("Your current timezone offset is set to 'UTC%s'.", prettyTzOffset(tz_offset)))
		return
	} else if len(params) > 1 { // 2 or more params: too many
		bc.configHelp(m, fmt.Errorf("invalid amount of parameters specified"))
		return
	}
	// Set new tz_offset
	newTzOffset := params[0]
	newTzParsed, err := strconv.Atoi(newTzOffset)
	if err != nil {
		bc.Bot.SendSilent(bc, Recipient(m), "An error ocurred saving your timezone offset preference: "+err.Error())
		return
	}
	err = bc.Repo.UserSetTzOffset(m, newTzParsed)
	if err != nil {
		bc.Bot.SendSilent(bc, Recipient(m), "An error ocurred saving your timezone offset preference: "+err.Error())
		return
	}
	bc.Bot.SendSilent(bc, Recipient(m), fmt.Sprintf("Changed timezone offset for default dates for all future transactions from 'UTC%s' to 'UTC%s'.", prettyTzOffset(tz_offset), prettyTzOffset(newTzParsed)))
}

func (bc *BotController) configHandleOmitLeadingSlash(m *tb.Message, params ...string) {
	bc.configHandleBooleanFeature(m, helpers.USERSET_OMITCMDSLASH, "Omitting leading slash support", params...)
}

func (bc *BotController) configHandleEnableApi(m *tb.Message, params ...string) {
	state := bc.configHandleBooleanFeature(m, helpers.USERSET_ENABLEAPI, "API support", params...)
	if state {
		var chatInformation string
		if crud.IsGroupChat(m) {
			chatInformation = "group chat"
		} else {
			chatInformation = "direct chat"
		}
		bc.Bot.SendSilent(bc, Recipient(m), fmt.Sprintf("Your chat ID to use for token verification: %d (%s with bot, make one request at a time, multiple tokens can be generated sequentially)", m.Chat.ID, chatInformation))
	}
}

func (bc *BotController) configHandleBooleanFeature(m *tb.Message, key string, name string, params ...string) (state bool) {
	var err error
	if len(params) == 0 { // 0 params: GET
		exists, value, err := bc.Repo.GetUserSetting(key, m.Chat.ID)
		if err != nil {
			bc.Bot.SendSilent(bc, Recipient(m), "There has been an error internally while retrieving the value currently set for the queried user setting. Please try again later.")
			return
		}
		if !exists || strings.ToUpper(value) != "TRUE" {
			bc.Bot.SendSilent(bc, Recipient(m), fmt.Sprintf("%s is currently turned off. Please check the help on how to turn it on.", name))
			return
		} else {
			bc.Bot.SendSilent(bc, Recipient(m), fmt.Sprintf("%s is currently turned on.", name))
			return true
		}
	} else if len(params) > 1 { // 2 or more params: too many
		bc.configHelp(m, fmt.Errorf("invalid amount of parameters specified"))
		return
	}
	// Set user setting
	newUserSettingValue := params[0]
	switch strings.ToUpper(newUserSettingValue) {
	case "ON":
		err = bc.Repo.SetUserSetting(key, "true", m.Chat.ID)
		if err != nil {
			bc.Bot.SendSilent(bc, Recipient(m), "There has been an error internally while setting your value. Please try again later.")
			return
		}
		bc.Bot.SendSilent(bc, Recipient(m), fmt.Sprintf("%s has successfully been turned on.", name))
		return true
	case "OFF":
		err = bc.Repo.SetUserSetting(key, "false", m.Chat.ID)
		if err != nil {
			bc.Bot.SendSilent(bc, Recipient(m), "There has been an error internally while setting your value. Please try again later.")
			return
		}
		bc.Bot.SendSilent(bc, Recipient(m), fmt.Sprintf("%s has successfully been turned off.", name))
		return
	default:
		bc.configHelp(m, fmt.Errorf("invalid setting value: '%s'. Not in ['ON', 'OFF']", newUserSettingValue))
		return
	}
}

func prettyTzOffset(tzOffset int) string {
	if tzOffset < 0 {
		return strconv.Itoa(tzOffset)
	}
	return "+" + strconv.Itoa(tzOffset)
}

type errors struct {
	operation string
	m         *tb.Message
	bc        *BotController
}

func (e *errors) handle1(err error) {
	if err != nil {
		e.bc.Logf(ERROR, e.m, "Handling error for operation '%s' (failing silently, proceeding): %s", e.operation, err.Error())
	}
}
