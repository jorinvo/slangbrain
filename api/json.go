package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/jorinvo/slangbrain/brain"
)

// JSON returns a handler that implements GET and POST methods as specified here:
// https://slangbrain.com/api/
func JSON(store brain.Store, errorLogger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := getID(store, errorLogger, w, r)
		if !ok {
			return
		}

		switch r.Method {
		case "GET":
			w.Header().Set("Content-Type", "application/json; charset=utf-8")

			phrases, err := store.GetAllPhrases(id)
			if err != nil {
				errorLogger.Println(err)
				jsonError(w, "failed reading phrases", http.StatusInternalServerError)
				return
			}

			e := json.NewEncoder(w)
			e.SetIndent("", "  ")
			if err := e.Encode(phrases); err != nil {
				errorLogger.Printf("failed generating JSON for %d: %v", id, err)
				jsonError(w, "failed generating JSON", http.StatusInternalServerError)
			}

		default:
			jsonError(w, "unsupported method", http.StatusMethodNotAllowed)
		}
	})
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	http.Error(w, `{ "error": "`+msg+`" }`, code)
}
