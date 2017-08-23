package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/jorinvo/slangbrain/brain"
)

// Phrases returns a handler that implements DELETE and PUT for /:phraseid?token=:token
func Phrases(store brain.Store, errorLogger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seq, err := strconv.Atoi(r.URL.Path)
		if err != nil {
			errorLogger.Printf("invalid phrase id '%s': %v", r.URL.Path, err)
			http.Error(w, "invalid phrase id", http.StatusBadRequest)
			return
		}

		id, ok := getID(store, errorLogger, w, r)
		if !ok {
			return
		}

		switch r.Method {
		case http.MethodPut:
			p := struct{ Phrase, Explanation string }{}
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				errorLogger.Printf("failed to parse body: %v", err)
				http.Error(w, "failed to parse body", http.StatusBadRequest)
				return
			}
			if err := store.UpdatePhrase(id, seq, p.Phrase, p.Explanation); err != nil {
				errorLogger.Printf("failed to update phrase: %v", err)
				http.Error(w, "failed to update phrase", http.StatusInternalServerError)
				return
			}

		case http.MethodDelete:
			if err := store.DeletePhrase(id, seq); err != nil {
				errorLogger.Printf("failed to delete phrase: %v", err)
				http.Error(w, "failed to delete phrase", http.StatusInternalServerError)
				return
			}

		default:
			http.Error(w, "unsupported method", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprintln(w, `{"status":"ok"}`)
	})
}
