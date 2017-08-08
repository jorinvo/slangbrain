package messenger_test

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"io/ioutil"

	"github.com/jorinvo/slangbrain/brain"
	"github.com/jorinvo/slangbrain/messenger"
)

var (
	url   = "https://api.slangbrain.com/"
	token = "some-test-token"
)

func initDB(t *testing.T) (brain.Store, func()) {
	f, err := ioutil.TempFile("", "slangbrain-test")
	fatal(t, err)
	fatal(t, f.Close())
	store, err := brain.New(f.Name())
	fatal(t, err)
	return store, func() {
		fatal(t, store.Close())
		fatal(t, os.Remove(f.Name()))
	}
}

func TestSample(t *testing.T) {
	store, cleanup := initDB(t)
	defer cleanup()

	// These are messages that the Facebook Server will receive and responses it will send to Slangbrain.
	tt := []struct {
		name     string
		url      string
		method   string
		message  string
		response string
	}{
		{
			name:     "get profile",
			method:   "GET",
			url:      "/123?fields=first_name,locale,timezone&access_token=some-test-token",
			response: `{ "first_name": "Smith", "locale": "us" }`,
		},
		{
			name:    "first welcome message",
			method:  "POST",
			url:     "/me/messages?access_token=some-test-token",
			message: `{"recipient":{"id":"123"},"message":{"text":"Hello Smith!\n\nWhenever you pick up a new phrase, just add it to your Slangbrain and remember it forever.\n\nYou begin by adding phrases and later Slangbrain will test your memories in a natural schedule."}}`,
		},
		{
			name:    "second welcome message",
			method:  "POST",
			url:     "/me/messages?access_token=some-test-token",
			message: `{"recipient":{"id":"123"},"message":{"text":"Please send me a phrase and its explanation.\nSeparate them with a linebreak.\nDon't worry if you send something wrong. You can delete phrases later.\n\nIf your mother tongue is English and you're studying Spanish, a message would look like this:\n\nHola\nHello\n\nGive it a try:"}}`,
		},
	}

	// Track state to make sure responses are in order.
	state := 0

	// Fake the Facebook server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tc := tt[state]
		state++
		t.Run(tc.name, func(t *testing.T) {
			if r.Method != tc.method {
				t.Errorf("expected %s; got %s", tc.method, r.Method)
			}
			if r.URL.String() != tc.url {
				t.Errorf("expected %s; got %s", tc.url, r.URL.String())
			}
			body, err := ioutil.ReadAll(r.Body)
			fatal(t, r.Body.Close())
			fatal(t, err)
			if string(body) != tc.message {
				t.Errorf("expected body to be:\n%s \n\ngot:\n%s", tc.message, body)
			}
			fmt.Fprint(w, tc.response)
		})
	}))
	defer ts.Close()

	bot, err := messenger.New(
		store,
		token,
		messenger.LogErr(log.New(os.Stderr, "", log.LstdFlags|log.Llongfile)),
		messenger.LogInfo(log.New(os.Stdout, "", log.LstdFlags)),
		messenger.FAPI(ts.URL),
	)

	// Send a message to Slangbrain.
	message := `{
		"entry": [
			{
				"messaging": [
					{
						"sender": {
							"id": "123"
						},
						"timestamp": 123,
						"message": {
							"text": "hi",
							"mid": "123"
						}
					}
				]
			}
		]
	}`

	w := httptest.NewRecorder()
	bot.ServeHTTP(w, httptest.NewRequest("POST", url, strings.NewReader(message)))
	body, err := ioutil.ReadAll(w.Result().Body)
	fatal(t, w.Result().Body.Close())
	fatal(t, err)
	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("expected status to be OK; got %v", w.Result().StatusCode)
	}
	if strings.TrimSpace(string(body)) != `{"status":"ok"}` {
		t.Errorf(`expected response to be {"status":"ok"}; got %s`, body)
	}
}

func fatal(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}
