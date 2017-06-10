// Package admin provides an admin server that can be used to make backups
// and to communicate with users via Slack.
package admin

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/jorinvo/slangbrain/brain"
)

// Admin is a HTTP handler that can be used for backups
// and to communicate with users via Slack.
type Admin struct {
	Store        brain.Store
	Err          *log.Logger
	SlackHook    string
	SlackToken   string
	ReplyHandler func(int64, string) error
}

// ServeHTTP serves the different endpoints the admin server provides.
func (a Admin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := r.Body.Close(); err != nil {
			a.Err.Println(err)
		}
	}()

	switch r.URL.Path {
	case "/backup":
		if r.Method != "GET" {
			return
		}
		a.Store.BackupTo(w)

	case "/phrase":
		if r.Method != "DELETE" {
			return
		}
		qChatID := r.URL.Query().Get("chatid")
		hasChatID := qChatID != ""
		phrase := r.URL.Query().Get("phrase")
		hasPhrase := phrase != ""
		explanation := r.URL.Query().Get("explanation")
		hasExplanation := explanation != ""
		score := r.URL.Query().Get("score")
		hasScore := score != ""
		if !(hasChatID || hasPhrase || hasExplanation || hasScore) {
			http.Error(w, "no query specified", 400)
			return
		}
		var chatID int64
		var err error
		if hasChatID {
			chatID, err = strconv.ParseInt(qChatID, 10, 64)
			if err != nil {
				http.Error(w, fmt.Sprintf("invalid chatid: '%s'", qChatID), 400)
				return
			}
		}
		count, err := a.Store.DeletePhrases(func(id int64, p brain.Phrase) bool {
			if hasChatID && id != chatID {
				return false
			}
			if hasPhrase && !strings.Contains(p.Phrase, phrase) {
				return false
			}
			if hasExplanation && !strings.Contains(p.Explanation, explanation) {
				return false
			}
			if hasScore && strconv.Itoa(int(p.Score)) != score {
				return false
			}
			return true
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("failed deleting phrases: %v", err), 500)
			return
		}
		fmt.Fprintf(w, "Deleted %d phrases.", count)

	case "/studynow":
		if r.Method != "GET" {
			return
		}
		if err := a.Store.StudyNow(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, "studies updated")

	case "/slack":
		if r.FormValue("token") != a.SlackToken {
			slackError(w, fmt.Errorf("invalid token"))
			return
		}
		if r.FormValue("bot_id") != "" {
			return
		}
		text := strings.TrimSpace(r.FormValue("text"))
		fields := strings.Fields(text)
		if len(fields) < 2 {
			slackError(w, fmt.Errorf("missing ID"))
			return
		}
		firstField := fields[0]
		id, err := strconv.Atoi(firstField)
		if err != nil {
			slackError(w, fmt.Errorf("failed parsing ID: %v", err))
			return
		}
		msg := strings.TrimSpace(strings.TrimPrefix(text, firstField))
		if err := a.ReplyHandler(int64(id), msg); err != nil {
			slackError(w, err)
			return
		}
	}
}

func slackError(w http.ResponseWriter, err error) {
	fmt.Fprint(w, fmt.Sprintf(`{ "text": "Error sending message: %s." }`, err))
}

// HandleMessage can be called to send user message to Slack.
func (a Admin) HandleMessage(id int64, name, msg string) error {
	tmpl := `{
		"username": "%s",
		"text": "%d\n\n%s"
	}`
	resp, err := http.Post(a.SlackHook, "application/json", strings.NewReader(fmt.Sprintf(tmpl, name, id, msg)))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("HTTP status code is not OK: %s", body)
	}
	return nil
}

func csvImport(store brain.Store, errLogger, infoLogger *log.Logger, toImport string) {
	// CSV import
	parts := strings.Split(toImport, ":")
	i, err := strconv.Atoi(parts[0])
	if err != nil {
		errLogger.Fatal(err)
	}
	chatID := int64(i)
	file := parts[1]
	infoLogger.Printf("Importing to chat ID %d from CSV file %s", chatID, file)
	f, err := os.Open(file)
	if err != nil {
		errLogger.Fatalln(err)
	}
	count := 0
	reader := csv.NewReader(f)
	for {
		row, err := reader.Read()
		if err == io.EOF {
			infoLogger.Printf("Imported %d phrases", count)
			return
		}
		if err != nil {
			errLogger.Fatalln(err)
		}
		if len(row) != 2 {
			errLogger.Printf("line %d has wrong number of fields, expected 2, had %d", count+1, len(row))
		} else {
			count++
			p := strings.TrimSpace(row[0])
			e := strings.TrimSpace(row[1])
			if err = store.AddPhrase(chatID, p, e); err != nil {
				errLogger.Fatalln(err)
			}
		}
	}
}
