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

func TestWelcome(t *testing.T) {
	store, cleanup := initDB(t)
	defer cleanup()

	var bot messenger.Bot

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
		},
	}

	// Track state to make sure responses are in order.
	state := 0

	// Fake the Facebook server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		checkCase(t, w, r, tt[state])
		state++
	}))
	defer ts.Close()

	bot, err := messenger.New(
		store,
		token,
		messenger.LogErr(log.New(os.Stderr, "", log.LstdFlags|log.Llongfile)),
		messenger.FAPI(ts.URL),
	)
	fatal(t, err)

	send(t, bot, fmt.Sprintf(formatMessage, "1", "hi"))

	if state != len(tt) {
		t.Errorf("expected state to be %d; got %d", len(tt), state)
	}
}
