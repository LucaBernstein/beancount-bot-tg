package bot

import (
	"fmt"
	"log"
	"strings"

	dbWrapper "github.com/LucaBernstein/beancount-bot-tg/db"
	"github.com/LucaBernstein/beancount-bot-tg/db/crud"
	tb "gopkg.in/tucnak/telebot.v2"
)

type CMD struct {
	Command  string
	Optional string
	Handler  func(m *tb.Message)
	Help     string
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
}

func (bc *BotController) ConfigureAndAttachBot(b IBot) *BotController {
	bc.Bot = b

	mappings := bc.commandMappings()

	for _, m := range mappings {
		b.Handle("/"+m.Command, m.Handler)
	}

	b.Handle(tb.OnText, bc.handleTextState)

	log.Printf("Starting bot '%s'", b.Me().Username)
	b.Start()

	return bc
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
)

func (bc *BotController) commandMappings() []*CMD {
	return []*CMD{
		{Command: CMD_HELP, Handler: bc.commandHelp, Help: "List this command help"},
		{Command: CMD_START, Handler: bc.commandStart, Help: "Give introduction into this bot"},
		{Command: CMD_CANCEL, Handler: bc.commandCancel, Help: "Cancel any running commands or transactions"},
		{Command: CMD_SIMPLE, Handler: bc.commandCreateSimpleTx, Help: "Record a simple transaction, defaults to today", Optional: "YYYY-MM-DD"},
		{Command: CMD_LIST, Handler: bc.commandList, Help: "List your recorded transactions", Optional: "archived"},
		{Command: CMD_SUGGEST, Handler: bc.commandSuggestions, Help: "List, add or remove suggestions"},
		{Command: CMD_CONFIG, Handler: bc.commandConfig, Help: "Bot configurations"},
		{Command: CMD_ARCHIVE_ALL, Handler: bc.commandArchiveTransactions, Help: "Archive recorded transactions"},
		{Command: CMD_DELETE_ALL, Handler: bc.commandDeleteTransactions, Help: "Permanently delete recorded transactions"},

		{Command: CMD_ADM_NOTIFY, Handler: bc.commandAdminNofify, Help: "Send notification to user(s): /" + CMD_ADM_NOTIFY + " [chatId] \"<message>\""},
	}
}

func (bc *BotController) commandStart(m *tb.Message) {
	log.Printf("Received start command from %s (ChatID: %d)", m.Chat.Username, m.Chat.ID)
	bc.Bot.Send(m.Sender, "Welcome to this beancount bot!\n"+
		"You can find more information in the repository under "+
		"https://github.com/LucaBernstein/beancount-bot-tg\n\n"+
		"Please check the commands I will send to you next that are available to you. "+
		"You can always reach the command help by typing /"+CMD_HELP, clearKeyboard())
	bc.commandHelp(m)
}

func (bc *BotController) commandHelp(m *tb.Message) {
	log.Printf("Sending help to %s (ChatID: %d)", m.Chat.Username, m.Chat.ID)
	helpMsg := ""
	adminCommands := []*CMD{}
	for i, cmd := range bc.commandMappings() {
		if cmd.Help == "" {
			continue
		}
		if strings.HasPrefix(cmd.Command, "admin") {
			adminCommands = append(adminCommands, cmd)
			continue
		}
		if i != 0 {
			helpMsg += "\n"
		}
		var optional string
		if cmd.Optional != "" {
			optional = " [" + cmd.Optional + "]"
		}
		helpMsg += fmt.Sprintf("/%s%s - %s", cmd.Command, optional, cmd.Help)
	}
	if len(adminCommands) > 0 && bc.Repo.UserIsAdmin(m) {
		helpMsg += "\n\n** ADMIN COMMANDS **"
		for _, cmd := range adminCommands {
			helpMsg += fmt.Sprintf("\n/%s - %s", cmd.Command, cmd.Help)
		}
	}
	bc.Bot.Send(m.Sender, helpMsg, clearKeyboard())
}

func (bc *BotController) commandCancel(m *tb.Message) {
	tx := bc.State.Get(m)
	isInTx := tx != nil
	log.Printf("Clearing state for %s (ChatID: %d). Was in tx? %t", m.Chat.Username, m.Chat.ID, isInTx)

	bc.State.Clear(m)

	msg := "There were no active transactions open to cancel."
	if isInTx {
		msg = "Your currently running transaction has been cancelled."
	}
	bc.Bot.Send(m.Sender, fmt.Sprintf("%s\nType /%s to see available commands or type /%s to start a new simple transaction.", msg, CMD_HELP, CMD_SIMPLE), clearKeyboard())
}

func (bc *BotController) commandCreateSimpleTx(m *tb.Message) {
	tx := bc.State.Get(m)
	if tx != nil {
		bc.Bot.Send(m.Sender, "You are already in a transaction. Please finish it or /cancel it before starting a new one.", clearKeyboard())
		return
	}
	log.Printf("Creating simple transaction for %s (ChatID: %d)", m.Chat.Username, m.Chat.ID)
	bc.Bot.Send(m.Sender, "In the following steps we will create a simple transaction. "+
		"I will guide you through.\n\n",
		clearKeyboard(),
	)
	tx, err := bc.State.SimpleTx(m) // create new tx
	if err != nil {
		bc.Bot.Send(m.Sender, "Something went wrong creating your transactions ("+err.Error()+"). Please check /help for usage."+
			"\n\nYou can create a simple transaction using this command: /simple [YYYY-MM-DD]\ne.g. /simple 2021-01-24\n"+
			"The date parameter is non-mandatory, if not specified, today's date will be taken.", clearKeyboard())
		return
	}
	hint := tx.NextHint(bc.Repo, m) // get first hint
	bc.Bot.Send(m.Sender, hint.Prompt, ReplyKeyboard(hint.KeyboardOptions))
}

func (bc *BotController) commandList(m *tb.Message) {
	command := strings.Split(m.Text, " ")
	if len(command) == 2 && command[1] != "archived" {
		bc.Bot.Send(m.Sender, fmt.Sprintf("Your parameter after %s could not be recognized. Please try again with '/list' or '/list archived'.", command[0]), clearKeyboard())
		return
	}
	isArchived := len(command) == 2 && command[1] == "archived"
	tx, err := bc.Repo.GetTransactions(m.Chat.ID, isArchived)
	if err != nil {
		bc.Bot.Send(m.Sender, "Something went wrong retrieving your transactions: "+err.Error(), clearKeyboard())
		return
	}
	if tx == "" {
		bc.Bot.Send(m.Sender, fmt.Sprintf("Your transaction list is empty. Create some first. Check /%s for commands to create a transaction."+
			"\nYou might also be looking for archived transactions using '/list archived'.", CMD_HELP), clearKeyboard())
		return
	}
	bc.Bot.Send(m.Sender, tx, clearKeyboard())
}

func (bc *BotController) commandArchiveTransactions(m *tb.Message) {
	err := bc.Repo.ArchiveTransactions(m.Chat.ID)
	if err != nil {
		bc.Bot.Send(m.Sender, "Something went wrong archiving your transactions: "+err.Error())
		return
	}
	bc.Bot.Send(m.Sender, fmt.Sprintf("Archived all transactions. Your /%s is empty again.", CMD_LIST), clearKeyboard())
}

func (bc *BotController) commandDeleteTransactions(m *tb.Message) {
	if !(strings.TrimSpace(strings.ToLower(m.Text)) == strings.ToLower("/"+CMD_DELETE_ALL+" YES")) {
		bc.Bot.Send(m.Sender, fmt.Sprintf("Please type '/%s yes' to confirm the deletion of your transactions", CMD_DELETE_ALL))
		return
	}
	err := bc.Repo.DeleteTransactions(m.Chat.ID)
	if err != nil {
		bc.Bot.Send(m.Sender, "Something went wrong deleting your transactions: "+err.Error())
		return
	}
	bc.Bot.Send(m.Sender, fmt.Sprintf("Permanently deleted all your transactions. Your /%s is empty again.", CMD_LIST), clearKeyboard())
}

func (bc *BotController) commandSuggestions(m *tb.Message) {
	bc.suggestionsHandler(m)
}

func (bc *BotController) commandConfig(m *tb.Message) {
	bc.configHandler(m)
}

type ReceiverImpl struct {
	chatId string
}

func (r ReceiverImpl) Recipient() string {
	return r.chatId
}

func (bc *BotController) commandAdminNofify(m *tb.Message) {
	isAdmin := bc.Repo.UserIsAdmin(m)
	if !isAdmin {
		log.Printf("Received admin command from non-admin user (%s, %d). Ignoring (treating as normal text input).", m.Chat.Username, m.Chat.ID)
		bc.handleTextState(m)
		return
	}
	text := strings.Split(m.Text, "\"")
	var notificationMessage string
	if len(text) >= 2 {
		notificationMessage = text[1]
	}
	if len(text) == 0 || len(notificationMessage) == 0 {
		bc.Bot.Send(m.Sender, "Something went wrong splitting your command parameters. Did you specify a text in double quotes (\")?")
		return
	}
	// text[0] = /command [chatId]
	command := strings.Split(strings.TrimRight(text[0], " "), " ")

	if len(command) == 0 || len(command) >= 3 {
		// invalid argument count
		bc.Bot.Send(m.Sender, "Please check the command syntax")
		return
	}

	var target string
	if len(command) == 2 {
		target = command[1]
	}

	receivers := bc.Repo.IndividualsWithNotifications(m.Chat.ID, target)
	if len(receivers) == 0 {
		bc.Bot.Send(m.Sender, "No receivers found to send notification to (you being excluded).")
		return
	}

	for _, recipient := range receivers {
		bc.Bot.Send(ReceiverImpl{chatId: recipient}, "*** Service notification ***\n\n"+notificationMessage)
		log.Printf("Sent notification to %s", recipient)
		// TODO: Add message like 'If you don't want to receive further service notifications, you can turn them off in the /settings with '/settings notif off'.'
		//  GitHub-issue: #28
	}
}

func (bc *BotController) handleTextState(m *tb.Message) {
	tx := bc.State.Get(m)
	if tx == nil {
		log.Printf("Received text without having any prior state from %s (ChatID: %d)", m.Chat.Username, m.Chat.ID)
		bc.Bot.Send(m.Sender, fmt.Sprintf("Please check /%s on how to use this bot. E.g. you might need to start a transaction first before sending data.", CMD_HELP), clearKeyboard())
		return
	}
	err := tx.Input(m)
	if err != nil {
		bc.Bot.Send(m.Sender, "Your last input seems to have not worked.\n"+
			fmt.Sprintf("(Error: %s)\n", err.Error())+
			"Please try again.",
		)
	}
	log.Printf("New data state for %s (ChatID: %d) is %v. (Last input was '%s')", m.Chat.Username, m.Chat.ID, tx.Debug(), m.Text)
	if tx.IsDone() {
		currency := bc.Repo.UserGetCurrency(m)
		transaction, err := tx.FillTemplate(currency)
		if err != nil {
			log.Printf("Something went wrong while templating the transaction: " + err.Error())
			bc.Bot.Send(m.Sender, "Something went wrong while templating the transaction: "+err.Error(), clearKeyboard())
			return
		}

		err = bc.Repo.RecordTransaction(m.Chat.ID, transaction)
		if err != nil {
			log.Printf("Something went wrong while recording your transaction: " + err.Error())
			bc.Bot.Send(m.Sender, "Something went wrong while recording your transaction: "+err.Error(), clearKeyboard())
			return
		}

		// TODO: Goroutine
		err = bc.Repo.PutCacheHints(m, tx.DataKeys())
		if err != nil {
			log.Printf("Something went wrong while caching transaction. Error: %s", err.Error())
			// Don't return, instead continue flow (if recording was successful)
		}

		bc.Bot.Send(m.Sender, fmt.Sprintf("Successfully recorded your transaction.\n"+
			"You can get a list of all your transactions using /%s. "+
			"With /%s you can delete all of them (e.g. once you copied them into your bookkeeping)."+
			"\n\nYou can start a new transaction with /%s or type /%s to see all commands available.",
			CMD_LIST, CMD_ARCHIVE_ALL, CMD_SIMPLE, CMD_HELP),
			clearKeyboard(),
		)

		bc.State.Clear(m)
		return
	}
	hint := tx.NextHint(bc.Repo, m)
	replyKeyboard := ReplyKeyboard(hint.KeyboardOptions)
	log.Printf("Sending hints for next step: %v", hint.KeyboardOptions)
	bc.Bot.Send(m.Sender, hint.Prompt, replyKeyboard)
}

func clearKeyboard() *tb.ReplyMarkup {
	return &tb.ReplyMarkup{ReplyKeyboardRemove: true}
}
