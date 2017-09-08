package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/jorinvo/slangbrain/brain"
)

// Get user id from token in request query, otherwise fail+log as unauthorized.
func getID(store brain.Store, errorLogger *log.Logger, w http.ResponseWriter, r *http.Request, isJSON bool) (int64, bool) {
	token := r.URL.Query().Get("token")
	id, err := store.LookupToken(token)
	if err != nil {
		errorLogger.Printf("invalid token '%s': %v", token, err)
		if isJSON {
			jsonError(w, "invalid token", http.StatusUnauthorized)
		} else {
			http.Error(w, "invalid token", http.StatusUnauthorized)
		}
		return id, false
	}
	return id, true
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	fmt.Fprintln(w, `{ "error": "`+msg+`" }`)
}
