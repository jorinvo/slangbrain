package fbot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// EventType ...
type EventType int

const (
	// EventUnknow ...
	EventUnknow EventType = iota
	// EventMessage ...
	EventMessage
	// EventPayload ...
	EventPayload
	// EventRead ...
	EventRead
)

// Event ...
type Event struct {
	Type    EventType
	ChatID  int64
	Time    time.Time
	Seq     int
	Text    string
	Payload string
}

// Webhook ...
func (c Client) Webhook(handler func(Event)) http.Handler {
	return webhook{EventHandler: handler, VerifyToken: c.verifyToken}
}

type webhook struct {
	EventHandler func(Event)
	VerifyToken  string
}

func (wh webhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		wh.verifyHandler(w, r)
		return
	}

	var rec receive
	if err := json.NewDecoder(r.Body).Decode(&rec); err != nil {
		fmt.Println("could not decode response:", err)
		fmt.Fprintln(w, `{status: 'not ok'}`)
		return
	}
	defer func() {
		_ = r.Body.Close()
	}()

	for _, e := range rec.Entry {
		for _, m := range e.Messaging {
			event := createEvent(m)
			fmt.Printf("%+v\n", event)
			if event.Type != EventUnknow {
				wh.EventHandler(event)
			}
		}
	}

	fmt.Fprintln(w, `{status: 'ok'}`)
}

func (wh webhook) verifyHandler(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("hub.verify_token") == wh.VerifyToken {
		fmt.Fprintln(w, r.FormValue("hub.challenge"))
		return
	}
	fmt.Fprintln(w, "Incorrect verify token.")
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
				Seq:     m.Message.Seq,
				Payload: m.Message.QuickReply.Payload,
			}
		}
		return Event{
			Type:   EventMessage,
			ChatID: m.Sender.ID,
			Time:   msToTime(m.Timestamp),
			Seq:    m.Message.Seq,
			Text:   m.Message.Text,
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
			Seq:    m.Read.Seq,
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
	Seq        int         `json:"seq"`
	Text       string      `json:"text"`
	QuickReply *quickReply `json:"quick_reply,omitempty"`
}

type read struct {
	Watermark int64 `json:"watermark"`
	Seq       int   `json:"seq"`
}

type postback struct {
	Payload string `json:"payload"`
}
