package integration

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jorinvo/slangbrain/brain"
	"github.com/jorinvo/slangbrain/messenger"
)

func TestAdd(t *testing.T) {
	store, cleanup := initDB(t)
	defer cleanup()
	fatal(t, store.SetMode(123, brain.ModeAdd))

	tt := []testCase{
		{
			name:   "save phrase",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Saved phrase:\nHola\n\nWith explanation:\nHello/Hi"}}`,
		},
		{
			name:   "add next",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Add next phrase.","quick_replies":[{"content_type":"text","title":"stop adding","payload":"PAYLOAD_STARTMENU"}]}}`,
			send:   fmt.Sprintf(formatMessage, "2", `Gracias\n\nThank you`),
		},
		{
			name:   "save phrase 2",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Saved phrase:\nGracias\n\nWith explanation:\nThank you"}}`,
		},
		{
			name:   "add next 2",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Add next phrase.","quick_replies":[{"content_type":"text","title":"stop adding","payload":"PAYLOAD_STARTMENU"}]}}`,
			send:   fmt.Sprintf(formatPayload, "PAYLOAD_STARTMENU"),
		},
		{
			name:   "menu",
			expect: `{"recipient":{"id":"123"},"message":{"text":"What would you like to do next?\nPlease use the buttons below.","quick_replies":[{"content_type":"text","title":"🏫 study","payload":"PAYLOAD_STARTSTUDY"},{"content_type":"text","title":"➕ phrases","payload":"PAYLOAD_STARTADD"},{"content_type":"text","title":"❓ help","payload":"PAYLOAD_SHOWHELP"},{"content_type":"text","title":"✔ done","payload":"PAYLOAD_IDLE"}]}}`,
			send:   fmt.Sprintf(formatPayload, "PAYLOAD_STARTSTUDY"),
		},
		{
			name:   "ask to subscribe",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Congrats, you finished all your studies for now!\nCome back in 2 hours.\n\nWould you like me to send you a message when there are phrases ready for studying?","quick_replies":[{"content_type":"text","title":"👌 sounds good","payload":"PAYLOAD_SUBSCRIBE"},{"content_type":"text","title":"no thanks","payload":"PAYLOAD_NOSUBSCRIPTION"}]}}`,
			send:   fmt.Sprintf(formatPayload, "PAYLOAD_NOSUBSCRIPTION"),
		},
		{
			name:   "refuse to subscribe",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Sure, you won't receive any notifications.\n\nWhat would you like to do next?\nPlease use the buttons below.","quick_replies":[{"content_type":"text","title":"🏫 study","payload":"PAYLOAD_STARTSTUDY"},{"content_type":"text","title":"➕ phrases","payload":"PAYLOAD_STARTADD"},{"content_type":"text","title":"❓ help","payload":"PAYLOAD_SHOWHELP"},{"content_type":"text","title":"✔ done","payload":"PAYLOAD_IDLE"}]}}`,
			send:   fmt.Sprintf(formatPayload, "PAYLOAD_IDLE"),
		},
		{
			name:   "idle",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Good, just send me a 👍 to continue with your studies."}}`,
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

	bot, err := messenger.New(
		store,
		token,
		messenger.LogErr(log.New(os.Stderr, "", log.LstdFlags|log.Llongfile)),
		messenger.FAPI(ts.URL),
	)
	fatal(t, err)

	go send(t, bot, fmt.Sprintf(formatMessage, "1", `Hola\nHello/Hi`))

	for s := range done {
		if s != "" {
			go send(t, bot, s)
		}
		state++
		if state >= len(tt) {
			close(done)
		}
	}
}