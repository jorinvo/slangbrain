package integration

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jorinvo/slangbrain/bot"
)

type profile struct{}

func (p profile) Name() string   { return "Martin" }
func (p profile) Locale() string { return "en_US" }
func (p profile) Timezone() int  { return 2 }

func TestProfileCache(t *testing.T) {
	store, cleanup := initDB(t)
	defer cleanup()

	// Set profile once, but it will be refetched becausee this one is expired already.
	fatal(t, store.SetProfile(123, profile{}, time.Now().Add(-60*24*time.Hour)))

	var b http.Handler

	tt := []testCase{
		{
			name:     "get profile 1",
			method:   "GET",
			url:      "/123?fields=first_name,locale,timezone&access_token=some-test-token&appsecret_proof=e5565c0a91022866f93ae581ad8e3bddca01e06c067b5816f0373fc76df3d1f0",
			response: `{ "first_name": "Martin" }`,
		},
		{
			name:   "welcome 1",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Hello Martin!\n\nWhenever you pick up a new phrase, just add it to your Slangbrain and remember it forever.\n\nYou save phrases from your everyday life in Slangbrain and Slangbrain will test your memories in a natural schedule."}}`,
		},
		{
			name:   "welcome 2",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Please send me a phrase and its explanation.\nSeparate them with a linebreak.\n\nIf your mother tongue is English and you're studying Spanish, a message would look like this:"}}`,
		},
		{
			name:   "welcome 3",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Hola\nHello"}}`,
		},
		{
			name:   "welcome 3",
			expect: `{"recipient":{"id":"123"},"message":{"text":"Now it's your turn:"}}`,
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

	b, _, err := bot.New(bot.Config{
		Store:       store,
		Token:       token,
		Secret:      secret,
		ErrLogger:   log.New(os.Stderr, "", log.LstdFlags|log.Llongfile),
		FacebookURL: ts.URL,
	})
	fatal(t, err)

	send(t, b, fmt.Sprintf(formatPayload, "PAYLOAD_GETSTARTED"))

	if state != len(tt) {
		t.Errorf("expected state to be %d; got %d", len(tt), state)
	}
}
