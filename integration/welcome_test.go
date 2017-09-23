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
			name:     "greeting",
			url:      "/me/messenger_profile?access_token=some-test-token&appsecret_proof=e5565c0a91022866f93ae581ad8e3bddca01e06c067b5816f0373fc76df3d1f0",
			expect:   `{"greeting":[{"locale":"default","text":"Hi {{user_first_name}}! Slangbrain helps you with our language studies.\nMaster the language you encounter daily instead of limiting yourself to a textbook."},{"locale":"de_DE","text":"Hi {{user_first_name}}! Mit Slangbrain kannst du Sprache lernen wie sie dir im Alltag begegnet statt ein Schulbuch auswendig zu lernen."},{"locale":"en_GB","text":"Hi {{user_first_name}}! Slangbrain helps you with our language studies.\nMaster the language you encounter daily instead of limiting yourself to a textbook."},{"locale":"en_US","text":"Hi {{user_first_name}}! Slangbrain helps you with our language studies.\nMaster the language you encounter daily instead of limiting yourself to a textbook."}]}`,
			response: `{"result":"success"}`,
		},
		{
			name:     "get started button",
			url:      "/me/messenger_profile?access_token=some-test-token&appsecret_proof=e5565c0a91022866f93ae581ad8e3bddca01e06c067b5816f0373fc76df3d1f0",
			expect:   `{"get_started":{"payload":"PAYLOAD_GETSTARTED"}}`,
			response: `{"result":"success"}`,
		},
		{
			name:     "get profile",
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

	bot, err := messenger.New(
		store,
		token,
		secret,
		messenger.Setup,
		messenger.LogErr(log.New(os.Stderr, "", log.LstdFlags|log.Llongfile)),
		messenger.FAPI(ts.URL),
	)
	fatal(t, err)

	send(t, bot, fmt.Sprintf(formatPayload, "PAYLOAD_GETSTARTED"))

	if state != len(tt) {
		t.Errorf("expected state to be %d; got %d", len(tt), state)
	}
}
