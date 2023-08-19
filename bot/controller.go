package bot

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	dbWrapper "github.com/LucaBernstein/beancount-bot-tg/v2/db"
	"github.com/LucaBernstein/beancount-bot-tg/v2/db/crud"
	"github.com/LucaBernstein/beancount-bot-tg/v2/helpers"
	"github.com/go-co-op/gocron"
	tb "gopkg.in/telebot.v3"
)

type CMD struct {
	CommandAlias []string
	Optional     []string
	Handler      tb.HandlerFunc
	Help         string
}

func NewBotController(db dbWrapper.DB) *BotController {
	return &BotController{
		Repo:  crud.NewRepo(db),
		State: NewStateHandler(),
	}
}

type BotController struct {
	Repo  *crud.Repo
	State *StateHandler
	Bot   IBot

	CronScheduler *gocron.Scheduler
}

func (bc *BotController) ConfigureCronScheduler() *BotController {
	s := gocron.NewScheduler(time.UTC)
	s.Cron("0 * * * *").Do(bc.cronNotifications)
	bc.CronScheduler = s
	return bc
}

func (bc *BotController) AddBotAndStart(b IBot) {
	bc.Bot = b

	mappings := bc.commandMappings()

	for _, m := range mappings {
		for _, alias := range m.CommandAlias {
			b.Handle("/"+alias, m.Handler)
		}
	}

	b.Handle(tb.OnText, bc.handleTextState)

	bc.Logf(TRACE, nil, "Starting bot '%s'", b.Me().Username)

	if bc.CronScheduler != nil {
		bc.CronScheduler.StartAsync()
	} else {
		bc.Logf(WARN, nil, "No cron scheduler has been attached!")
	}
	b.Start() // Blocking
}

const (
	CMD_START       = "start"
	CMD_HELP        = "help"
	CMD_CANCEL      = "cancel"
	CMD_SIMPLE      = "simple"
	CMD_LIST        = "list"
	CMD_ARCHIVE_ALL = "archiveAll"
	CMD_DELETE_ALL  = "deleteAll"
	CMD_SUGGEST     = "suggestions"
	CMD_CONFIG      = "config"

	CMD_ADM_NOTIFY = "admin_notify"
	CMD_ADM_CRON   = "admin_cron"
)

var (
	CMD_COMMENT  = []string{"comment", "c"}
	CMD_TEMPLATE = []string{"template", "t"}
)

func (bc *BotController) commandMappings() []*CMD {
	return []*CMD{
		{CommandAlias: []string{CMD_HELP}, Handler: bc.commandHelp, Help: "List this command help"},
		{CommandAlias: []string{CMD_START}, Handler: bc.commandStart, Help: "Give introduction into this bot"},
		{CommandAlias: []string{CMD_CANCEL}, Handler: bc.commandCancel, Help: "Cancel any running commands or transactions"},
		{CommandAlias: []string{CMD_SIMPLE}, Handler: bc.commandCreateSimpleTx, Help: "Record a simple transaction, defaults to today; Can be omitted by sending amount directy", Optional: []string{"date"}},
		{CommandAlias: CMD_COMMENT, Handler: bc.commandAddComment, Help: "Add arbitrary text to transaction list"},
		{CommandAlias: CMD_TEMPLATE, Handler: bc.commandTemplates, Help: "Create and use template transactions"},
		{CommandAlias: []string{CMD_LIST}, Handler: bc.commandList, Help: "List your recorded transactions or remove entries", Optional: []string{"archived", "dated", "numbered", "rm <number>"}},
		{CommandAlias: []string{CMD_SUGGEST}, Handler: bc.commandSuggestions, Help: "List, add or remove suggestions"},
		{CommandAlias: []string{CMD_CONFIG}, Handler: bc.commandConfig, Help: "Bot configurations"},
		{CommandAlias: []string{CMD_ARCHIVE_ALL}, Handler: bc.commandArchiveTransactions, Help: "Archive recorded transactions"},
		{CommandAlias: []string{CMD_DELETE_ALL}, Handler: bc.commandDeleteTransactions, Help: "Permanently delete recorded transactions"},

		{CommandAlias: []string{CMD_ADM_NOTIFY}, Handler: bc.commandAdminNofify, Help: "Send notification to user(s): /" + CMD_ADM_NOTIFY + " [chatId] \"<message>\""},
		{CommandAlias: []string{CMD_ADM_CRON}, Handler: bc.commandAdminCronInfo, Help: "Check cron status"},
	}
}

func (bc *BotController) commandStart(c tb.Context) error {
	bc.Logf(TRACE, c.Message(), "Start command")
	bc.Bot.SendSilent(bc.Logf, Recipient(c.Message()), "Welcome to this beancount bot!\n"+
		"You can find more information in the repository under "+
		"https://github.com/LucaBernstein/beancount-bot-tg\n\n"+
		"Please check the commands I will send to you next that are available to you. "+
		"You can always reach the command help by typing /"+CMD_HELP, clearKeyboard())
	bc.commandHelp(c)
	return nil
}

func (bc *BotController) commandHelp(c tb.Context) error {
	bc.Logf(TRACE, c.Message(), "Sending help")
	helpMsg := ""
	adminCommands := []*CMD{}
	for i, cmd := range bc.commandMappings() {
		if cmd.Help == "" {
			continue
		}
		if strings.HasPrefix(cmd.CommandAlias[0], "admin") {
			adminCommands = append(adminCommands, cmd)
			continue
		}
		if i != 0 {
			helpMsg += "\n"
		}
		var optional string
		if cmd.Optional != nil {
			for _, opt := range cmd.Optional {
				optional += " [" + opt + "]"
			}
		}
		helpMsg += fmt.Sprintf("/%s%s - %s", cmd.CommandAlias[0], optional, cmd.Help)
	}
	if len(adminCommands) > 0 && bc.Repo.UserIsAdmin(c.Message()) {
		helpMsg += "\n\n** ADMIN COMMANDS **"
		for _, cmd := range adminCommands {
			helpMsg += fmt.Sprintf("\n/%s - %s", cmd.CommandAlias[0], cmd.Help)
		}
	}
	bc.Bot.SendSilent(bc.Logf, Recipient(c.Message()), helpMsg, clearKeyboard())
	return nil
}

func (bc *BotController) commandCancel(c tb.Context) error {
	tx := bc.State.GetType(c.Message())
	hasState := tx != ST_NONE
	bc.Logf(TRACE, c.Message(), "Clearing state. Had state? %t > '%s'", hasState, tx)

	bc.State.Clear(c.Message())

	msg := "You did not currently have any state or transaction open that could be cancelled."
	if hasState {
		if tx == ST_TPL {
			msg = "Your currently running template creation has been cancelled."
		} else {
			msg = "Your currently running transaction has been cancelled."
		}
	}
	bc.Bot.SendSilent(bc.Logf, Recipient(c.Message()), fmt.Sprintf("%s\nType /%s to get available commands.", msg, CMD_HELP), clearKeyboard())
	return nil
}

const MSG_UNFINISHED_STATE = "You have an unfinished operation running. Please finish it or /cancel it before starting a new one."

type Sender struct {
	recipient string
}

func (s Sender) Recipient() string {
	return s.recipient
}

func Recipient(m *tb.Message) tb.Recipient {
	return &Sender{recipient: fmt.Sprintf("%d", m.Chat.ID)}
}

func (bc *BotController) commandCreateSimpleTx(c tb.Context) error {
	state := bc.State.GetType(c.Message())
	if state != ST_NONE {
		bc.Bot.SendSilent(bc.Logf, Recipient(c.Message()), MSG_UNFINISHED_STATE)
		return nil
	}
	bc.Logf(TRACE, c.Message(), "Creating simple transaction")
	bc.Bot.SendSilent(bc.Logf, Recipient(c.Message()), "In the following steps we will create a simple transaction. "+
		"I will guide you through.\n\n",
		clearKeyboard(),
	)
	tx, err := bc.State.SimpleTx(c.Message(), bc.Repo.UserGetCurrency(c.Message())) // create new tx
	if err != nil {
		bc.Bot.SendSilent(bc.Logf, Recipient(c.Message()), "Something went wrong creating your transactions ("+err.Error()+"). Please check /help for usage."+
			"\n\nYou can create a simple transaction using this command: /simple [date]\ne.g. /simple 2021-01-24\n"+
			"The date parameter is non-mandatory, if not specified, today's date will be taken."+
			"Alternatively it is also possible to send an amount directly to start a new simple transaction.", clearKeyboard())
		return nil
	}
	if tx.IsDone() {
		bc.finishTransaction(c.Message(), tx)
		return nil
	}
	hint := tx.NextHint(bc.Repo, c.Message())
	bc.sendNextTxHint(hint, c.Message())
	return nil
}

func (bc *BotController) commandAddComment(c tb.Context) error {
	if bc.State.GetType(c.Message()) != ST_NONE {
		bc.Logf(INFO, c.Message(), "commandAddComment while in another transaction")
		bc.Bot.SendSilent(bc.Logf, Recipient(c.Message()), MSG_UNFINISHED_STATE)
		return nil
	}
	base := CMD_COMMENT[0]
	if !strings.HasPrefix(c.Message().Text, "/"+base) {
		base = CMD_COMMENT[1]
	}
	remainingCommand := strings.TrimPrefix(strings.TrimLeft(c.Message().Text, ""), "/"+base)

	// Issue #91: Support unquoted comments
	comment := strings.TrimSpace(remainingCommand)
	if strings.HasPrefix(comment, "\"") && strings.HasSuffix(comment, "\"") {
		comment = strings.Trim(comment, "\"")
	}
	comment = strings.ReplaceAll(comment, "\\\"", "\"")

	err := bc.Repo.RecordTransaction(c.Message().Chat.ID, comment+"\n")
	if err != nil {
		bc.Logf(ERROR, c.Message(), "Something went wrong while recording the comment: "+err.Error())
		bc.Bot.SendSilent(bc.Logf, Recipient(c.Message()), "Something went wrong while recording your comment: "+err.Error(), clearKeyboard())
		return nil
	}
	bc.Bot.SendSilent(bc.Logf, Recipient(c.Message()), "Successfully added the comment to your transaction /list", clearKeyboard())
	return nil
}

func (bc *BotController) commandList(c tb.Context) error {
	bc.Logf(TRACE, c.Message(), "Listing transactions")
	command := strings.Split(c.Message().Text, " ")
	isArchived := false
	isDated := false
	isNumbered := false
	isDeleteCommand := false
	elementNumber := -1
	if len(command) > 1 {
		for _, option := range command[1:] {
			if option == "archived" {
				isArchived = true
				continue
			} else if option == "dated" {
				isDated = true
				continue
			} else if option == "numbered" {
				isNumbered = true
				continue
			} else if option == "rm" {
				isDeleteCommand = true
				continue
			} else {
				var err error
				elementNumber, err = strconv.Atoi(option)
				if err != nil {
					bc.Bot.SendSilent(bc.Logf, Recipient(c.Message()), fmt.Sprintf("The option '%s' could not be recognized. Please try again with '/list', with options added to the end separated by space.", option), clearKeyboard())
					return nil
				}
				continue
			}
		}
	}
	if isDeleteCommand && (isNumbered || isDated || elementNumber <= 0) {
		bc.Bot.SendSilent(bc.Logf, Recipient(c.Message()), "For removing a single element from the list, determine it's number by sending the command '/list numbered' and then removing an entry by sending '/list rm <number>'.", clearKeyboard())
		return nil
	}
	tx, err := bc.Repo.GetTransactions(c.Message(), isArchived)
	if err != nil {
		bc.Bot.SendSilent(bc.Logf, Recipient(c.Message()), "Something went wrong retrieving your transactions: "+err.Error(), clearKeyboard())
		return nil
	}
	if tx == nil {
		bc.Logf(ERROR, c.Message(), "Tx unexpectedly was nil")
		return nil
	}
	if isDeleteCommand {
		var err error
		if elementNumber <= len(tx) {
			elementDbId := tx[elementNumber-1].Id
			_, err = bc.Repo.DeleteTransaction(c.Message(), isArchived, elementDbId)
		} else {
			err = fmt.Errorf("the number you specified was too high. Please use a correct number as seen from '/list [archived] numbered'")
		}
		if err != nil {
			bc.Bot.SendSilent(bc.Logf, Recipient(c.Message()), "Something went wrong while trying to delete a single transaction: "+err.Error(), clearKeyboard())
			return nil
		}
		bc.Bot.SendSilent(bc.Logf, Recipient(c.Message()), "Successfully deleted the list entry specified.", clearKeyboard())
		return nil
	}
	SEP := "\n"
	txList := []string{}
	txEntryNumber := 0
	for _, t := range tx {
		var dateComment string
		txEntryNumber++
		if isDated {
			tzOffset := bc.Repo.UserGetTzOffset(c.Message())
			timezoneOff := time.Duration(tzOffset) * time.Hour
			// 2022-03-30T14:24:50.390084Z
			dateParsed, err := time.Parse("2006-01-02T15:04:05Z", t.Date)
			if err != nil {
				bc.Logf(ERROR, c.Message(), "Parsing time failed: %s", err.Error())
				bc.Logf(WARN, c.Message(), "Turning off dated option!")
				isDated = false
			} else {
				date := dateParsed.Add(timezoneOff).Format(helpers.BEANCOUNT_DATE_FORMAT + " 15:04")
				dateComment = "; recorded on " + date + SEP
			}
		}
		numberPrefix := ""
		if isNumbered {
			numberPrefix = fmt.Sprintf("%d) ", txEntryNumber)
		}
		txMessage := dateComment + numberPrefix + t.Tx
		txList = append(txList, txMessage)
	}
	messageSplits := bc.MergeMessagesHonorSendLimit(txList, "\n")
	if len(messageSplits) == 0 {
		archivedSuggestion := ""
		if !isArchived {
			archivedSuggestion = " archived"
		}
		bc.Bot.SendSilent(bc.Logf, Recipient(c.Message()), fmt.Sprintf("Your transaction list is empty. Create some first. Check /%s for commands to create a transaction."+
			"\nYou might also be looking for%s transactions using '/list%s'.", CMD_HELP, archivedSuggestion, archivedSuggestion), clearKeyboard())
		return nil
	}
	for _, message := range messageSplits {
		bc.Bot.SendSilent(bc.Logf, Recipient(c.Message()), message, clearKeyboard())
	}
	return nil
}

func (bc *BotController) MergeMessagesHonorSendLimit(m []string, sep string) []string {
	messages := []string{}
	currentMessageBlock := ""
	for _, msg := range m {
		if len(currentMessageBlock)+len(msg) >= helpers.TG_MAX_MSG_CHAR_LEN {
			bc.Logf(TRACE, nil, "Listed messages extend max message length. Splitting into multiple messages.")
			messages = append(messages, currentMessageBlock)
			currentMessageBlock = ""
		}
		if currentMessageBlock != "" {
			currentMessageBlock += sep
		}
		currentMessageBlock += msg
	}
	if currentMessageBlock != "" {
		messages = append(messages, currentMessageBlock)
	}
	return messages
}

func (bc *BotController) commandArchiveTransactions(c tb.Context) error {
	bc.Logf(TRACE, c.Message(), "Archiving transactions")
	err := bc.Repo.ArchiveTransactions(c.Message())
	if err != nil {
		bc.Bot.SendSilent(bc.Logf, Recipient(c.Message()), "Something went wrong archiving your transactions: "+err.Error())
		return nil
	}
	bc.Bot.SendSilent(bc.Logf, Recipient(c.Message()), fmt.Sprintf("Archived all transactions. Your /%s is empty again.", CMD_LIST), clearKeyboard())
	return nil
}

func (bc *BotController) commandDeleteTransactions(c tb.Context) error {
	if !(strings.TrimSpace(strings.ToLower(c.Message().Text)) == strings.ToLower("/"+CMD_DELETE_ALL+" YES")) {
		bc.Bot.SendSilent(bc.Logf, Recipient(c.Message()), fmt.Sprintf("Please type '/%s yes' to confirm the deletion of your transactions", CMD_DELETE_ALL))
		return nil
	}
	bc.Logf(TRACE, c.Message(), "Deleting transactions")
	_, err := bc.Repo.DeleteTransactions(c.Message())
	if err != nil {
		bc.Bot.SendSilent(bc.Logf, Recipient(c.Message()), "Something went wrong deleting your transactions: "+err.Error())
		return nil
	}
	bc.Bot.SendSilent(bc.Logf, Recipient(c.Message()), fmt.Sprintf("Permanently deleted all your transactions. Your /%s is empty again.", CMD_LIST), clearKeyboard())
	return nil
}

func (bc *BotController) commandTemplates(c tb.Context) error {
	bc.templatesHandler(c.Message())
	return nil
}

func (bc *BotController) commandSuggestions(c tb.Context) error {
	bc.suggestionsHandler(c.Message())
	return nil
}

func (bc *BotController) commandConfig(c tb.Context) error {
	bc.configHandler(c.Message())
	return nil
}

func (bc *BotController) cronInfo() string {
	message := "Job overview:"
	for i, j := range bc.CronScheduler.Jobs() {
		var err string
		if j.Error() != nil {
			err = j.Error().Error()
		}
		message += fmt.Sprintf("\n  %d - running: %t, error: '%s', lastRun: %v, nextRun: %v, runCount: %d",
			i, j.IsRunning(), err, j.LastRun(), j.NextRun(), j.RunCount())
	}
	return fmt.Sprintf(message+"\n  Current timestamp: %v (hour: %d)", time.Now(), time.Now().Hour())
}

func (bc *BotController) cronNotifications() {
	bc.Logf(INFO, nil, "Running notifications job.")
	rows, err := bc.Repo.GetUsersToNotify()
	if err != nil {
		bc.Logf(ERROR, nil, "Error getting users to notify: %s", err.Error())
	}
	defer rows.Close()

	var (
		tgChatId  string
		openCount int
		overdue   int
	)
	for rows.Next() {
		err = rows.Scan(&tgChatId, &overdue, &openCount)
		if err != nil {
			bc.Logf(ERROR, nil, "Error occurred extracting tgChatId to send open tx notification to: %s", err.Error())
			continue
		}
		bc.Logf(TRACE, nil, "Sending notification for %d open transaction(s) to %s", openCount, tgChatId)
		s := "s"
		if openCount == 1 {
			s = ""
		}
		bc.Bot.SendSilent(bc.Logf, ReceiverImpl{ChatId: tgChatId}, fmt.Sprintf(
			// TODO: Replace hard-coded command directives:
			" This is your reminder to inform you that you currently have %d open transaction%s (%d triggering this notification). Check '/list' to see your open transactions. If you don't need them anymore you can /archiveAll or /delete them."+
				"\n\nYou are getting this message because you enabled reminder notifications for open transactions in /config.", openCount, s, overdue))
	}

	bc.Logf(TRACE, nil, bc.cronInfo())
}

type ReceiverImpl struct {
	ChatId string
}

func (r ReceiverImpl) Recipient() string {
	return r.ChatId
}

func (bc *BotController) commandAdminCronInfo(c tb.Context) error {
	isAdmin := bc.Repo.UserIsAdmin(c.Message())
	if !isAdmin {
		bc.Logf(WARN, c.Message(), "Received admin command from non-admin user. Ignoring (treating as normal text input).")
		bc.handleTextState(c)
		return nil
	}
	bc.Bot.SendSilent(bc.Logf, Recipient(c.Message()), bc.cronInfo())
	return nil
}

func (bc *BotController) commandAdminNofify(c tb.Context) error {
	isAdmin := bc.Repo.UserIsAdmin(c.Message())
	if !isAdmin {
		bc.Logf(WARN, c.Message(), "Received admin command from non-admin user. Ignoring (treating as normal text input).")
		return bc.handleTextState(c)
	}
	text := strings.Split(c.Message().Text, "\"")
	var notificationMessage string
	if len(text) >= 2 {
		notificationMessage = text[1]
	}
	if len(text) == 0 || len(notificationMessage) == 0 {
		bc.Bot.SendSilent(bc.Logf, Recipient(c.Message()), "Something went wrong splitting your command parameters. Did you specify a text in double quotes (\")?")
		return nil
	}
	// text[0] = /command [chatId]
	command := strings.Split(strings.TrimRight(text[0], " "), " ")

	if len(command) == 0 || len(command) >= 3 {
		// invalid argument count
		bc.Bot.SendSilent(bc.Logf, Recipient(c.Message()), "Please check the command syntax")
		return nil
	}

	var target string
	if len(command) == 2 {
		target = command[1]
	}

	receivers := bc.Repo.IndividualsWithNotifications(target)
	if len(receivers) == 0 {
		bc.Bot.SendSilent(bc.Logf, Recipient(c.Message()), "No receivers found to send notification to (you being excluded).")
		return nil
	}

	for _, recipient := range receivers {
		bc.Bot.SendSilent(bc.Logf, ReceiverImpl{ChatId: recipient}, "*** Service notification ***\n\n"+notificationMessage)
		bc.Logf(TRACE, c.Message(), "Sent notification to %s", recipient)
		// TODO: Add message like 'If you don't want to receive further service notifications, you can turn them off in the /settings with '/settings notif off'.'
		//  GitHub-issue: #28
	}
	return nil
}

func (bc *BotController) matchesCommandWithoutLeadingSlash(m tb.Context) tb.HandlerFunc {
	// Check if feature has been enabled by user
	exists, value, err := bc.Repo.GetUserSetting(helpers.USERSET_OMITCMDSLASH, m.Message().Chat.ID)
	if err != nil || !exists || strings.ToUpper(value) != "TRUE" {
		bc.Logf(INFO, m.Message(), "Not trying to match command without leading slash as user has not enabled this feature. - Exists: %t - Value: %s - Err?: %v", exists, value, err)
		return nil
	}

	potentialCommandSplits := strings.Split(m.Message().Text, " ")
	potentialCommandSplit := ""
	if len(potentialCommandSplits) >= 1 {
		potentialCommandSplit = strings.TrimSpace(potentialCommandSplits[0])
	}
	if potentialCommandSplit == "" {
		return nil
	}
	for _, mapping := range bc.commandMappings() {
		for _, command := range mapping.CommandAlias {
			if command == potentialCommandSplit {
				return mapping.Handler
			}
		}
	}
	return nil
}

func (bc *BotController) handleTextState(c tb.Context) error {
	state := bc.State.GetType(c.Message())
	if state == ST_NONE {
		if _, err := HandleFloat(c.Message()); err == nil { // Not in tx, but input would suffice for correct parsing of amount field of new tx
			bc.Logf(DEBUG, c.Message(), "Creating new simple transaction as amount has been entered though not in tx")
			_, err = bc.State.SimpleTx(c.Message(), bc.Repo.UserGetCurrency(c.Message())) // create new tx
			if err != nil {
				bc.Bot.SendSilent(bc.Logf, Recipient(c.Message()), "Something went wrong creating a new transaction: "+err.Error(), clearKeyboard())
				return nil
			}
			bc.Bot.SendSilent(bc.Logf, Recipient(c.Message()), "Automatically created a new transaction for you. If you think this was a mistake you can /cancel it.", clearKeyboard())
			bc.handleTextState(c)
			return nil
		} else if handlerFunc := bc.matchesCommandWithoutLeadingSlash(c); handlerFunc != nil {
			bc.Logf(TRACE, c.Message(), "matched command handler without leading slash for message: %s", c.Message().Text)
			c.Message().Text = "/" + c.Message().Text
			handlerFunc(c)
			return nil
		}

		if crud.IsGroupChat(c.Message()) && !strings.HasPrefix(c.Message().Text, "/") {
			bc.Logf(DEBUG, c.Message(), "Received text without having any prior state but am in a group chat. Ignoring and not sending error to user")
			return nil
		}

		bc.Logf(WARN, c.Message(), "Received text without having any prior state and not in group chat or message starts with '/'")
		bc.Bot.SendSilent(bc.Logf, Recipient(c.Message()), fmt.Sprintf("Please check /%s on how to use this bot. E.g. you might need to start a transaction first before sending data.", CMD_HELP), clearKeyboard())
		return nil
	} else if state == ST_TX {
		tx := bc.State.GetTx(c.Message())
		_, err := tx.Input(c.Message())
		if err != nil {
			bc.Logf(WARN, c.Message(), "Invalid text state input: '%s'. Err: %s", c.Message().Text, err.Error())
			bc.Bot.SendSilent(bc.Logf, Recipient(c.Message()), "Your last input seems to have not worked.\n"+
				fmt.Sprintf("(Error: %s)\n", err.Error())+
				"Please try again.",
			)
		}
		bc.Logf(TRACE, c.Message(), "New data state is %v. (Last input was '%s')", tx.Debug(), c.Message().Text)
		if tx.IsDone() {
			bc.finishTransaction(c.Message(), tx)
			return nil
		}
		hint := tx.NextHint(bc.Repo, c.Message())
		bc.sendNextTxHint(hint, c.Message())
		return nil
	} else if state == ST_TPL {
		if bc.processNewTemplateResponse(c.Message(), bc.State.tplStates[chatId(c.Message().Chat.ID)]) {
			bc.State.Clear(c.Message())
		}
		return nil
	}
	bc.Logf(ERROR, c.Message(), "Something went wrong processing text input. Ran to end, though should have been caught by a branch. "+
		"Are there new state types not maintained yet?")
	return nil
}

func (bc *BotController) sendNextTxHint(hint *Hint, m *tb.Message) {
	replyKeyboard := ReplyKeyboard(hint.KeyboardOptions)
	bc.Logf(TRACE, m, "Sending hints for next step: %v", hint.KeyboardOptions)
	bc.Bot.SendSilent(bc.Logf, Recipient(m), escapeCharacters(hint.Prompt, "(", ")", ".", "!"), replyKeyboard, tb.ModeMarkdownV2)
}

func clearKeyboard() *tb.ReplyMarkup {
	return &tb.ReplyMarkup{RemoveKeyboard: true}
}

func (bc *BotController) finishTransaction(m *tb.Message, tx Tx) {
	currency := bc.Repo.UserGetCurrency(m)
	tag := bc.Repo.UserGetTag(m)
	tzOffset := bc.Repo.UserGetTzOffset(m)
	transaction, err := tx.FillTemplate(currency, tag, tzOffset)
	if err != nil {
		bc.Logf(ERROR, m, "Something went wrong while templating the transaction: "+err.Error())
		bc.Bot.SendSilent(bc.Logf, Recipient(m), "Something went wrong while templating the transaction: "+err.Error(), clearKeyboard())
		return
	}

	err = bc.Repo.RecordTransaction(m.Chat.ID, transaction)
	if err != nil {
		bc.Logf(ERROR, m, "Something went wrong while recording the transaction: "+err.Error())
		bc.Bot.SendSilent(bc.Logf, Recipient(m), "Something went wrong while recording your transaction: "+err.Error(), clearKeyboard())
		return
	}

	// TODO: Goroutine
	err = bc.Repo.PutCacheHints(m, tx.CacheData())
	if err != nil {
		bc.Logf(ERROR, m, "Something went wrong while caching transaction. Error: %s", err.Error())
		// Don't return, instead continue flow (if recording was successful)
	}

	bc.Bot.SendSilent(bc.Logf, Recipient(m), fmt.Sprintf("Successfully recorded your transaction.\n"+
		"You can get a list of all your transactions using /%s. "+
		"With /%s you can delete all of them (e.g. once you copied them into your bookkeeping)."+
		"\n\nYou can start a new transaction with /%s or type /%s to see all commands available.",
		CMD_LIST, CMD_ARCHIVE_ALL, CMD_SIMPLE, CMD_HELP),
		clearKeyboard(),
	)

	bc.State.Clear(m)
}
