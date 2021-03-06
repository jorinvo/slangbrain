package integration

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jorinvo/slangbrain/bot"
)

func TestFeedback(t *testing.T) {
	store, cleanup := initDB(t)
	defer cleanup()

	tt := []testCase{
		{
			name:     "get profile",
			method:   "GET",
			url:      "/123?fields=first_name,locale,timezone&access_token=some-test-token&appsecret_proof=e5565c0a91022866f93ae581ad8e3bddca01e06c067b5816f0373fc76df3d1f0",
			response: `{ "first_name": "Max", "locale": "de_DE" }`,
		},
		{
			name:   "welcome 1",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Hallo Max!\n\nJedes Mal wenn du ein neues Wort im Alltag lernst, füge es einfach zu Slangbrain hinzu und vergesse es nie wieder.\n\nNachdem du Vokabeln gespeichert hast, wird Slangbrain dich automatisch in sinnvollen Abständen abfragen und du wirst dich immer an die Wörter erinnern."}}`,
		},
		{
			name:   "welcome 2",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Bitte schicke jetzt einen Satz in der Sprache die du lernst, und nach einer leeren Zeile kannst du eine Erklärung auf Deutsch hinzufügen.\n\nEin Beispiel wäre, wenn du Französisch lernst, dann könntest du folgende Nachricht schicken:"}}`,
		},
		{
			name:   "welcome 3",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Bonjour !\nGuten Tag!"}}`,
		},
		{
			name:   "welcome 4",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Jetzt bist du dran:"}}`,
			send:   fmt.Sprintf(formatMessage, "2", `Hola\nHallo/Hi`),
		},
		{
			name:   "save phrase",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Gespeichert:\nHola\n\nMit Erklärung:\nHallo/Hi"}}`,
		},
		{
			name:   "add next",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Schicke die nächste Vokabel.","quick_replies":[{"content_type":"text","title":"stop","payload":"PAYLOAD_STARTMENU"}]}}`,
			send:   fmt.Sprintf(formatPayload, "PAYLOAD_STARTMENU"),
		},
		{
			name:   "menu",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Was willst du als nächstes machen?","quick_replies":[{"content_type":"text","title":"🏫 lernen","payload":"PAYLOAD_STARTSTUDY"},{"content_type":"text","title":"➕ neu","payload":"PAYLOAD_STARTADD"},{"content_type":"text","title":"❓ Hilfe","payload":"PAYLOAD_SHOWHELP"},{"content_type":"text","title":"✔ fertig","payload":"PAYLOAD_IDLE"}]}}`,
			send:   fmt.Sprintf(formatPayload, "PAYLOAD_SHOWHELP"),
		},
		{
			name:   "help",
			expect: `{"recipient":{"id":"123"},"message":{"attachment":{"type":"template","payload":{"template_type":"button","text":"Wie kann ich dir weiterhelfen?","buttons":[{"type":"web_url","title":"slangbrain.com","url":"https://slangbrain.com/de/blog/","webview_share_button":"hide"}]}},"quick_replies":[{"content_type":"text","title":"zurück","payload":"PAYLOAD_STARTMENU"},{"content_type":"text","title":"✔ Benachrichtigung","payload":"PAYLOAD_SUBSCRIBE"},{"content_type":"text","title":"Feedback geben","payload":"PAYLOAD_FEEDBACK"},{"content_type":"text","title":"Vokabeln importieren","payload":"PAYLOAD_IMPORTHELP"},{"content_type":"text","title":"API Token","payload":"PAYLOAD_GETTOKEN"}]}}`,
			send:   fmt.Sprintf(formatPayload, "PAYLOAD_FEEDBACK"),
		},
		{
			name:   "feedback",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Ein Problem ist aufgetreten, du hast einen Verbesserungsvorschlag für uns oder du willst einfach nur hallo sagen? Sende jetzt eine Nachricht und sie wird weitergeleitet an die Menschen die Slangbrain entschickeln.","quick_replies":[{"content_type":"text","title":"❌ abbrechen","payload":"PAYLOAD_STARTMENU"}]}}`,
			send:   fmt.Sprintf(formatMessage, "3", "Ich mag dich."),
		},
		{
			name:   "done",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Danke Max, wir melden uns bei dir sobald wie möglich."}}`,
		},
		{
			name:   "menu 2",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Was willst du als nächstes machen?","quick_replies":[{"content_type":"text","title":"🏫 lernen","payload":"PAYLOAD_STARTSTUDY"},{"content_type":"text","title":"➕ neu","payload":"PAYLOAD_STARTADD"},{"content_type":"text","title":"❓ Hilfe","payload":"PAYLOAD_SHOWHELP"},{"content_type":"text","title":"✔ fertig","payload":"PAYLOAD_IDLE"}]}}`,
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

	feedback := make(chan bot.Feedback)
	go func() {
		if f := <-feedback; f.ChatID != 123 || f.Username != "Max" || f.Message != "Ich mag dich." {
			t.Errorf("unexpected feedback: %v", f)
		}
	}()

	b, _, err := bot.New(bot.Config{
		Store:       store,
		Token:       token,
		Secret:      secret,
		ErrLogger:   log.New(os.Stderr, "", log.LstdFlags|log.Llongfile),
		FacebookURL: ts.URL,
		Feedback:    feedback,
	})
	fatal(t, err)

	go send(t, b, fmt.Sprintf(formatMessage, "1", "hi"))

	for s := range msg {
		if s != "" {
			go send(t, b, s)
		}
	}
}
