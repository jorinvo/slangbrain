package webview

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/jorinvo/slangbrain/brain"
	"github.com/jorinvo/slangbrain/translate"
	"github.com/jorinvo/slangbrain/user"
)

// Webview can be used as an http.Handler to render the manager webview.
// Always use New() for initialization.
type Webview struct {
	store    brain.Store
	err      *log.Logger
	template *template.Template
	content  translate.Translator
	api      string
}

// New creates a new Webview.
func New(s brain.Store, errLog *log.Logger, t translate.Translator, api string) http.Handler {
	return Webview{
		store:    s,
		err:      errLog,
		template: template.Must(template.New("manage").Parse(html)),
		content:  t,
		api:      strings.TrimSuffix(api, "/") + "/phrases",
	}
}

// ServeHTTP handles a HTTP request by rendering the manager HTML page.
// Requires a token in the path to authenticate a user.
// ALso restricts IFrame usage to only Facebook domains.
func (view Webview) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Allow rendering inline on web
	allowFrom := []string{"https://www.messenger.com/", "https://www.facebook.com/"}
	for _, s := range allowFrom {
		if strings.HasPrefix(r.Referer(), s) {
			w.Header().Set("X-Frame-Options", "ALLOW-FROM "+s)
			break
		}
	}

	// Validate token
	token := r.URL.Path
	id, err := view.store.LookupToken(token)
	if err != nil {
		view.err.Printf("failed looking up token '%s': %v", token, err)
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	// Get phrases, get localized content and render template
	phrases, err := view.store.GetAllPhrases(id)
	if err != nil {
		view.err.Println(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	u := user.Get(id, view.store, view.err, view.content, nil)
	data := struct {
		Phrases map[int64]brain.Phrase
		Label   translate.Web
		API     string
		Token   string
	}{phrases, u.Web, view.api, token}
	if err := view.template.Execute(w, data); err != nil {
		view.err.Printf("failed to render template: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}
