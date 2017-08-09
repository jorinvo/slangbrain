package integration

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jorinvo/slangbrain/messenger"
)

func TestFeedback(t *testing.T) {
	store, cleanup := initDB(t)
	defer cleanup()

	tt := []testCase{
		{
			name:     "get profile",
			method:   "GET",
			url:      "/123?fields=first_name,locale,timezone&access_token=some-test-token",
			response: `{ "first_name": "Smith", "locale": "us" }`,
		},
		{
			name:   "first welcome message",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Hello Smith!\n\nWhenever you pick up a new phrase, just add it to your Slangbrain and remember it forever.\n\nYou begin by adding phrases and later Slangbrain will test your memories in a natural schedule."}}`,
		},
		{
			name:   "second welcome message",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Please send me a phrase and its explanation.\nSeparate them with a linebreak.\nDon't worry if you send something wrong. You can delete phrases later.\n\nIf your mother tongue is English and you're studying Spanish, a message would look like this:\n\nHola\nHello\n\nGive it a try:"}}`,
			send:   fmt.Sprintf(formatMessage, "2", `Hola\nHello/Hi`),
		},
		{
			name:   "save phrase",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Saved phrase:\nHola\n\nWith explanation:\nHello/Hi"}}`,
		},
		{
			name:   "add next",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Add next phrase.","quick_replies":[{"content_type":"text","title":"stop adding","payload":"PAYLOAD_STARTMENU"}]}}`,
			send:   fmt.Sprintf(formatPayload, "PAYLOAD_STARTMENU"),
		},
		{
			name:   "menu",
			expect: `{"recipient":{"id":"123"},"message":{"text":"What would you like to do next?\nPlease use the buttons below.","quick_replies":[{"content_type":"text","title":"🏫 study","payload":"PAYLOAD_STARTSTUDY"},{"content_type":"text","title":"➕ phrases","payload":"PAYLOAD_STARTADD"},{"content_type":"text","title":"❓ help","payload":"PAYLOAD_SHOWHELP"},{"content_type":"text","title":"✔ done","payload":"PAYLOAD_IDLE"}]}}`,
			send:   fmt.Sprintf(formatPayload, "PAYLOAD_SHOWHELP"),
		},
		{
			name:   "help",
			expect: `{"recipient":{"id":"123"},"message":{"text":"How can I help you?","quick_replies":[{"content_type":"text","title":"send feedback","payload":"PAYLOAD_FEEDBACK"},{"content_type":"text","title":"all good","payload":"PAYLOAD_STARTMENU"}]}}`,
			send:   fmt.Sprintf(formatPayload, "PAYLOAD_FEEDBACK"),
		},
		{
			name:   "feedback",
			expect: `{"recipient":{"id":"123"},"message":{"text":"If you run into a problem, have any feedback for the people behind Slangbrain or just like to say hello, you can send a message now and we will get back to you as soon as possible.","quick_replies":[{"content_type":"text","title":"❌ cancel","payload":"PAYLOAD_STARTMENU"}]}}`,
			send:   fmt.Sprintf(formatMessage, "3", "I like you."),
		},
		{
			name:   "done",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Thanks Smith, you will hear from us soon."}}`,
		},
		{
			name:   "menu 2",
			expect: `{"recipient":{"id":"123"},"message":{"text":"What would you like to do next?\nPlease use the buttons below.","quick_replies":[{"content_type":"text","title":"🏫 study","payload":"PAYLOAD_STARTSTUDY"},{"content_type":"text","title":"➕ phrases","payload":"PAYLOAD_STARTADD"},{"content_type":"text","title":"❓ help","payload":"PAYLOAD_SHOWHELP"},{"content_type":"text","title":"✔ done","payload":"PAYLOAD_IDLE"}]}}`,
		},
	}

	state := 0
	done := make(chan string)

	// Fake the Facebook server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tc := tt[state]
		checkCase(t, w, r, tc)
		done <- tc.send
	}))
	defer ts.Close()

	feedback := make(chan messenger.Feedback, 1)
	bot, err := messenger.New(
		store,
		token,
		messenger.LogErr(log.New(os.Stderr, "", log.LstdFlags|log.Llongfile)),
		messenger.FAPI(ts.URL),
		messenger.GetFeedback(feedback),
	)
	fatal(t, err)

	go send(t, bot, fmt.Sprintf(formatMessage, "1", "hi"))

	for s := range done {
		if s != "" {
			go send(t, bot, s)
		}
		state++
		if state >= len(tt) {
			close(done)
		}
	}
	if f := <-feedback; f.ChatID != 123 || f.Username != "Smith" || f.Message != "I like you." {
		t.Errorf("unexpected feedback: %v", f)
	}
}
