package bot

import (
	"fmt"
	"strings"
	"time"

	dbWrapper "github.com/LucaBernstein/beancount-bot-tg/db"
	"github.com/LucaBernstein/beancount-bot-tg/db/crud"
	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	"github.com/go-co-op/gocron"
	tb "gopkg.in/tucnak/telebot.v2"
)

type CMD struct {
	CommandAlias []string
	Optional     []string
	Handler      func(m *tb.Message)
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

	// Todo: Add generic callback handler
	// Route callback by ID splits
	bc.Bot.Handle(&btnSuggListAccFrom, func(c *tb.Callback) {
		bc.Logf(DEBUG, nil, "Handling callback on button. Chat: %d", c.Message.Chat.ID)
		c.Message.Text = "/suggestions list accFrom"
		// TODO: What happens in group chats?
		c.Message.Sender = &tb.User{ID: c.Message.Chat.ID} // hack to send chat user a message (in private chats userId = chatId)
		bc.suggestionsHandler(c.Message)
		bc.Bot.Respond(c, &tb.CallbackResponse{}) // Always respond
	})
	bc.Bot.Handle(&btnSuggListAccTo, func(c *tb.Callback) {
		bc.Logf(DEBUG, nil, "Handling callback on button. Chat: %d", c.Message.Chat.ID)
		c.Message.Text = "/suggestions list accTo"
		c.Message.Sender = &tb.User{ID: c.Message.Chat.ID}
		bc.suggestionsHandler(c.Message)
		bc.Bot.Respond(c, &tb.CallbackResponse{}) // Always respond
	})
	bc.Bot.Handle(&btnSuggListTxDesc, func(c *tb.Callback) {
		bc.Logf(DEBUG, nil, "Handling callback on button. Chat: %d", c.Message.Chat.ID)
		c.Message.Text = "/suggestions list txDesc"
		c.Message.Sender = &tb.User{ID: c.Message.Chat.ID}
		bc.suggestionsHandler(c.Message)
		bc.Bot.Respond(c, &tb.CallbackResponse{}) // Always respond
	})

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
	CMD_COMMENT     = "comment"
	CMD_LIST        = "list"
	CMD_ARCHIVE_ALL = "archiveAll"
	CMD_DELETE_ALL  = "deleteAll"
	CMD_SUGGEST     = "suggestions"
	CMD_CONFIG      = "config"

	CMD_ADM_NOTIFY = "admin_notify"
	CMD_ADM_CRON   = "admin_cron"
)

var (
	CMD_TEMPLATE = []string{"template", "t"}
)

func (bc *BotController) commandMappings() []*CMD {
	return []*CMD{
		{CommandAlias: []string{CMD_HELP}, Handler: bc.commandHelp, Help: "List this command help"},
		{CommandAlias: []string{CMD_START}, Handler: bc.commandStart, Help: "Give introduction into this bot"},
		{CommandAlias: []string{CMD_CANCEL}, Handler: bc.commandCancel, Help: "Cancel any running commands or transactions"},
		{CommandAlias: []string{CMD_SIMPLE}, Handler: bc.commandCreateSimpleTx, Help: "Record a simple transaction, defaults to today; Can be ommitted by sending amount directy", Optional: []string{"YYYY-MM-DD"}},
		{CommandAlias: []string{CMD_COMMENT}, Handler: bc.commandAddComment, Help: "Add arbitrary text to transaction list"},
		{CommandAlias: CMD_TEMPLATE, Handler: bc.commandTemplates, Help: "Create and use template transactions"},
		{CommandAlias: []string{CMD_LIST}, Handler: bc.commandList, Help: "List your recorded transactions", Optional: []string{"archived", "dated"}},
		{CommandAlias: []string{CMD_SUGGEST}, Handler: bc.commandSuggestions, Help: "List, add or remove suggestions"},
		{CommandAlias: []string{CMD_CONFIG}, Handler: bc.commandConfig, Help: "Bot configurations"},
		{CommandAlias: []string{CMD_ARCHIVE_ALL}, Handler: bc.commandArchiveTransactions, Help: "Archive recorded transactions"},
		{CommandAlias: []string{CMD_DELETE_ALL}, Handler: bc.commandDeleteTransactions, Help: "Permanently delete recorded transactions"},

		{CommandAlias: []string{CMD_ADM_NOTIFY}, Handler: bc.commandAdminNofify, Help: "Send notification to user(s): /" + CMD_ADM_NOTIFY + " [chatId] \"<message>\""},
		{CommandAlias: []string{CMD_ADM_CRON}, Handler: bc.commandAdminCronInfo, Help: "Check cron status"},
	}
}

func (bc *BotController) commandStart(m *tb.Message) {
	bc.Logf(TRACE, m, "Start command")
	_, err := bc.Bot.Send(m.Sender, "Welcome to this beancount bot!\n"+
		"You can find more information in the repository under "+
		"https://github.com/LucaBernstein/beancount-bot-tg\n\n"+
		"Please check the commands I will send to you next that are available to you. "+
		"You can always reach the command help by typing /"+CMD_HELP, clearKeyboard())
	if err != nil {
		bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
	}
	bc.commandHelp(m)
}

func (bc *BotController) commandHelp(m *tb.Message) {
	bc.Logf(TRACE, m, "Sending help")
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
	if len(adminCommands) > 0 && bc.Repo.UserIsAdmin(m) {
		helpMsg += "\n\n** ADMIN COMMANDS **"
		for _, cmd := range adminCommands {
			helpMsg += fmt.Sprintf("\n/%s - %s", cmd.CommandAlias[0], cmd.Help)
		}
	}
	_, err := bc.Bot.Send(m.Sender, helpMsg, clearKeyboard())
	if err != nil {
		bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
	}
}

func (bc *BotController) commandCancel(m *tb.Message) {
	tx := bc.State.GetType(m)
	hasState := tx != ST_NONE
	bc.Logf(TRACE, m, "Clearing state. Had state? %t > '%s'", hasState, tx)

	bc.State.Clear(m)

	msg := "You did not currently have any state or transaction open that could be cancelled."
	if hasState {
		if tx == ST_TPL {
			msg = "Your currently running template creation has been cancelled."
		} else {
			msg = "Your currently running transaction has been cancelled."
		}
	}
	_, err := bc.Bot.Send(m.Sender, fmt.Sprintf("%s\nType /%s to get available commands.", msg, CMD_HELP), clearKeyboard())
	if err != nil {
		bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
	}
}

const MSG_UNFINISHED_STATE = "You have an unfinished operation running. Please finish it or /cancel it before starting a new one."

func (bc *BotController) commandCreateSimpleTx(m *tb.Message) {
	state := bc.State.GetType(m)
	if state != ST_NONE {
		_, err := bc.Bot.Send(m.Sender, MSG_UNFINISHED_STATE)
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}
	bc.Logf(TRACE, m, "Creating simple transaction")
	_, err := bc.Bot.Send(m.Sender, "In the following steps we will create a simple transaction. "+
		"I will guide you through.\n\n",
		clearKeyboard(),
	)
	if err != nil {
		bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
	}
	tx, err := bc.State.SimpleTx(m, bc.Repo.UserGetCurrency(m)) // create new tx
	if err != nil {
		_, err := bc.Bot.Send(m.Sender, "Something went wrong creating your transactions ("+err.Error()+"). Please check /help for usage."+
			"\n\nYou can create a simple transaction using this command: /simple [YYYY-MM-DD]\ne.g. /simple 2021-01-24\n"+
			"The date parameter is non-mandatory, if not specified, today's date will be taken."+
			"Alternatively it is also possible to send an amount directly to start a new simple transaction.", clearKeyboard())
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}
	if tx.IsDone() {
		bc.finishTransaction(m, tx)
		return
	}
	hint := tx.NextHint(bc.Repo, m)
	_, err = bc.Bot.Send(m.Sender, hint.Prompt, ReplyKeyboard(hint.KeyboardOptions))
	if err != nil {
		bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
	}
}

func (bc *BotController) commandAddComment(m *tb.Message) {
	if bc.State.GetType(m) != ST_NONE {
		bc.Logf(INFO, m, "commandAddComment while in another transaction")
		_, err := bc.Bot.Send(m.Sender, MSG_UNFINISHED_STATE)
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}
	remainingCommand := strings.TrimPrefix(strings.TrimLeft(m.Text, ""), "/"+CMD_COMMENT)

	// Issue #91: Support unquoted comments
	comment := strings.TrimSpace(remainingCommand)
	if strings.HasPrefix(comment, "\"") && strings.HasSuffix(comment, "\"") {
		comment = strings.Trim(comment, "\"")
	}
	comment = strings.ReplaceAll(comment, "\\\"", "\"")

	err := bc.Repo.RecordTransaction(m.Chat.ID, comment)
	if err != nil {
		bc.Logf(ERROR, m, "Something went wrong while recording the comment: "+err.Error())
		_, err := bc.Bot.Send(m.Sender, "Something went wrong while recording your comment: "+err.Error(), clearKeyboard())
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}
	_, err = bc.Bot.Send(m.Sender, "Successfully added the comment to your transaction /list", clearKeyboard())
	if err != nil {
		bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
	}
}

func (bc *BotController) commandList(m *tb.Message) {
	bc.Logf(TRACE, m, "Listing transactions")
	command := strings.Split(m.Text, " ")
	isArchived := false
	isDated := false
	if len(command) > 1 {
		for _, option := range command[1:] {
			if option == "archived" {
				isArchived = true
				continue
			} else if option == "dated" {
				isDated = true
				continue
			}

			_, err := bc.Bot.Send(m.Sender, fmt.Sprintf("The option '%s' could not be recognized. Please try again with '/list', with options added to the end separated by space.", option), clearKeyboard())
			if err != nil {
				bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
			}
			return
		}
	}
	tx, err := bc.Repo.GetTransactions(m, isArchived)
	if err != nil {
		_, err := bc.Bot.Send(m.Sender, "Something went wrong retrieving your transactions: "+err.Error(), clearKeyboard())
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}
	if tx == nil {
		bc.Logf(ERROR, m, "Tx unexpectedly was nil")
		return
	}
	SEP := "\n"
	TG_MAX_MSG_CHAR_LEN := 4096
	txMessages := []string{}
	transactionsList := ""
	for _, t := range tx {
		if len(transactionsList)+len(t.Tx) >= TG_MAX_MSG_CHAR_LEN {
			bc.Logf(TRACE, m, "Listed messages extend max message length. Splitting into multiple messages.")
			txMessages = append(txMessages, transactionsList)
			transactionsList = ""
		}
		var dateComment string
		if isDated {
			tzOffset := bc.Repo.UserGetTzOffset(m)
			timezoneOff := time.Duration(tzOffset) * time.Hour
			// 2022-03-30T14:24:50.390084Z
			dateParsed, err := time.Parse("2006-01-02T15:04:05Z", t.Date)
			if err != nil {
				bc.Logf(ERROR, m, "Parsing time failed: %s", err.Error())
				bc.Logf(WARN, m, "Turning off dated option!")
				isDated = false
			} else {
				date := dateParsed.Add(timezoneOff).Format(helpers.BEANCOUNT_DATE_FORMAT + " 15:04")
				dateComment = "; recorded on " + date + SEP
			}
		}
		transactionsList += dateComment + t.Tx + SEP
	}
	if transactionsList != "" {
		txMessages = append(txMessages, transactionsList)
	}
	if len(txMessages) == 0 {
		_, err := bc.Bot.Send(m.Sender, fmt.Sprintf("Your transaction list is empty. Create some first. Check /%s for commands to create a transaction."+
			"\nYou might also be looking for archived transactions using '/list archived'.", CMD_HELP), clearKeyboard())
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}
	for _, message := range txMessages {
		_, err = bc.Bot.Send(m.Sender, message, clearKeyboard())
	}
	if err != nil {
		bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
	}
}

func (bc *BotController) commandArchiveTransactions(m *tb.Message) {
	bc.Logf(TRACE, m, "Archiving transactions")
	err := bc.Repo.ArchiveTransactions(m)
	if err != nil {
		_, err := bc.Bot.Send(m.Sender, "Something went wrong archiving your transactions: "+err.Error())
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}
	_, err = bc.Bot.Send(m.Sender, fmt.Sprintf("Archived all transactions. Your /%s is empty again.", CMD_LIST), clearKeyboard())
	if err != nil {
		bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
	}
}

func (bc *BotController) commandDeleteTransactions(m *tb.Message) {
	if !(strings.TrimSpace(strings.ToLower(m.Text)) == strings.ToLower("/"+CMD_DELETE_ALL+" YES")) {
		_, err := bc.Bot.Send(m.Sender, fmt.Sprintf("Please type '/%s yes' to confirm the deletion of your transactions", CMD_DELETE_ALL))
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}
	bc.Logf(TRACE, m, "Deleting transactions")
	err := bc.Repo.DeleteTransactions(m)
	if err != nil {
		_, err := bc.Bot.Send(m.Sender, "Something went wrong deleting your transactions: "+err.Error())
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}
	_, err = bc.Bot.Send(m.Sender, fmt.Sprintf("Permanently deleted all your transactions. Your /%s is empty again.", CMD_LIST), clearKeyboard())
	if err != nil {
		bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
	}
}

func (bc *BotController) commandTemplates(m *tb.Message) {
	bc.templatesHandler(m)
}

func (bc *BotController) commandSuggestions(m *tb.Message) {
	bc.suggestionsHandler(m)
}

func (bc *BotController) commandConfig(m *tb.Message) {
	bc.configHandler(m)
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
		_, err := bc.Bot.Send(ReceiverImpl{chatId: tgChatId}, fmt.Sprintf(
			// TODO: Replace hard-coded command directives:
			" This is your reminder to inform you that you currently have %d open transaction%s (%d triggering this notification). Check '/list' to see your open transactions. If you don't need them anymore you can /archiveAll or /delete them."+
				"\n\nYou are getting this message because you enabled reminder notifications for open transactions in /config.", openCount, s, overdue))
		if err != nil {
			bc.Logf(ERROR, nil, "Sending bot message failed: %s", err.Error())
		}
	}

	bc.Logf(TRACE, nil, bc.cronInfo())
}

type ReceiverImpl struct {
	chatId string
}

func (r ReceiverImpl) Recipient() string {
	return r.chatId
}

func (bc *BotController) commandAdminCronInfo(m *tb.Message) {
	isAdmin := bc.Repo.UserIsAdmin(m)
	if !isAdmin {
		bc.Logf(WARN, m, "Received admin command from non-admin user. Ignoring (treating as normal text input).")
		bc.handleTextState(m)
		return
	}
	_, err := bc.Bot.Send(m.Sender, bc.cronInfo())
	if err != nil {
		bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
	}
}

func (bc *BotController) commandAdminNofify(m *tb.Message) {
	isAdmin := bc.Repo.UserIsAdmin(m)
	if !isAdmin {
		bc.Logf(WARN, m, "Received admin command from non-admin user. Ignoring (treating as normal text input).")
		bc.handleTextState(m)
		return
	}
	text := strings.Split(m.Text, "\"")
	var notificationMessage string
	if len(text) >= 2 {
		notificationMessage = text[1]
	}
	if len(text) == 0 || len(notificationMessage) == 0 {
		_, err := bc.Bot.Send(m.Sender, "Something went wrong splitting your command parameters. Did you specify a text in double quotes (\")?")
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}
	// text[0] = /command [chatId]
	command := strings.Split(strings.TrimRight(text[0], " "), " ")

	if len(command) == 0 || len(command) >= 3 {
		// invalid argument count
		_, err := bc.Bot.Send(m.Sender, "Please check the command syntax")
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}

	var target string
	if len(command) == 2 {
		target = command[1]
	}

	receivers := bc.Repo.IndividualsWithNotifications(target)
	if len(receivers) == 0 {
		_, err := bc.Bot.Send(m.Sender, "No receivers found to send notification to (you being excluded).")
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}

	for _, recipient := range receivers {
		_, err := bc.Bot.Send(ReceiverImpl{chatId: recipient}, "*** Service notification ***\n\n"+notificationMessage)
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		bc.Logf(TRACE, m, "Sent notification to %s", recipient)
		// TODO: Add message like 'If you don't want to receive further service notifications, you can turn them off in the /settings with '/settings notif off'.'
		//  GitHub-issue: #28
	}
}

func (bc *BotController) handleTextState(m *tb.Message) {
	state := bc.State.GetType(m)
	if state == ST_NONE {
		if _, err := HandleFloat(m); err == nil { // Not in tx, but input would suffice for correct parsing of amount field of new tx
			bc.Logf(DEBUG, m, "Creating new simple transaction as amount has been entered though not in tx")
			_, err = bc.State.SimpleTx(m, bc.Repo.UserGetCurrency(m)) // create new tx
			if err != nil {
				_, err := bc.Bot.Send(m.Sender, "Something went wrong creating a new transaction: "+err.Error(), clearKeyboard())
				if err != nil {
					bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
				}
				return
			}
			_, err := bc.Bot.Send(m.Sender, "Automatically created a new transaction for you. If you think this was a mistake you can /cancel it.", clearKeyboard())
			if err != nil {
				bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
			}
			bc.handleTextState(m)
			return
		}

		// If number has been entered
		// Create new tx, inform user that tx has automatically been started, call handleTextState with same message again (infininite loop protection?)
		// return
		// else: warn

		bc.Logf(WARN, m, "Received text without having any prior state")
		_, err := bc.Bot.Send(m.Sender, fmt.Sprintf("Please check /%s on how to use this bot. E.g. you might need to start a transaction first before sending data.", CMD_HELP), clearKeyboard())
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	} else if state == ST_TX {
		tx := bc.State.GetTx(m)
		err := tx.Input(m)
		if err != nil {
			bc.Logf(WARN, m, "Invalid text state input: '%s'. Err: %s", m.Text, err.Error())
			_, err := bc.Bot.Send(m.Sender, "Your last input seems to have not worked.\n"+
				fmt.Sprintf("(Error: %s)\n", err.Error())+
				"Please try again.",
			)
			if err != nil {
				bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
			}
		}
		bc.Logf(TRACE, m, "New data state is %v. (Last input was '%s')", tx.Debug(), m.Text)
		if tx.IsDone() {
			bc.finishTransaction(m, tx)
			return
		}
		hint := tx.NextHint(bc.Repo, m)
		replyKeyboard := ReplyKeyboard(hint.KeyboardOptions)
		bc.Logf(TRACE, m, "Sending hints for next step: %v", hint.KeyboardOptions)
		_, err = bc.Bot.Send(m.Sender, hint.Prompt, replyKeyboard)
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	} else if state == ST_TPL {
		if bc.processNewTemplateResponse(m, bc.State.tplStates[chatId(m.Chat.ID)]) {
			bc.State.Clear(m)
		}
		return
	}
	bc.Logf(ERROR, m, "Something went wrong processing text input. Ran to end, though should have been caught by a branch. "+
		"Are there new state types not maintained yet?")
}

func clearKeyboard() *tb.ReplyMarkup {
	return &tb.ReplyMarkup{ReplyKeyboardRemove: true}
}

func (bc *BotController) finishTransaction(m *tb.Message, tx Tx) {
	currency := bc.Repo.UserGetCurrency(m)
	tag := bc.Repo.UserGetTag(m)
	tzOffset := bc.Repo.UserGetTzOffset(m)
	transaction, err := tx.FillTemplate(currency, tag, tzOffset)
	if err != nil {
		bc.Logf(ERROR, m, "Something went wrong while templating the transaction: "+err.Error())
		_, err := bc.Bot.Send(m.Sender, "Something went wrong while templating the transaction: "+err.Error(), clearKeyboard())
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}

	err = bc.Repo.RecordTransaction(m.Chat.ID, transaction)
	if err != nil {
		bc.Logf(ERROR, m, "Something went wrong while recording the transaction: "+err.Error())
		_, err := bc.Bot.Send(m.Sender, "Something went wrong while recording your transaction: "+err.Error(), clearKeyboard())
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}

	// TODO: Goroutine
	err = bc.Repo.PutCacheHints(m, tx.DataKeys())
	if err != nil {
		bc.Logf(ERROR, m, "Something went wrong while caching transaction. Error: %s", err.Error())
		// Don't return, instead continue flow (if recording was successful)
	}
	err = bc.Repo.PruneUserCachedSuggestions(m)
	if err != nil {
		bc.Logf(ERROR, m, "Something went wrong while pruning suggestions cache. Error: %s", err.Error())
		// Don't return, instead continue flow (if recording was successful)
	}

	_, err = bc.Bot.Send(m.Sender, fmt.Sprintf("Successfully recorded your transaction.\n"+
		"You can get a list of all your transactions using /%s. "+
		"With /%s you can delete all of them (e.g. once you copied them into your bookkeeping)."+
		"\n\nYou can start a new transaction with /%s or type /%s to see all commands available.",
		CMD_LIST, CMD_ARCHIVE_ALL, CMD_SIMPLE, CMD_HELP),
		clearKeyboard(),
	)
	if err != nil {
		bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
	}

	bc.State.Clear(m)
}
