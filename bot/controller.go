package bot

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/LucaBernstein/beancount-bot-tg/db/crud"
	tb "gopkg.in/tucnak/telebot.v2"
)

type CMD struct {
	Command string
	Handler func(m *tb.Message)
	Help    string
}

func NewBotController(db *sql.DB) *BotController {
	return &BotController{
		Repo:  crud.NewRepo(db),
		State: NewStateHandler(),
	}
}

type BotController struct {
	Repo  *crud.Repo
	State *StateHandler
	Bot   *tb.Bot
}

func (bc *BotController) ConfigureAndAttachBot(b *tb.Bot) *BotController {
	bc.Bot = b

	mappings := bc.commandMappings()

	for _, m := range mappings {
		b.Handle("/"+m.Command, m.Handler)
	}

	b.Handle(tb.OnText, bc.handleTextState)

	log.Printf("Starting bot '%s'", b.Me.Username)
	b.Start()

	return bc
}

func (bc *BotController) commandMappings() []*CMD {
	return []*CMD{
		{Command: "start", Handler: bc.commandStart},
		{Command: "help", Handler: bc.commandHelp, Help: "List this command help"},
		{Command: "cancel", Handler: bc.commandCancel, Help: "Cancel any running commands"},
		{Command: "simple", Handler: bc.commandCreateSimpleTx, Help: "Record a simple transaction"},
		{Command: "list", Handler: bc.commandList, Help: "List your recorded transactions"},
		{Command: "archiveAll", Handler: bc.commandArchiveTransactions, Help: "Archive recorded transactions"},
	}
}

func (bc *BotController) commandStart(m *tb.Message) {
	log.Printf("Received start command from %s (ChatID: %d)", m.Chat.Username, m.Chat.ID)
	bc.Bot.Send(m.Sender, "Welcome to this beancount bot!\n"+
		"You can find more information in the repository under "+
		"https://github.com/LucaBernstein/beancount-bot-tg\n\n"+
		"Please check the commands I will send to you next that are available to you. "+
		"You can always reach the command help by typing /help", clearKeyboard())
	bc.commandHelp(m)
}

func (bc *BotController) commandHelp(m *tb.Message) {
	log.Printf("Sending help to %s (ChatID: %d)", m.Chat.Username, m.Chat.ID)
	helpMsg := ""
	for i, cmd := range bc.commandMappings() {
		if cmd.Help == "" {
			continue
		}
		if i != 0 {
			helpMsg += "\n"
		}
		helpMsg += fmt.Sprintf("/%s - %s", cmd.Command, cmd.Help)
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
	bc.Bot.Send(m.Sender, msg+"\nType /help to see available commands or type /simple to start a new simple transaction.", clearKeyboard())
}

func (bc *BotController) commandCreateSimpleTx(m *tb.Message) {
	log.Printf("Creating simple transaction for %s (ChatID: %d)", m.Chat.Username, m.Chat.ID)
	bc.Bot.Send(m.Sender, "In the following steps we will create a simple transaction. "+
		"I will guide you through.\n\n"+
		"Please enter the amount of money.",
		clearKeyboard(),
	)
	bc.State.SimpleTx(m)
}

func (bc *BotController) commandList(m *tb.Message) {
	tx, err := bc.Repo.GetTransactions(m.Chat.ID)
	if err != nil {
		bc.Bot.Send(m.Sender, "Something went wrong retrieving your transactions: "+err.Error(), clearKeyboard())
		return
	}
	if tx == "" {
		bc.Bot.Send(m.Sender, "Your transaction list is already empty. Create some first. Check /simple or /help for commands.", clearKeyboard())
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
	bc.Bot.Send(m.Sender, "Archived all transactions. Your /list is empty again.", clearKeyboard())
}

func (bc *BotController) handleTextState(m *tb.Message) {
	tx := bc.State.Get(m)
	if tx == nil {
		log.Printf("Received text without having any prior state from %s (ChatID: %d)", m.Chat.Username, m.Chat.ID)
		bc.Bot.Send(m.Sender, "Please check /help on how to use this bot. E.g. you might need to start a transaction first before sending data.", clearKeyboard())
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
		transaction, err := tx.FillTemplate()
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

		bc.Bot.Send(m.Sender, "Successfully recorded your transaction.\n"+
			"You can get a list of all your transactions using /list. "+
			"With /archiveAll you can delete all of them (e.g. once you copied them into your bookkeeping)."+
			"\n\nYou can start a new transaction with /simple or type /help to see all commands available.",
			clearKeyboard(),
		)
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
