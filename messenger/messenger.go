// Package messenger implements the Messenger bot
// and handles all the user interaction.
package messenger

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/jorinvo/slangbrain/brain"
	"github.com/jorinvo/slangbrain/fbot"
	"github.com/jorinvo/slangbrain/payload"
	"github.com/jorinvo/slangbrain/translate"
	"github.com/jorinvo/slangbrain/user"
)

// Channel to send unhandled user messages and attachments to
const slackUnhandled = "#slangbrain-unhandled"

// Feedback describes a message from a user a human has to react to.
// Channel is optional and make sure to not forget the "#" in the beginning.
type Feedback struct {
	ChatID   int64
	Username string
	Message  string
	Channel  string
}

// Bot is a messenger bot handling webhook events and notifications.
// Use New to setup and use register Bot as a http.Handler.
type Bot struct {
	store        brain.Store
	content      translate.Translator
	err          *log.Logger
	info         *log.Logger
	do           func(req *http.Request) (*http.Response, error)
	client       fbot.Client
	feedback     chan<- Feedback
	notifyTimers map[int64]*time.Timer
	messageDelay time.Duration
	// Fields below are only used for initialization
	http.Handler
	setup         bool
	verifyToken   string
	furl          string
	hasTranslator bool
}

// Setup sends greetings and the getting started message to Facebook.
func Setup(b *Bot) {
	b.setup = true
}

// LogInfo is an option to set the info logger of the bot.
func LogInfo(l *log.Logger) func(*Bot) {
	return func(b *Bot) {
		b.info = l
	}
}

// LogErr is an option to set the error logger of the bot.
func LogErr(l *log.Logger) func(*Bot) {
	return func(b *Bot) {
		b.err = l
	}
}

// Verify is an option to enable verification of the webhook.
func Verify(token string) func(*Bot) {
	return func(b *Bot) {
		b.verifyToken = token
	}
}

// GetFeedback sets up user feedback to be sent to the given channel.
func GetFeedback(f chan<- Feedback) func(*Bot) {
	return func(b *Bot) {
		b.feedback = f
	}
}

// Notify enables sending notifications when studies are ready.
func Notify(b *Bot) {
	b.notifyTimers = map[int64]*time.Timer{}
}

// FAPI overwrites the default URL of the Facebook API.
// This is used for testing.
func FAPI(url string) func(*Bot) {
	return func(b *Bot) {
		b.furl = url
	}
}

// MessageDelay sets a time for which to wait between sending messages when sending multiple in a row.
func MessageDelay(t time.Duration) func(*Bot) {
	return func(b *Bot) {
		b.messageDelay = t
	}
}

// Translate sets the translator service to be used.
func Translate(t translate.Translator) func(*Bot) {
	return func(b *Bot) {
		b.hasTranslator = true
		b.content = t
	}
}

// HTTPDo sets the function to use for HTTP requests.
// http.Client.Do can be used. If not specified http.DefaultClient is used.
func HTTPDo(do func(req *http.Request) (*http.Response, error)) func(*Bot) {
	return func(b *Bot) {
		b.do = do
	}
}

// New creates a Bot.
// It can be used as a HTTP handler for the webhook.
// A store, a Facebook API token and a Facebook app secret are required.
// The options Setup, LogInfo, LogErr, Notify, Verify, GetFeedback, FURL, MessageDelay, Translate, HTTPDo can be used.
func New(store brain.Store, token, secret string, options ...func(*Bot)) (Bot, error) {
	b := Bot{
		store: store,
	}
	for _, option := range options {
		option(&b)
	}
	if b.info == nil {
		b.info = log.New(ioutil.Discard, "", 0)
	}
	if b.err == nil {
		b.err = log.New(ioutil.Discard, "", 0)
	}
	if token == "" {
		b.err.Println("created Bot with empty token; cannot make API request")
	}
	if secret == "" {
		b.err.Println("created Bot with empty secret; cannot verify webhook requests")
	}
	if b.hasTranslator == false {
		b.info.Println("created Bot without Translator; disabled manager link")
		b.content = translate.New("")
	}
	if b.do == nil {
		b.do = http.DefaultClient.Do
	}
	if b.furl == "" {
		b.client = fbot.New(token)
	} else {
		b.client = fbot.New(token, fbot.API(b.furl))
	}
	b.Handler = b.client.Webhook(b.HandleEvent, secret, b.verifyToken)
	if b.feedback == nil {
		feedback := make(chan Feedback)
		b.feedback = feedback
		go func() {
			for f := range feedback {
				b.err.Printf("[id=%d, name=%s, channel=%s] got unhandled feedback: %s", f.ChatID, f.Username, f.Channel, f.Message)
			}
		}()
	}

	if b.setup {
		greetings := []fbot.Greeting{
			{
				Locale: "default",
				Text:   b.content.Load("").Msg.Greeting,
			},
		}
		for _, lang := range b.content.Langs() {
			g := fbot.Greeting{
				Locale: lang,
				Text:   b.content.Load(lang).Msg.Greeting,
			}
			greetings = append(greetings, g)
		}
		if err := b.client.SetGreetings(greetings); err != nil {
			return b, fmt.Errorf("failed to set greeting: %v", err)
		}
		b.info.Println("Greeting set")
		if err := b.client.SetGetStartedPayload(payload.GetStarted); err != nil {
			return b, fmt.Errorf("failed to enable Get Started button: %v", err)
		}
		b.info.Printf("Get Started button activated")
	}

	if b.notifyTimers != nil {
		b.info.Println("Notifications enabled")
		if err := b.store.EachActiveChat(b.scheduleNotify); err != nil {
			return b, err
		}
	}

	return b, nil
}

// SendMessage sends a message to a specific user.
func (b Bot) SendMessage(id int64, msg string) error {
	if err := b.client.Send(id, msg, nil); err != nil {
		return err
	}
	u := user.Get(id, b.store, b.err, b.content, b.client.GetProfile)
	b.send(b.messageStartMenu(u))
	return nil
}
