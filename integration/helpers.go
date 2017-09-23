package integration

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jorinvo/slangbrain/brain"
)

const (
	appURL        = "https://api.slangbrain.com/"
	token         = "some-test-token"
	secret        = "some-test-secret"
	defaultMethod = "POST"
	defaultURI    = "/me/messages?access_token=some-test-token&appsecret_proof=e5565c0a91022866f93ae581ad8e3bddca01e06c067b5816f0373fc76df3d1f0"
)

const formatMessage = `
	{
		"entry": [
			{
				"messaging": [
					{
						"sender": {
							"id": "123"
						},
						"timestamp": 0,
						"message": {
							"mid": "%s",
							"text": "%s"
						}
					}
				]
			}
		]
	}
`

const formatPayload = `
	{
		"entry": [
			{
				"messaging": [
					{
						"sender": {
							"id": "123"
						},
						"timestamp": 0,
						"postback": {
							"payload": "%s"
						}
					}
				]
			}
		]
	}
`

// testCase describes a message that the Facebook Server will receive and a response it will send to Slangbrain.
type testCase struct {
	name     string
	url      string
	method   string
	expect   string
	response string
	send     string
}

func fatal(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

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

func checkCase(t *testing.T, w http.ResponseWriter, r *http.Request, tc testCase) {
	t.Run(tc.name, func(t *testing.T) {
		// Check method
		if noMatchNorDefault(r.Method, tc.method, defaultMethod) {
			t.Errorf("expected %s; got %s", tc.method, r.Method)
		}
		// Check URL
		if noMatchNorDefault(r.URL.String(), tc.url, defaultURI) {
			t.Errorf("expected %s; got %s", tc.url, r.URL.String())
		}
		// Check body
		body, err := ioutil.ReadAll(r.Body)
		fatal(t, r.Body.Close())
		fatal(t, err)
		if string(body) != tc.expect {
			t.Errorf("expected body to be:\n%s \n\ngot:\n%s", tc.expect, body)
		}
		// Respond
		fmt.Fprint(w, tc.response)
	})
}

func noMatchNorDefault(val, expect, defaultVal string) bool {
	return (expect != "" && val != expect) || (expect == "" && val != defaultVal)
}

// Send a message to Slangbrain.
func send(t *testing.T, handler http.Handler, message string) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", appURL, strings.NewReader(message))

	// Add signature header for authentication
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write([]byte(message))
	req.Header.Set("X-Hub-Signature", "sha1="+hex.EncodeToString(mac.Sum(nil)))

	handler.ServeHTTP(w, req)
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
