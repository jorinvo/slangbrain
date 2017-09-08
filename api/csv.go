package api

import (
	"encoding/csv"
	"log"
	"net/http"

	"github.com/jorinvo/slangbrain/brain"
)

// CSV returns a handler that implements GET to return a CSV containing all phrases of the authenticated user.
func CSV(store brain.Store, errorLogger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := getID(store, errorLogger, w, r)
		if !ok {
			return
		}

		if r.Method != "GET" {
			http.Error(w, "unsupported method", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "text/csv; charset=utf-8")

		// Get phrases and write them as CSV to the HTTP resonse
		csvW := csv.NewWriter(w)

		phrases, err := store.GetAllPhrases(id)
		if err != nil {
			errorLogger.Println(err)
			http.Error(w, "failed reading phrases", http.StatusInternalServerError)
			return
		}

		for _, p := range phrases {
			if err := csvW.Write([]string{p.Phrase, p.Explanation}); err != nil {
				errorLogger.Printf("failed generating CSV file for %d, at phrase '%s': %v", id, p.Phrase, err)
				http.Error(w, "failed generating CSV file", http.StatusInternalServerError)
			}
		}

		csvW.Flush()
		if err := csvW.Error(); err != nil {
			errorLogger.Printf("failed generating CSV file for %d: %v", id, err)
			http.Error(w, "failed generating CSV file", http.StatusInternalServerError)
		}
	})
}
