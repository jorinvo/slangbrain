package main

// Thanks for help with unicode http://apps.timwhitlock.info/emoji/tables/unicode

import (
	"fmt"
	"log"
	"strings"

	tg "gopkg.in/telegram-bot-api.v4"
)

type chatBrains map[int64][]string
type int64Set map[int64]bool
type userChats map[int]int64Set

var welcomeMsg = `Hello %s!
I will help you study your own way.
You teach me some facts.
I help you to remember them.

Start by adding a new fact using the button below.`

var startCmd = "/start"
var addCmd = "/add"

// Emoji :pencil2:
var addButton = addCmd + " \xE2\x9C\x8F"

// Emoji :memo:
var addReply = "Send me a new fact \xF0\x9F\x93\x9D"

// Emoji :scream:
var defaultReply = "Something went wrong, try again \xF0\x9F\x98\xB1"

// Emoji :+1:
var addDoneReply = "Fact added \xF0\x9F\x91\x8D"

func main() {
	bot, err := tg.NewBotAPI("241302379:AAH6qAM-bnqp81zR2pl_PQePHec-FpYwRBI")
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tg.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	var brains = make(chatBrains)
	var chats = make(userChats)

	for update := range updates {
		handleMessage(bot, update.Message, brains, chats)
		handleInlineQuery(bot, update.InlineQuery, brains, chats)
	}
}

func handleMessage(bot *tg.BotAPI, msg *tg.Message, brains chatBrains, chats userChats) {
	if msg == nil {
		return
	}

	userID := msg.From.ID
	chatID := msg.Chat.ID
	text := msg.Text
	replyToMessage := msg.ReplyToMessage

	chats.add(userID, chatID)

	var reply tg.MessageConfig
	switch {
	case strings.HasPrefix(text, startCmd):
		reply = addKeyboard(tg.NewMessage(chatID, fmt.Sprintf(welcomeMsg, msg.From.FirstName)))
	case strings.HasPrefix(text, addCmd):
		reply = forceReply(tg.NewMessage(chatID, addReply))
	case replyToMessage != nil && replyToMessage.Text == addReply:
		brains[chatID] = append(brains[chatID], text)
		reply = addKeyboard(tg.NewMessage(chatID, addDoneReply))
		reply.ReplyToMessageID = msg.MessageID
	default:
		reply = addKeyboard(tg.NewMessage(chatID, defaultReply))
	}

	log.Printf("  msg: [%s] %s", msg.From.UserName, text)

	_, err := bot.Send(reply)

	if err != nil {
		log.Print(err)
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

func (chats userChats) add(userID int, chatID int64) {
	if chats[userID] == nil {
		chats[userID] = make(int64Set)
	}
	chats[userID][chatID] = true
}

func handleInlineQuery(bot *tg.BotAPI, q *tg.InlineQuery, brains chatBrains, chats userChats) {
	if q == nil {
		return
	}

	log.Printf("query: [%s] %s", q.From.UserName, q.Query)

	userID := q.From.ID
	var results []interface{}
	// search in all brains of user
	for chatID := range chats[userID] {
		for i, fact := range brains[chatID] {
			// query id is used by Telegram for caching
			result := tg.NewInlineQueryResultArticle(fmt.Sprint(userID, chatID, i), fmt.Sprintf("Fact %v", i), fact)
			result.Description = fact
			results = append(results, result)
		}
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
