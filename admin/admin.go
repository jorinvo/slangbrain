package admin

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/jorinvo/slangbrain/brain"
)

// New ...
func New(store brain.Store, errorLogger *log.Logger) http.Handler {
	return admin{store, errorLogger}
}

type admin struct {
	store brain.Store
	err   *log.Logger
}

func (a admin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	}
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
