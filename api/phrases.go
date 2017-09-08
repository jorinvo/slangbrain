package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/jorinvo/slangbrain/brain"
)

// Phrases returns a handler that implements GET and POST for / and DELETE and PUT for /:phraseid?token=:token
// For more see: https://slangbrain.com/api/
func Phrases(store brain.Store, errorLogger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			if err := r.Body.Close(); err != nil {
				errorLogger.Printf("method=%s; path=%s] failed closing body: %v", r.Method, r.URL.Path, err)
			}
		}()

		if r.URL.Path == "" {
			handlePhrases(store, errorLogger, w, r)
		} else {
			handlePhrase(store, errorLogger, w, r)
		}

	})
}

func handlePhrases(store brain.Store, errorLogger *log.Logger, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	id, ok := getID(store, errorLogger, w, r, true)
	if !ok {
		return
	}

	switch r.Method {
	case "GET":
		phrases, err := store.GetAllPhrases(id)
		if err != nil {
			errorLogger.Println(err)
			jsonError(w, "failed reading phrases", http.StatusInternalServerError)
			return
		}

		data := struct {
			Data []brain.IDPhrase `json:"data"`
		}{phrases}

		e := json.NewEncoder(w)
		e.SetIndent("", "  ")
		if err := e.Encode(data); err != nil {
			errorLogger.Printf("failed generating JSON for %d: %v", id, err)
			jsonError(w, "failed generating JSON", http.StatusInternalServerError)
		}

	case "POST":
		var data struct {
			Data []brain.Phrase `json:"data"`
		}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			jsonError(w, fmt.Sprintf("JSON is malformed: %v", err), http.StatusBadRequest)
			return
		}

		count, err := store.Import(id, data.Data)
		if err != nil {
			errorLogger.Printf("failed to add phrases <%#v> for %d: %v", data.Data, id, err)
			jsonError(w, "failed to add phrases", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, `{ "status": "ok", "count": "`+strconv.Itoa(count)+`" }`)

	default:
		jsonError(w, "unsupported method", http.StatusMethodNotAllowed)
	}
}

func handlePhrase(store brain.Store, errorLogger *log.Logger, w http.ResponseWriter, r *http.Request) {
	seq, err := strconv.Atoi(r.URL.Path)
	if err != nil {
		errorLogger.Printf("invalid phrase id '%s': %v", r.URL.Path, err)
		jsonError(w, "invalid phrase id", http.StatusBadRequest)
		return
	}

	id, ok := getID(store, errorLogger, w, r, true)
	if !ok {
		return
	}

	switch r.Method {
	case "PUT":
		var data struct {
			Data brain.Phrase `json:"data"`
		}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			jsonError(w, "failed to parse body", http.StatusBadRequest)
			return
		}
		if err := store.UpdatePhrase(id, seq, data.Data.Phrase, data.Data.Explanation); err != nil {
			errorLogger.Printf("failed to update phrase: %v", err)
			jsonError(w, "failed to update phrase", http.StatusInternalServerError)
			return
		}

	case "DELETE":
		if err := store.DeletePhrase(id, seq); err != nil {
			if err == brain.ErrNotFound {
				jsonError(w, "phrase does not exist", http.StatusNotFound)
				return
			}
			errorLogger.Printf("failed to delete phrase: %v", err)
			jsonError(w, "failed to delete phrase", http.StatusInternalServerError)
			return
		}

	default:
		jsonError(w, "unsupported method", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintln(w, `{ "status": "ok" }`)
}
