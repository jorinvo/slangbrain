package main

// Thanks for help with unicode http://apps.timwhitlock.info/emoji/tables/unicode

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"slangbrain-telegram/brain"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	tg "gopkg.in/telegram-bot-api.v4"
)

var verboseFlag = flag.Bool("verbose", false, "print progress information")

var (
	startCmd = "/start"
	helpCmd  = "/help"
	addCmd   = "/add"
)

// Emoji :pencil2:
var addButton = addCmd + " \xE2\x9C\x8F"

var (
	// Markdown formatted
	helpReply = `Hello %s!
I help you study your own way.
You teach me some facts.
I help you to remember them.

Start by adding a new fact using the __/add__ button below.

After you added some facts type  _@slangbrainbot ..._ to search them.`

	// Emoji :memo:
	addReply = "Send me a new fact \xF0\x9F\x93\x9D"

	// Emoji :scream:
	errReply = "Something went wrong, try again \xF0\x9F\x98\xB1"

	// Emoji :+1:
	addDoneReply = "Fact added \xF0\x9F\x91\x8D"

	// Markdown formatted
	inlineQueryReply = `_%s_

%s`
)

func main() {
	flag.Parse()
	token := os.Getenv("BOT_TOKEN")
	if len(token) < 1 {
		log.Panic("Env var BOT_TOKEN not set. Get a token from Botfather.")
	}
	dbFile := os.Getenv("DB_FILE")
	if len(dbFile) < 1 {
		log.Panic("Env var DB_FILE not set. Specify the path for your sqlite database file.")
	}

	bot, err := tg.NewBotAPI(token)
	if err != nil {
		log.Panic(errors.Wrap(err, "failed to connect to Telegram bot API"))
	}

	store, err := brain.CreateStore("sqlite3", dbFile)
	if err != nil {
		log.Panic(err)
	}

	verbose("Authorized on account", bot.Self.UserName)

	u := tg.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		handleMessage(store, bot, update.Message)
		handleInlineQuery(store, bot, update.InlineQuery)
	}
}

func handleMessage(store brain.Store, bot *tg.BotAPI, msg *tg.Message) {
	if msg == nil {
		return
	}

	userID := msg.From.ID
	chatID := msg.Chat.ID
	chatTitle := msg.Chat.Title
	if chatTitle == "" {
		chatTitle = msg.Chat.UserName
	}

	verbose("[m]", msg.From.UserName, "-", msg.Text)

	var reply tg.MessageConfig

	err := store.AddChat(userID, chatID, chatTitle)
	if err != nil {
		log.Println(errors.Wrap(err, "failed to save chat"))
		reply = addKeyboard(tg.NewMessage(chatID, errReply))
		send(bot, reply)
		return
	}

	switch {
	// `/start` or `/help`
	case strings.HasPrefix(msg.Text, startCmd) || strings.HasPrefix(msg.Text, helpCmd):
		reply = addKeyboard(tg.NewMessage(chatID, fmt.Sprintf(helpReply, msg.From.FirstName)))
		reply.ParseMode = "Markdown"
	// `/add`
	case strings.HasPrefix(msg.Text, addCmd):
		reply = forceReply(tg.NewMessage(chatID, addReply))
	// `/add` reply
	case msg.ReplyToMessage != nil && msg.ReplyToMessage.Text == addReply:
		err := store.AddFact(chatID, msg.Text)
		if err != nil {
			log.Println(err)
			reply = addKeyboard(tg.NewMessage(chatID, errReply))
		} else {
			reply = addKeyboard(tg.NewMessage(chatID, addDoneReply))
			reply.ReplyToMessageID = msg.MessageID
		}
	}

	// Ignore unhandled messages
	if reply.Text == "" {
		return
	}

	send(bot, reply)

}

func handleInlineQuery(store brain.Store, bot *tg.BotAPI, q *tg.InlineQuery) {
	if q == nil {
		return
	}

	verbose("[q]", q.From.UserName, "-", q.Query)

	userID := q.From.ID
	var results []interface{}
	facts, err := store.FindFacts(userID, q.Query)
	if err != nil {
		log.Println(err)
	}

	// search in all brains of user
	for i, fact := range facts {
		title := fmt.Sprintf("%v (%s) %s", i+1, fact.ChatTitle, firstChars(10, fact.Content))
		msg := fmt.Sprintf(inlineQueryReply, q.Query, fact.ChatTitle, fact.Content)
		result := tg.NewInlineQueryResultArticleMarkdown(strconv.Itoa(fact.ID), title, msg)
		result.Description = fact.Content
		results = append(results, result)
	}

	inlineConf := tg.InlineConfig{
		InlineQueryID: q.ID,
		IsPersonal:    true,
		Results:       results,
	}

	if _, err := bot.AnswerInlineQuery(inlineConf); err != nil {
		log.Println(err)
	}
}

func send(bot *tg.BotAPI, reply tg.MessageConfig) {
	_, err := bot.Send(reply)
	if err != nil {
		log.Println(errors.Wrap(err, "failed sending reply"))
	}
}

func addKeyboard(msg tg.MessageConfig) tg.MessageConfig {
	keyboard := tg.NewReplyKeyboard(tg.NewKeyboardButtonRow(tg.NewKeyboardButton(addButton)))
	keyboard.OneTimeKeyboard = true
	msg.ReplyMarkup = keyboard
	return msg
}

func forceReply(msg tg.MessageConfig) tg.MessageConfig {
	msg.ReplyMarkup = tg.ForceReply{ForceReply: true}
	return msg
}

// TODO: don't break runes
func firstChars(i int, s string) string {
	if i > len(s) {
		return s[0:len(s)]
	}
	return s[0:i] + "..."
}

func verbose(info ...interface{}) {
	if *verboseFlag {
		log.Println(info...)
	}
}
