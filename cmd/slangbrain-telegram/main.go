package main

// Thanks for help with unicode http://apps.timwhitlock.info/emoji/tables/unicode

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/jorinvo/slangbrain/brain"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	tg "gopkg.in/telegram-bot-api.v4"
)

var (
	startCmd = "/start"
	helpCmd  = "/help"
	addCmd   = "/add"
)

// Emoji :pencil2:
var addButton = addCmd + " \u270f\ufe0f"

var (
	// Markdown formatted
	helpReply = `Hello %s!
I help you study your own way.
You teach me some facts.
I help you to remember them.

Start by adding a new fact using the __/add__ button below.

After you added some facts type  _@slangbrainbot ..._ to search them.`

	// Emoji :memo:
	addReply = "Send me a new fact \U0001f4dd"

	// Emoji :scream:
	errReply = "Something went wrong, try again \U0001f631"

	// Emoji :+1:
	addDoneReply = "Fact added \U0001f44d"

	// Markdown formatted
	inlineQueryReply = `_%s_

%s`
)

type config struct {
	logger  *log.Logger
	verbose bool
	store   brain.Store
	bot     *tg.BotAPI
}

func main() {
	logger := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)

	verboseVar := os.Getenv("VERBOSE")
	var (
		verbose bool
		err     error
	)
	if len(verboseVar) > 0 {
		verbose, err = strconv.ParseBool(verboseVar)
		if err != nil {
			logger.Println("Env var VERBOSE has to be boolean.")
			os.Exit(1)
		}
	}
	token := os.Getenv("BOT_TOKEN")
	if len(token) < 1 {
		logger.Println("Env var BOT_TOKEN not set. Get a token from Botfather.")
		os.Exit(1)
	}
	dbFile := os.Getenv("DB_FILE")
	if len(dbFile) < 1 {
		logger.Println("Env var DB_FILE not set. Specify the path for your sqlite database file.")
		os.Exit(1)
	}
	addr := os.Getenv("ADDR")
	if len(addr) < 1 {
		logger.Println("Env var ADDR not set. Use address to bind webhook server to. Can also be only :<port>.")
		os.Exit(1)
	}
	webhook, err := url.Parse(os.Getenv("WEBHOOK"))
	if err != nil {
		logger.Println("Env var WEBHOOK not valid. Specify the URL Telegram should send updates to.")
		os.Exit(1)
	}

	store, err := brain.CreateStore("sqlite3", dbFile)
	if err != nil {
		logger.Println("failed to create store: ", err)
		os.Exit(1)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			logger.Println(err)
		}
	}()

	bot, err := tg.NewBotAPI(token)
	if err != nil {
		logger.Println("failed to connect to Telegram bot API:", err)
		os.Exit(1)
	}
	if verbose {
		logger.Println("Authorized on account", bot.Self.UserName)
	}

	webhook.Path = path.Join(webhook.Path, bot.Token)
	_, err = bot.SetWebhook(tg.NewWebhook(webhook.String()))
	if err != nil {
		logger.Println(err)
		os.Exit(1)
	}

	updates := bot.ListenForWebhook("/" + bot.Token)
	go func() {
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			logger.Println(err)
		}
	}()

	c := &config{logger, verbose, store, bot}

	for update := range updates {
		handleMessage(c, update.Message)
		handleInlineQuery(c, update.InlineQuery)
	}
}

func handleMessage(c *config, msg *tg.Message) {
	if msg == nil {
		return
	}

	if c.verbose {
		c.logger.Println("[m]", msg.From.UserName, "-", msg.Text)
	}

	reply, err := getReply(c.store, msg)
	if err != nil {
		c.logger.Println(err)
	}
	// Ignore unhandled messages
	if reply.Text == "" {
		return
	}
	_, err = c.bot.Send(reply)
	if err != nil {
		c.logger.Println(errors.Wrap(err, "failed sending reply"))
	}
}

func getReply(store brain.Store, msg *tg.Message) (tg.MessageConfig, error) {
	chatTitle := msg.Chat.Title
	if chatTitle == "" {
		chatTitle = msg.Chat.UserName
	}

	chat, err := store.UseChat(msg.Chat.ID, msg.From.ID, chatTitle)
	if err != nil {
		return replyWithKeyboard(chat.ID, errReply), err
	}

	switch {
	// `/start` or `/help`
	case strings.HasPrefix(msg.Text, startCmd) || strings.HasPrefix(msg.Text, helpCmd):
		reply := replyWithKeyboard(chat.ID, fmt.Sprintf(helpReply, msg.From.FirstName))
		reply.ParseMode = "Markdown"
		return reply, nil
	// `/add`
	case strings.HasPrefix(msg.Text, addCmd):
		err := store.SetMode(chat.ID, brain.AddMode)
		return tg.NewMessage(chat.ID, addReply), err
	// `/add` reply
	case chat.Mode == brain.AddMode:
		err := store.AddFact(chat.ID, msg.Text)
		if err != nil {
			return replyWithKeyboard(chat.ID, errReply), err
		}
		err = store.SetMode(chat.ID, brain.IdleMode)
		return replyWithKeyboard(chat.ID, addDoneReply), err
	}

	return tg.MessageConfig{}, nil
}

func handleInlineQuery(c *config, q *tg.InlineQuery) {
	if q == nil {
		return
	}

	if c.verbose {
		c.logger.Println("[q]", q.From.UserName, "-", q.Query)
	}

	userID := q.From.ID
	var results []interface{}
	facts, err := c.store.FindFacts(userID, q.Query)
	if err != nil {
		c.logger.Println(err)
	}

	// search in all brains of user
	for i, fact := range facts {
		title := fmt.Sprintf("%v (%s) %s", i+1, fact.Chat.Title, firstChars(10, fact.Content))
		msg := fmt.Sprintf(inlineQueryReply, q.Query, fact.Content)
		result := tg.NewInlineQueryResultArticleMarkdown(strconv.Itoa(i+1), title, msg)
		result.Description = fact.Content
		results = append(results, result)
	}

	inlineConf := tg.InlineConfig{
		InlineQueryID: q.ID,
		IsPersonal:    true,
		Results:       results,
	}

	if _, err := c.bot.AnswerInlineQuery(inlineConf); err != nil {
		c.logger.Println(err)
	}
}

func replyWithKeyboard(chatID int64, text string) tg.MessageConfig {
	msg := tg.NewMessage(chatID, text)
	keyboard := tg.NewReplyKeyboard(tg.NewKeyboardButtonRow(tg.NewKeyboardButton(addButton)))
	keyboard.OneTimeKeyboard = true
	msg.ReplyMarkup = keyboard
	return msg
}

// TODO: don't break runes
func firstChars(i int, s string) string {
	if i > len(s) {
		return s[0:len(s)]
	}
	return s[0:i] + "..."
}
