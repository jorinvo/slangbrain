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
)

// Feedback describes a message from a user a human has to react to
type Feedback struct {
	ChatID   int64
	Username string
	Message  string
}

// Bot is a messenger bot handling webhook events and notifications.
// Use New to setup and use register Bot as a http.Handler.
type Bot struct {
	store          brain.Store
	err            *log.Logger
	info           *log.Logger
	client         fbot.Client
	feedback       chan<- Feedback
	notifyInterval time.Duration
	http.Handler
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

// Notify enables the sending of notifications in the give interval.
func Notify(interval time.Duration) func(*Bot) {
	return func(b *Bot) {
		b.notifyInterval = interval
	}
}

// GetFeedback sets up user feedback to be sent to the given channel.
func GetFeedback(f chan<- Feedback) func(*Bot) {
	return func(b *Bot) {
		b.feedback = f
	}
}

// New sets up and starts a messenger bot.
// Greetings and Getting started messages are set
// and notfication sending is run in an interval.
// It returns the HTTP handler for the webhook.
// The options LogInfo, LogErr, Notify, GetFeedback can be used.
func New(store brain.Store, token, verifyToken string, options ...func(*Bot)) (Bot, error) {
	client := fbot.New(token, verifyToken)
	b := Bot{
		store:  store,
		client: client,
	}
	b.Handler = client.Webhook(b.HandleEvent)

	for _, option := range options {
		option(&b)
	}
	if b.info == nil {
		b.info = log.New(ioutil.Discard, "", 0)
	}
	if b.err == nil {
		b.err = log.New(ioutil.Discard, "", 0)
	}

	if err := client.SetGreetings(map[string]string{"default": greeting}); err != nil {
		return b, fmt.Errorf("failed to set greeting: %v", err)
	}
	b.info.Println("Greeting set")

	if err := client.SetGetStartedPayload(payloadGetStarted); err != nil {
		return b, fmt.Errorf("failed to enable Get Started button: %v", err)
	}
	b.info.Printf("Get Started button activated")

	if b.notifyInterval > 0 {
		b.info.Println("Notifications enabled")
		go func() {
			for range time.Tick(b.notifyInterval) {
				b.info.Println("Sending notifications")
				dueStudies, err := b.store.GetDueStudies()
				if err != nil {
					b.err.Println(err)
					return
				}
				now := time.Now()
				for chatID, count := range dueStudies {
					p, err := b.client.GetProfile(chatID)
					name := p.Name
					if err != nil {
						name = "there"
						b.err.Printf("failed to get profile for %d: %v", chatID, err)
					}
					msg := fmt.Sprintf(messageStudiesDue, name, count)
					if err = b.client.Send(chatID, msg, buttonsStudiesDue); err != nil {
						b.err.Printf("failed to notify user %d: %v", chatID, err)
					}
					b.trackActivity(chatID, now)
				}
			}
		}()
	}

	return b, nil
}

// SendMessage sends a message to a specific user.
func (b Bot) SendMessage(id int64, msg string) error {
	if err := b.client.Send(id, msg, nil); err != nil {
		return err
	}
	b.send(b.messageStartMenu(id))
	return nil
}
