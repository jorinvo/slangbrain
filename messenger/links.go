package messenger

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/jorinvo/slangbrain/brain"
	"github.com/jorinvo/slangbrain/user"
)

// Helper to handle links sent to Slangbrain.
// It doesn't matter if the links come from file uploads, sharing, referral links or inside messages.
//
// For now we only handle CSV files.
// For all other links admins are notified to look into them manually.
func (b Bot) handleLinks(u user.User, links []string) {
	// Collect CSV files from links
	var csvFiles []struct{ Name, URL string }
	for _, link := range links {
		// Parse URL
		f, err := url.ParseRequestURI(link)
		if err != nil {
			b.err.Printf("[id=%d,url=%s] failed to parse URL: %v", u.ID, link, err)
			continue
		}

		// Notify admin for non-csv files
		if strings.ToLower(path.Ext(f.Path)) != ".csv" {
			if b.feedback != nil {
				b.feedback <- Feedback{ChatID: u.ID, Username: u.Name(), Message: fmt.Sprintf("[unhandled link: %s]", link)}
			} else {
				b.err.Printf("[id=%d] unhandled link from %s: %s", u.ID, u.Name(), link)
			}
			continue
		}

		csvFiles = append(csvFiles, struct{ Name, URL string }{path.Base(f.Path), link})
	}

	// Go back to menu mode in any case
	if err := b.store.SetMode(u.ID, brain.ModeMenu); err != nil {
		b.err.Println(err)
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
		if cols := len(records[0]); cols < 2 {
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
