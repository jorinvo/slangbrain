// Package slack provides an HTTP handler that can be used to communicate with users via Slack.
package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/kyokomi/emoji"
)

// Slack is an HTTP handler that can be used to communicate with users via Slack.
type Slack struct {
	err          *log.Logger
	hook         string
	token        string
	replyHandler func(int64, string) error
}

// Reply is an option to enable /slack to receive replies from Slack.
// token is used to validate posts to the webhook.
// fn is called with a chatID and a message.
func Reply(token string, fn func(int64, string) error) func(*Slack) {
	return func(a *Slack) {
		a.token = token
		a.replyHandler = fn
	}
}

// LogErr is an option to set the error logger.
func LogErr(l *log.Logger) func(*Slack) {
	return func(a *Slack) {
		a.err = l
	}
}

// New returns a new Slack which can be used as an http.Handler.
// Optionally pass Reply or LogErr.
func New(hook string, options ...func(*Slack)) Slack {
	a := Slack{
		hook: hook,
	}
	for _, option := range options {
		option(&a)
	}
	if a.err == nil {
		a.err = log.New(ioutil.Discard, "", 0)
	}
	return a
}

// ServeHTTP serves an endpoint that can be registered as an Outgoing Webhook with Slack.
func (a Slack) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := r.Body.Close(); err != nil {
			a.err.Println(err)
		}
	}()

	if r.Method != "POST" {
		slackError(w, fmt.Errorf("illegal method: %s", r.Method))
		return
	}
	if a.token == "" {
		slackError(w, fmt.Errorf("webhook is disabled"))
	}
	// Validate token
	if r.FormValue("token") != a.token {
		slackError(w, fmt.Errorf("invalid token"))
		return
	}
	// Ignore messages coming from bots
	if r.FormValue("bot_id") != "" {
		return
	}
	// Parse message into ID and message
	// Message needs to begin with a user id, followed by a message.
	text := strings.TrimSpace(r.FormValue("text"))
	fields := strings.Fields(text)
	if len(fields) < 2 {
		slackError(w, fmt.Errorf("missing ID"))
		return
	}
	firstField := fields[0]
	id, err := strconv.Atoi(firstField)
	if err != nil {
		slackError(w, fmt.Errorf("failed parsing ID: %v", err))
		return
	}
	// Convert emojis coming from Slack in the form like :smile: to unicode
	msg := emoji.Sprint(strings.TrimSpace(strings.TrimPrefix(text, firstField)))
	// Send message to user via registered callback
	if err := a.replyHandler(int64(id), msg); err != nil {
		slackError(w, err)
		return
	}
}

// HandleMessage can be called to send a user message to Slack.
func (a Slack) HandleMessage(id int64, name, msg, channel string) {
	slackMsg := struct {
		Username string `json:"username"`
		Text     string `json:"text"`
		Channel  string `json:"channel,omitempty"`
	}{
		Username: name,
		Text:     fmt.Sprintf("%d\n\n%s", id, msg),
		Channel:  channel,
	}
	buf, err := json.Marshal(slackMsg)
	if err != nil {
		a.err.Printf("json marshal %#v: %v", slackMsg, err)
		return
	}
	resp, err := http.Post(a.hook, "application/json", bytes.NewBuffer(buf))
	if err != nil {
		a.err.Printf("failed to post message from %s (%d) to Slack: %s", name, id, msg)
		return
	}
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			a.err.Printf("failed to read response for Slack message '%s' from %s (%d): %v", msg, name, id, err)
			return
		}
		a.err.Printf("HTTP status code is not OK (%d) for Slack message '%s' from %s (%d): %s", resp.StatusCode, msg, name, id, body)
		return
	}
}

func slackError(w http.ResponseWriter, err error) {
	fmt.Fprint(w, fmt.Sprintf(`{ "text": "Error sending message: %s." }`, err))
}
