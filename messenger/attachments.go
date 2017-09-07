package messenger

import (
	"fmt"
	"net/url"

	"github.com/jorinvo/slangbrain/fbot"
	"github.com/jorinvo/slangbrain/user"
)

// Handle the upload of CSV files to import phrases.
// Other attachments are handled only by notifying the admin to look into them manually.
func (b Bot) handleAttachments(u user.User, attachments []fbot.Attachment) {
	var links []string
	for _, a := range attachments {
		// Support sharing CSV files to Slangbrain:
		// Extract fallback URL and treat it like an uploaded file
		if a.Type == "fallback" {
			f, err := url.ParseRequestURI(a.URL)
			if err != nil {
				b.err.Printf("[id=%d,url=%s] failed to parse fallback URL: %v", u.ID, a.URL, err)
				continue
			}
			a.URL, err = url.QueryUnescape(f.Query().Get("u"))
			if err != nil {
				b.err.Printf("[id=%d,url=%s] failed to unescape fallback URL: %v", u.ID, a.URL, err)
				continue
			}
			links = append(links, a.URL)
			continue
		}

		// Ignore stickers for now, since 'like' button is sent a lot
		if a.Sticker != 0 {
			continue
		}

		// Notify admin for non-file and non-fallback attachments
		if a.Type != "file" {
			b.feedback <- Feedback{
				ChatID:   u.ID,
				Username: u.Name(),
				Message:  fmt.Sprintf("[ user sent unhandled '%s' (sticker %d): %s ]", a.Type, a.Sticker, a.URL),
				Channel:  slackUnhandled,
			}

			continue
		}

		links = append(links, a.URL)
	}

	b.handleLinks(u, links)
}
