package main

import (
	"fmt"
	"log"
	"time"

	"github.com/LucaBernstein/beancount-bot-tg/bot"
	"github.com/LucaBernstein/beancount-bot-tg/db"
	"github.com/LucaBernstein/beancount-bot-tg/db/crud"
	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
)

const (
	ENV_TG_BOT_API_KEY = "BOT_API_KEY"
)

var (
	CRUD_REPO *crud.Repo
	STATE     *bot.StateHandler
)

func main() {
	db := db.PostgresConnection()
	defer db.Close()
	CRUD_REPO = crud.NewRepo(db)
	STATE = bot.NewStateHandler()

	botToken := envTgBotToken()
	poller := &tb.LongPoller{Timeout: 20 * time.Second}
	userGuardPoller := tb.NewMiddlewarePoller(poller, func(upd *tb.Update) bool {
		// TODO: Start goroutine to update data?
		CRUD_REPO.EnrichUserData(upd.Message)
		return true
	})

	b, err := tb.NewBot(tb.Settings{
		Token:  botToken,
		Poller: userGuardPoller,
	})
	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle("/start", func(m *tb.Message) {
		commandStart(b, m)
	})

	b.Handle("/help", func(m *tb.Message) {
		commandHelp(b, m)
	})

	b.Handle("/clear", func(m *tb.Message) {
		commandClear(b, m)
	})

	b.Handle("/simple", func(m *tb.Message) {
		commandCreateSimpleTx(b, m)
	})

	b.Handle("/archiveTransactions", func(m *tb.Message) {
		commandArchiveTransactions(b, m)
	})

	b.Handle("/list", func(m *tb.Message) {
		commandList(b, m)
	})

	b.Handle(tb.OnText, func(m *tb.Message) {
		handleTextState(b, m)
	})

	b.Handle(tb.OnQuery, func(q *tb.Query) {
		// incoming inline queries
	})

	b.Handle(tb.OnPhoto, func(m *tb.Message) {
		// photos only
	})

	log.Printf("Starting bot %s", b.Me.Username)

	b.Start()
}

func commandStart(b *tb.Bot, m *tb.Message) {
	log.Printf("Received start command from %s (ChatID: %d)", m.Chat.Username, m.Chat.ID)
	b.Send(m.Sender, "Welcome to this beancount bot!\n"+
		"You can find more information in the repository under "+
		"https://github.com/LucaBernstein/beancount-bot-tg\n\n"+
		"Please check the commands I will send to you next that are available to you. "+
		"You can always reach the command help by typing /help")
	commandHelp(b, m)
}

func commandCreateSimpleTx(b *tb.Bot, m *tb.Message) {
	log.Printf("Creating simple transaction for %s (ChatID: %d)", m.Chat.Username, m.Chat.ID)
	b.Send(m.Sender, "In the following steps we will create a simple transaction. "+
		"I will guide you through.\n\n"+
		"Please enter the amount of money.",
	)
	STATE.SimpleTx(m)
}

func handleTextState(b *tb.Bot, m *tb.Message) {
	tx := STATE.Get(m)
	if tx == nil {
		log.Printf("Received text without having any prior state from %s (ChatID: %d)", m.Chat.Username, m.Chat.ID)
		b.Send(m.Sender, "Please check /help on how to use this bot. E.g. you might need to start a transaction first before sending data.")
		return
	}
	err := tx.Input(m)
	if err != nil {
		b.Send(m.Sender, "Your last input seems to have not worked.\n"+
			fmt.Sprintf("(Error: %s)\n", err.Error())+
			"Please try again.",
		)
	}
	log.Printf("New data state for %s (ChatID: %d) is %v. (Input now was %s)", m.Chat.Username, m.Chat.ID, tx.Debug(), m.Text)
	if tx.IsDone() {
		transaction, err := tx.FillTemplate()
		if err != nil {
			b.Send(m.Sender, "Something went wrong while templating the transaction: "+err.Error())
			return
		}

		err = CRUD_REPO.RecordTransaction(m.Chat.ID, transaction)
		if err != nil {
			b.Send(m.Sender, "Something went wrong while recording your transaction: "+err.Error())
			return
		}

		b.Send(m.Sender, "Successfully recorded your transaction.\n"+
			"You can get a list of all your transactions using /list. "+
			"With /archiveTransactions you can delete all of them (e.g. once you copied them into your bookkeeping)."+
			"\n\nYou can start a new transaction with /simple or type /help to see all commands available.",
		)
		return
	}
	b.Send(m.Sender, (string)(tx.NextHint()))
}

func commandHelp(b *tb.Bot, m *tb.Message) {
	log.Printf("Sending help to %s (ChatID: %d)", m.Chat.Username, m.Chat.ID)
	b.Send(m.Sender,
		"/help - List this command help\n"+
			"/clear - Cancel any running commands\n"+
			"/simple - Record a simple transaction",
	)
}

func commandClear(b *tb.Bot, m *tb.Message) {
	tx := STATE.Get(m)
	isInTx := tx != nil
	log.Printf("Clearing state for %s (ChatID: %d). Was in tx? %t", m.Chat.Username, m.Chat.ID, isInTx)

	STATE.Clear(m)

	msg := "There were no active transactions open to cancel."
	if isInTx {
		msg = "Your currently running transaction has been cancelled."
	}
	b.Send(m.Sender, msg+"\nType /help to see available commands or type /simple to start a new simple transaction.")
}

func commandArchiveTransactions(b *tb.Bot, m *tb.Message) {
	err := CRUD_REPO.ArchiveTransactions(m.Chat.ID)
	if err != nil {
		b.Send(m.Sender, "Something went wrong archiving your transactions: "+err.Error())
		return
	}
	b.Send(m.Sender, "Archived all transactions. Your /list is empty again.")
}

func commandList(b *tb.Bot, m *tb.Message) {
	tx, err := CRUD_REPO.GetTransactions(m.Chat.ID)
	if err != nil {
		b.Send(m.Sender, "Something went wrong retrieving your transactions: "+err.Error())
		return
	}
	if tx == "" {
		b.Send(m.Sender, "Your transaction list is already empty. Create some first. Check /simple or /help for commands.")
		return
	}
	b.Send(m.Sender, tx)
}

func envTgBotToken() string {
	TG_BOT_APIKEY := helpers.Env(ENV_TG_BOT_API_KEY)
	if TG_BOT_APIKEY == "" {
		log.Fatalf("Please provide Telegram bot API key as ENV var '%s'", ENV_TG_BOT_API_KEY)
	}
	return TG_BOT_APIKEY
}
