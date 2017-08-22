package messenger

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/jorinvo/slangbrain/brain"
	"github.com/jorinvo/slangbrain/fbot"
	"github.com/jorinvo/slangbrain/user"
)

// Handle the upload of CSV files to import phrases.
// Other attachments are handled only by notifying the admin to look into them manually.
func (b Bot) handleAttachments(u user.User, attachments []fbot.Attachment) {
	var csvFiles []struct{ Name, URL string }
	for _, a := range attachments {
		// Notify admin for non-file attachments
		if a.Type != "file" {
			// Ignore stickers for now, 'like' button is used, but ignored.
			if a.Sticker != 0 {
				continue
			}
			if b.feedback != nil {
				b.feedback <- Feedback{ChatID: u.ID, Username: u.Name(), Message: "[user sent unhandled attachment of type '" + a.Type + "']"}
			} else {
				b.err.Printf("got unhandled attachment (%s) from %s (%d)", a.Type, u.Name(), u.ID)
			}
			continue
		}

		// Notify admin for non-csv files
		f, err := url.ParseRequestURI(a.URL)
		if err != nil {
			b.send(u.ID, u.Msg.Error, nil, fmt.Errorf("failed to parse URL of file at %s (user %d): %v", a.URL, u.ID, err))
			return
		}
		if strings.ToLower(path.Ext(f.Path)) != ".csv" {
			if b.feedback != nil {
				b.feedback <- Feedback{ChatID: u.ID, Username: u.Name(), Message: "[unhandled file: " + a.URL + "]"}
			} else {
				b.err.Printf("got unhandled file from %s (%d): %s", u.Name(), u.ID, a.URL)
			}
			continue
		}

		csvFiles = append(csvFiles, struct{ Name, URL string }{path.Base(f.Path), a.URL})
	}

	if len(csvFiles) == 0 {
		b.send(u.ID, u.Msg.Menu, u.Rpl.MenuMode, nil)
		return
	}

	var allRecords [][]string
	for _, file := range csvFiles {
		// Get contents and parse CSV
		req, err := http.NewRequest("GET", file.URL, nil)
		if err != nil {
			b.send(u.ID, u.Msg.Error, nil, fmt.Errorf("failed to create request to file %s (user %d): %v", file.URL, u.ID, err))
			return
		}

		res, err := b.do(req)
		if err != nil {
			b.send(u.ID, u.Msg.Error, nil, fmt.Errorf("failed to get file %s (user %d): %v", file.URL, u.ID, err))
			return
		}
		defer func(f string) {
			if err := res.Body.Close(); err != nil {
				b.err.Printf("failed to close body for request to %s (user %d): %v", f, u.ID, err)
			}
		}(file.URL)

		records, err := csv.NewReader(res.Body).ReadAll()
		if err != nil {
			b.send(u.ID, fmt.Sprintf(u.Msg.ImportErrParse, file.Name, err), nil, fmt.Errorf("failed to parse CSV file %s (user %d): %v", file.URL, u.ID, err))
			return
		}

		// Check CSV formatting
		if len(records) == 0 {
			continue
		}
		if cols := len(records[0]); cols != 2 {
			b.send(u.ID, fmt.Sprintf(u.Msg.ImportErrCols, file.Name, cols), nil, nil)
			return
		}

		allRecords = append(allRecords, records...)
	}

	if len(allRecords) < 1 {
		b.send(u.ID, u.Msg.ImportEmpty, nil, nil)
		return
	}

	// Prevent duplicates
	var phrases []brain.Phrase
	for _, r := range allRecords {
		p := brain.Phrase{
			Phrase:      strings.TrimSpace(r[0]),
			Explanation: strings.TrimSpace(r[1]),
		}
		for _, prev := range phrases {
			if p.Explanation == prev.Explanation {
				b.send(u.ID, fmt.Sprintf(u.Msg.ImportErrDuplicate, p.Explanation), nil, nil)
				return
			}
		}
		phrases = append(phrases, p)
	}

	// Queue import and ask user for confirmation
	valid, existing, err := b.store.QueueImport(u.ID, phrases)
	if err != nil {
		b.send(u.ID, u.Msg.Error, u.Rpl.MenuMode, err)
		return
	}
	if existing == 0 {
		msg := fmt.Sprintf(u.Msg.ImportPrompt, valid)
		b.send(u.ID, msg, u.Rpl.Import, nil)
	} else if valid == 0 {
		msg := fmt.Sprintf(u.Msg.ImportNone+"\n\n"+u.Msg.Menu, existing)
		b.send(u.ID, msg, u.Rpl.MenuMode, nil)
	} else {
		msg := fmt.Sprintf(u.Msg.ImportPromptIgnore, valid, existing)
		b.send(u.ID, msg, u.Rpl.Import, nil)
	}
}
