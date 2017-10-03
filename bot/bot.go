// Package bot implements the Messenger bot
// and handles all the user interaction.
package bot

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/jorinvo/slangbrain/brain"
	"github.com/jorinvo/slangbrain/payload"
	"github.com/jorinvo/slangbrain/translate"
	"qvl.io/fbot"
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

// bot is a messenger bot handling webhook events and notifications.
// Use New to setup. Use Bot as a http.Handler.
type bot struct {
	store        brain.Store
	content      translate.Translator
	err          *log.Logger
	info         *log.Logger
	do           func(req *http.Request) (*http.Response, error)
	client       fbot.Client
	feedback     chan<- Feedback
	notifyTimers map[int64]*time.Timer
	messageDelay time.Duration
	furl         string
}

// Config to pass to new for creating a Bot.
type Config struct {
	Store        brain.Store          // Required.
	Token        string               // Required.
	Secret       string               // Required.
	VerifyToken  string               // Required.
	Logger       *log.Logger          // Optional. Logs are discared outerwise.
	ErrLogger    *log.Logger          // Optional. Errors are ignored outerwise.
	Feedback     chan<- Feedback      // Optional. Messages for admins are sent to this channel.
	Notify       bool                 // Enables sending notifications when studies are ready.
	Translator   translate.Translator // Optional. Set the translator service to enable linking.
	FacebookURL  string               // Optional. Overwrite the default URL of the Facebook API.
	MessageDelay time.Duration        // Optional. Time to wait between sending messages when sending multiple in a row.
	Setup        bool
	Doer         func(req *http.Request) (*http.Response, error) // Optional. Pass http.Client.Do. Default is http.DefaultClient.
}

// New creates a Bot.
// It can be used as an HTTP handler for the webhook.
func New(c Config) (http.Handler, func(id int64, msg string) error, error) {
	logs := c.Logger
	if logs == nil {
		logs = log.New(ioutil.Discard, "", 0)
	}
	errs := c.ErrLogger
	if errs == nil {
		errs = log.New(ioutil.Discard, "", 0)
	}

	if c.Token == "" {
		errs.Println("created Bot with empty token; cannot make API request")
	}

	if c.Secret == "" {
		errs.Println("created Bot with empty secret; cannot verify webhook requests")
	}

	translator := c.Translator
	if len(translator.Langs()) == 0 {
		logs.Println("created Bot without Translator; disabled manager link")
		translator = translate.New("")
	}

	doer := c.Doer
	if doer == nil {
		doer = http.DefaultClient.Do
	}

	feedback := c.Feedback
	if feedback == nil {
		f := make(chan Feedback)
		go func() {
			for f := range f {
				errs.Printf("[id=%d, name=%s, channel=%s] got unhandled feedback: %s", f.ChatID, f.Username, f.Channel, f.Message)
			}
		}()
		feedback = f
	}

	// Init map because scheduleNotify will check if it is initialized.
	var notifyTimers map[int64]*time.Timer
	if c.Notify {
		notifyTimers = map[int64]*time.Timer{}
	}

	b := bot{
		store:        c.Store,
		info:         logs,
		err:          errs,
		content:      translator,
		do:           doer,
		feedback:     feedback,
		client:       fbot.New(fbot.Config{Token: c.Token, Secret: c.Secret, API: c.FacebookURL}),
		notifyTimers: notifyTimers,
		messageDelay: c.MessageDelay,
	}
	h := b.client.Webhook(b.handleEvent, c.Secret, c.VerifyToken)

	if c.Setup {
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
			return h, nil, fmt.Errorf("failed to set greeting: %v", err)
		}
		logs.Println("Greeting set")
		if err := b.client.SetGetStartedPayload(payload.GetStarted); err != nil {
			return h, nil, fmt.Errorf("failed to enable Get Started button: %v", err)
		}
		logs.Printf("Get Started button activated")
	}

	if c.Notify {
		logs.Println("Notifications enabled")
		if err := b.store.EachActiveChat(b.scheduleNotify); err != nil {
			return h, nil, err
		}
	}

	return h, b.SendMessage, nil
}

// SendMessage sends a message to a specific user.
func (b bot) SendMessage(id int64, msg string) error {
	if err := b.client.Send(id, msg, nil); err != nil {
		return err
	}
	u := b.getUser(id)
	b.send(b.messageStartMenu(u))
	return nil
}

// handleEvent handles a Messenger event.
func (b bot) handleEvent(e fbot.Event) {
	if e.Type == fbot.EventError {
		b.err.Println("webhook error:", e.Text)
		return
	}

	if e.Type == fbot.EventRead {
		if err := b.store.SetRead(e.ChatID, e.Time); err != nil {
			b.err.Printf("set read fail: %d, %v\n", e.ChatID, e.Time)
		}
		b.scheduleNotify(e.ChatID)
		return
	}

	u := b.getUser(e.ChatID)

	if e.Type == fbot.EventReferral {
		ref, err := url.QueryUnescape(e.Ref)
		if err != nil {
			b.err.Printf("non-unescapeable ref %#v for %d: %v\n", e.Ref, u.ID, err)
			return
		}
		if links := getLinks(ref); links != nil {
			b.handleLinks(u, links)
			return
		}
		b.err.Printf("unhandled ref for %d: %#v\n", u.ID, e.Ref)
		return
	}

	if e.Type == fbot.EventPayload {
		b.handlePayload(u, e.Payload, e.Ref)
		return
	}

	if err := b.store.QueueMessage(e.MessageID); err != nil {
		if err == brain.ErrExists {
			b.info.Printf("Message already processed: %v", e.MessageID)
			return
		}
		b.err.Println("unqueued message ID:", err)
		return
	}

	if e.Type == fbot.EventMessage {
		b.handleMessage(u, e.Text)
		return
	}

	if e.Type == fbot.EventAttachment {
		b.handleAttachments(u, e.Attachments)
		return
	}

	b.err.Printf("unhandled event: %#v\n", e)
}
