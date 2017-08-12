package integration

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jorinvo/slangbrain/brain"
	"github.com/jorinvo/slangbrain/messenger"
)

func TestStudy(t *testing.T) {
	store, cleanup := initDB(t)
	defer cleanup()
	fatal(t, store.SetMode(123, brain.ModeStudy))
	yesterday := time.Now().Add(-24 * time.Hour)
	store.AddPhrase(123, "phrase1", "explanation1", yesterday)
	store.AddPhrase(123, "phrase2", "explanation2", yesterday)
	store.AddPhrase(123, "phrase3", "explanation3", yesterday)
	store.AddPhrase(123, "phrase4", "explanation4", yesterday)
	store.AddPhrase(123, "phrase5", "explanation5", yesterday)
	store.AddPhrase(123, "phrase6", "explanation6", yesterday)
	store.AddPhrase(123, "phrase7", "explanation7", time.Now())

	tt := []testCase{
		{
			name:     "get profile",
			method:   "GET",
			url:      "/123?fields=first_name,locale,timezone&access_token=some-test-token",
			response: `{ "first_name": "Smith", "locale": "us" }`,
		},
		{
			name:   "correct",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Correct!"}}`,
		},
		{
			name:   "review 2",
			expect: `{"recipient":{"id":"123"},"message":{"text":"5. Do you remember how to say this?\n\nexplanation2\n\nUse the buttons or type the phrase.","quick_replies":[{"content_type":"text","title":"❌","payload":"PAYLOAD_DELETE"},{"content_type":"text","title":"done studying","payload":"PAYLOAD_STARTMENU"},{"content_type":"text","title":"👉 show phrase","payload":"PAYLOAD_SHOWSTUDY"}]}}`,
			send:   fmt.Sprintf(formatMessage, "2", "wrong"),
		},
		{
			name:   "wrong",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Sorry, the right version is:\n\nphrase2"}}`,
		},
		{
			name:   "review 3",
			expect: `{"recipient":{"id":"123"},"message":{"text":"4. Do you remember how to say this?\n\nexplanation3\n\nUse the buttons or type the phrase.","quick_replies":[{"content_type":"text","title":"❌","payload":"PAYLOAD_DELETE"},{"content_type":"text","title":"done studying","payload":"PAYLOAD_STARTMENU"},{"content_type":"text","title":"👉 show phrase","payload":"PAYLOAD_SHOWSTUDY"}]}}`,
			send:   fmt.Sprintf(formatPayload, "PAYLOAD_SHOWSTUDY"),
		},
		{
			name:   "score bad",
			expect: `{"recipient":{"id":"123"},"message":{"text":"phrase3","quick_replies":[{"content_type":"text","title":"❌","payload":"PAYLOAD_DELETE"},{"content_type":"text","title":"👎 didn't know","payload":"PAYLOAD_SCOREBAD"},{"content_type":"text","title":"🤔","payload":"PAYLOAD_SCOREOK"},{"content_type":"text","title":"👌 got it","payload":"PAYLOAD_SCOREGOOD"}]}}`,
			send:   fmt.Sprintf(formatPayload, "PAYLOAD_SCOREBAD"),
		},
		{
			name:   "review 4",
			expect: `{"recipient":{"id":"123"},"message":{"text":"3. Do you remember how to say this?\n\nexplanation4\n\nUse the buttons or type the phrase.","quick_replies":[{"content_type":"text","title":"❌","payload":"PAYLOAD_DELETE"},{"content_type":"text","title":"done studying","payload":"PAYLOAD_STARTMENU"},{"content_type":"text","title":"👉 show phrase","payload":"PAYLOAD_SHOWSTUDY"}]}}`,
			send:   fmt.Sprintf(formatPayload, "PAYLOAD_DELETE"),
		},
		{
			name:   "delete",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Are you sure, you want to delete this phrase?","quick_replies":[{"content_type":"text","title":"❌ delete phrase","payload":"PAYLOAD_CONFIRMDELETE"},{"content_type":"text","title":"cancel","payload":"PAYLOAD_CANCELDELETE"}]}}`,
			send:   fmt.Sprintf(formatPayload, "PAYLOAD_CANCELDELETE"),
		},
		{
			name:   "delete canceled",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Good, let's keep that phrase and continue studying."}}`,
		},
		{
			name:   "continue review",
			expect: `{"recipient":{"id":"123"},"message":{"text":"3. Do you remember how to say this?\n\nexplanation4\n\nUse the buttons or type the phrase.","quick_replies":[{"content_type":"text","title":"❌","payload":"PAYLOAD_DELETE"},{"content_type":"text","title":"done studying","payload":"PAYLOAD_STARTMENU"},{"content_type":"text","title":"👉 show phrase","payload":"PAYLOAD_SHOWSTUDY"}]}}`,
			send:   fmt.Sprintf(formatPayload, "PAYLOAD_DELETE"),
		},
		{
			name:   "delete again",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Are you sure, you want to delete this phrase?","quick_replies":[{"content_type":"text","title":"❌ delete phrase","payload":"PAYLOAD_CONFIRMDELETE"},{"content_type":"text","title":"cancel","payload":"PAYLOAD_CANCELDELETE"}]}}`,
			send:   fmt.Sprintf(formatPayload, "PAYLOAD_CONFIRMDELETE"),
		},
		{
			name:   "confirm delete",
			expect: `{"recipient":{"id":"123"},"message":{"text":"The phrase has been deleted. Let's continue studying other phrases."}}`,
		},
		{
			name:   "review 5",
			expect: `{"recipient":{"id":"123"},"message":{"text":"2. Do you remember how to say this?\n\nexplanation5\n\nUse the buttons or type the phrase.","quick_replies":[{"content_type":"text","title":"❌","payload":"PAYLOAD_DELETE"},{"content_type":"text","title":"done studying","payload":"PAYLOAD_STARTMENU"},{"content_type":"text","title":"👉 show phrase","payload":"PAYLOAD_SHOWSTUDY"}]}}`,
			send:   fmt.Sprintf(formatPayload, "PAYLOAD_SHOWSTUDY"),
		},
		{
			name:   "score ok",
			expect: `{"recipient":{"id":"123"},"message":{"text":"phrase5","quick_replies":[{"content_type":"text","title":"❌","payload":"PAYLOAD_DELETE"},{"content_type":"text","title":"👎 didn't know","payload":"PAYLOAD_SCOREBAD"},{"content_type":"text","title":"🤔","payload":"PAYLOAD_SCOREOK"},{"content_type":"text","title":"👌 got it","payload":"PAYLOAD_SCOREGOOD"}]}}`,
			send:   fmt.Sprintf(formatPayload, "PAYLOAD_SCOREOK"),
		},
		{
			name:   "review 6",
			expect: `{"recipient":{"id":"123"},"message":{"text":"1. Do you remember how to say this?\n\nexplanation6\n\nUse the buttons or type the phrase.","quick_replies":[{"content_type":"text","title":"❌","payload":"PAYLOAD_DELETE"},{"content_type":"text","title":"done studying","payload":"PAYLOAD_STARTMENU"},{"content_type":"text","title":"👉 show phrase","payload":"PAYLOAD_SHOWSTUDY"}]}}`,
			send:   fmt.Sprintf(formatPayload, "PAYLOAD_SHOWSTUDY"),
		},
		{
			name:   "score good",
			expect: `{"recipient":{"id":"123"},"message":{"text":"phrase6","quick_replies":[{"content_type":"text","title":"❌","payload":"PAYLOAD_DELETE"},{"content_type":"text","title":"👎 didn't know","payload":"PAYLOAD_SCOREBAD"},{"content_type":"text","title":"🤔","payload":"PAYLOAD_SCOREOK"},{"content_type":"text","title":"👌 got it","payload":"PAYLOAD_SCOREGOOD"}]}}`,
			send:   fmt.Sprintf(formatPayload, "PAYLOAD_SCOREGOOD"),
		},
		{
			name:   "done",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Congrats, you finished all your studies for now!\nCome back in 2 hours.\n\nWould you like me to send you a message when there are phrases ready for studying?","quick_replies":[{"content_type":"text","title":"👌 sounds good","payload":"PAYLOAD_SUBSCRIBE"},{"content_type":"text","title":"no thanks","payload":"PAYLOAD_NOSUBSCRIPTION"}]}}`,
		},
	}

	state := 0
	msg := make(chan string)

	// Fake the Facebook server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tc := tt[state]
		checkCase(t, w, r, tc)
		msg <- tc.send
	}))
	defer ts.Close()

	bot, err := messenger.New(
		store,
		token,
		messenger.LogErr(log.New(os.Stderr, "", log.LstdFlags|log.Llongfile)),
		messenger.FAPI(ts.URL),
	)
	fatal(t, err)

	go send(t, bot, fmt.Sprintf(formatMessage, "1", `phrase1`))

	for s := range msg {
		if s != "" {
			go send(t, bot, s)
		}
		state++
		if state >= len(tt) {
			close(msg)
		}
	}
}
