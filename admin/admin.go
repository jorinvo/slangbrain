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
	"github.com/jorinvo/slangbrain/fbot"
)

// New ...
func New(store brain.Store, slackHook, slackToken string, errorLogger *log.Logger, client fbot.Client) Admin {
	return Admin{store, errorLogger, slackHook, slackToken, client}
}

// Admin ...
type Admin struct {
	store      brain.Store
	err        *log.Logger
	slackHook  string
	slackToken string
	client     fbot.Client
}

func (a Admin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := r.Body.Close(); err != nil {
			a.err.Println(err)
		}
	}()

	switch r.URL.Path {
	case "/backup":
		if r.Method != "GET" {
			return
		}
		a.store.BackupTo(w)
	case "/studynow":
		if r.Method != "GET" {
			return
		}
		if err := a.store.StudyNow(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, "studies updated")
	case "/slack":
		if r.FormValue("token") != a.slackToken {
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
		if err := a.client.Send(int64(id), msg, nil); err != nil {
			slackError(w, err)
			return
		}
	}
}

func slackError(w http.ResponseWriter, err error) {
	fmt.Fprint(w, fmt.Sprintf(`{ "text": "Error sending message: %s." }`, err))
}

// HandleMessage ...
func (a Admin) HandleMessage(id int64, name, msg string) error {
	tmpl := `{
		"username": "%s",
		"text": "%d\n\n%s"
	}`
	resp, err := http.Post(a.slackHook, "application/json", strings.NewReader(fmt.Sprintf(tmpl, name, id, msg)))
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
