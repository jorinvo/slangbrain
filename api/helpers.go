package api

import (
	"log"
	"net/http"

	"github.com/jorinvo/slangbrain/brain"
)

// Get user id from token in request query, otherwise fail+log as unauthorized.
func getID(store brain.Store, errorLogger *log.Logger, w http.ResponseWriter, r *http.Request) (int64, bool) {
	token := r.URL.Query().Get("token")
	id, err := store.LookupToken(token)
	if err != nil {
		errorLogger.Printf("invalid token '%s': %v", token, err)
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return id, false
	}
	return id, true
}
