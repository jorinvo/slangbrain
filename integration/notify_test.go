package integration

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jorinvo/slangbrain/bot"
	"github.com/jorinvo/slangbrain/brain"
	"github.com/jorinvo/slangbrain/payload"
)

func TestNotify(t *testing.T) {
	store, cleanup := initDB(t)
	defer cleanup()
	fatal(t, store.SetMode(123, brain.ModeMenu))

	tt := []testCase{
		{
			name:     "get profile",
			method:   "GET",
			url:      "/123?fields=first_name,locale,timezone&access_token=some-test-token&appsecret_proof=e5565c0a91022866f93ae581ad8e3bddca01e06c067b5816f0373fc76df3d1f0",
			response: `{ "first_name": "Max", "locale": "de_DE" }`,
		},
		{
			name:   "help 1",
			expect: `{"recipient":{"id":"123"},"message":{"attachment":{"type":"template","payload":{"template_type":"button","text":"Wie kann ich dir weiterhelfen?","buttons":[{"type":"web_url","title":"slangbrain.com","url":"https://slangbrain.com/de/blog/","webview_share_button":"hide"}]}},"quick_replies":[{"content_type":"text","title":"zurück","payload":"PAYLOAD_STARTMENU"},{"content_type":"text","title":"✔ Benachrichtigung","payload":"PAYLOAD_SUBSCRIBE"},{"content_type":"text","title":"Feedback geben","payload":"PAYLOAD_FEEDBACK"},{"content_type":"text","title":"Vokabeln importieren","payload":"PAYLOAD_IMPORTHELP"},{"content_type":"text","title":"API Token","payload":"PAYLOAD_GETTOKEN"}]}}`,
			send:   fmt.Sprintf(formatPayload, payload.Subscribe),
		},
		{
			name:   "subscribed",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Ich schicke dir eine Nachricht sobald es etwas zu wiederholen gibt.\n\nWas willst du als nächstes machen?","quick_replies":[{"content_type":"text","title":"🏫 lernen","payload":"PAYLOAD_STARTSTUDY"},{"content_type":"text","title":"➕ neu","payload":"PAYLOAD_STARTADD"},{"content_type":"text","title":"❓ Hilfe","payload":"PAYLOAD_SHOWHELP"},{"content_type":"text","title":"✔ fertig","payload":"PAYLOAD_IDLE"}]}}`,
			send:   fmt.Sprintf(formatPayload, payload.Help),
		},
		{
			name:   "help 2",
			expect: `{"recipient":{"id":"123"},"message":{"attachment":{"type":"template","payload":{"template_type":"button","text":"Wie kann ich dir weiterhelfen?","buttons":[{"type":"web_url","title":"slangbrain.com","url":"https://slangbrain.com/de/blog/","webview_share_button":"hide"}]}},"quick_replies":[{"content_type":"text","title":"zurück","payload":"PAYLOAD_STARTMENU"},{"content_type":"text","title":"❌ Benachrichtigung","payload":"PAYLOAD_UNSUBSCRIBE"},{"content_type":"text","title":"Feedback geben","payload":"PAYLOAD_FEEDBACK"},{"content_type":"text","title":"Vokabeln importieren","payload":"PAYLOAD_IMPORTHELP"},{"content_type":"text","title":"API Token","payload":"PAYLOAD_GETTOKEN"}]}}`,
			send:   fmt.Sprintf(formatPayload, payload.Unsubscribe),
		},
		{
			name:   "unsubscribed",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Alles klar, du bekommst in Zukunft keine Benachrichtigungen mehr.\n\nWas willst du als nächstes machen?","quick_replies":[{"content_type":"text","title":"🏫 lernen","payload":"PAYLOAD_STARTSTUDY"},{"content_type":"text","title":"➕ neu","payload":"PAYLOAD_STARTADD"},{"content_type":"text","title":"❓ Hilfe","payload":"PAYLOAD_SHOWHELP"},{"content_type":"text","title":"✔ fertig","payload":"PAYLOAD_IDLE"}]}}`,
			send:   fmt.Sprintf(formatPayload, payload.Help),
		},
		{
			name:   "help 3",
			expect: `{"recipient":{"id":"123"},"message":{"attachment":{"type":"template","payload":{"template_type":"button","text":"Wie kann ich dir weiterhelfen?","buttons":[{"type":"web_url","title":"slangbrain.com","url":"https://slangbrain.com/de/blog/","webview_share_button":"hide"}]}},"quick_replies":[{"content_type":"text","title":"zurück","payload":"PAYLOAD_STARTMENU"},{"content_type":"text","title":"✔ Benachrichtigung","payload":"PAYLOAD_SUBSCRIBE"},{"content_type":"text","title":"Feedback geben","payload":"PAYLOAD_FEEDBACK"},{"content_type":"text","title":"Vokabeln importieren","payload":"PAYLOAD_IMPORTHELP"},{"content_type":"text","title":"API Token","payload":"PAYLOAD_GETTOKEN"}]}}`,
		},
	}

	state := 0
	msg := make(chan string)

	// Fake the Facebook server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tc := tt[state]
		checkCase(t, w, r, tc)
		msg <- tc.send
		state++
		if state == len(tt) {
			close(msg)
		}
	}))
	defer ts.Close()

	b, err := bot.New(
		store,
		token,
		secret,
		bot.LogErr(log.New(os.Stderr, "", log.LstdFlags|log.Llongfile)),
		bot.FAPI(ts.URL),
	)
	fatal(t, err)

	go send(t, b, fmt.Sprintf(formatPayload, payload.Help))

	for s := range msg {
		if s != "" {
			go send(t, b, s)
		}
	}
}
