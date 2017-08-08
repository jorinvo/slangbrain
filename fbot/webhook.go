package fbot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// EventType helps to distinguish the different type of events.
type EventType int

const (
	// EventUnknown is the default and will be used if none of the other types match.
	EventUnknown EventType = iota
	// EventMessage is triggered when a user sends Text, stickers or other content.
	// Only text is available at the moment.
	EventMessage
	// EventPayload is triggered when a quickReply or postback Payload is sent.
	EventPayload
	// EventRead is triggered when a user read a message.
	EventRead
	// EventError is triggered when the webhook is called with invalid JSON content.
	EventError
)

// Event contains information about a user action.
type Event struct {
	// Type helps to decide how to react to an event.
	Type EventType
	// ChatID identifies the user. It's a Facebook user ID.
	ChatID int64
	// Time describes when the event occured.
	Time time.Time
	// Text is a message a user send for EventMessage and error description for EventError.
	Text string
	// Payload is a predefined payload for a quick reply or postback sent with EventPayload.
	Payload string
	// MessageID is a unique ID for each message.
	MessageID string
}

// Webhook returns a handler for HTTP requests that can be registered with Facebook.
// The passed event handler will be called with all received events.
func (c Client) Webhook(handler func(Event), verifyToken string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			if r.FormValue("hub.verify_token") == verifyToken {
				fmt.Fprintln(w, r.FormValue("hub.challenge"))
				return
			}
			fmt.Fprintln(w, "Incorrect verify token.")
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var rec receive
		if err := json.NewDecoder(r.Body).Decode(&rec); err != nil {
			http.Error(w, "JSON invalid", http.StatusBadRequest)
			handler(Event{Type: EventError, Text: err.Error()})
			return
		}
		_ = r.Body.Close()

		// Return response as soon as possible.
		// Facebook doesn't care about the event handling.
		// Responses are sent separatly.
		fmt.Fprintln(w, `{"status":"ok"}`)

		for _, e := range rec.Entry {
			for _, m := range e.Messaging {
				event := createEvent(m)
				if event.Type != EventUnknown {
					handler(event)
				}
			}
		}
	})
}

func createEvent(m messageInfo) Event {
	if m.Message != nil {
		if m.Message.IsEcho {
			return Event{}
		}
		if m.Message.QuickReply != nil {
			return Event{
				Type:    EventPayload,
				ChatID:  m.Sender.ID,
				Time:    msToTime(m.Timestamp),
				Payload: m.Message.QuickReply.Payload,
			}
		}
		return Event{
			Type:      EventMessage,
			ChatID:    m.Sender.ID,
			Time:      msToTime(m.Timestamp),
			Text:      m.Message.Text,
			MessageID: m.Message.MID,
		}
	}
	if m.Postback != nil {
		return Event{
			Type:    EventPayload,
			ChatID:  m.Sender.ID,
			Time:    msToTime(m.Timestamp),
			Payload: m.Postback.Payload,
		}
	}
	if m.Read != nil {
		return Event{
			Type:   EventRead,
			ChatID: m.Sender.ID,
			Time:   msToTime(m.Read.Watermark),
		}
	}
	return Event{}
}

func msToTime(ms int64) time.Time {
	return time.Unix(ms/int64(time.Microsecond), 0)
}

type receive struct {
	Entry []entry `json:"entry"`
}

type entry struct {
	Messaging []messageInfo `json:"messaging"`
}

type messageInfo struct {
	Sender    sender    `json:"sender"`
	Timestamp int64     `json:"timestamp"`
	Message   *message  `json:"message"`
	Postback  *postback `json:"postback"`
	Read      *read     `json:"read"`
}

type sender struct {
	ID int64 `json:"id,string"`
}

type message struct {
	IsEcho     bool        `json:"is_echo,omitempty"`
	Text       string      `json:"text"`
	QuickReply *quickReply `json:"quick_reply,omitempty"`
	MID        string      `json:"mid,omitempty"`
}

type read struct {
	Watermark int64 `json:"watermark"`
}

type postback struct {
	Payload string `json:"payload"`
}
