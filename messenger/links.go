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

// Wrapper around extractPhrases that queues found phrases and sends an appropriate answer to the user.
func (b Bot) handleLinks(u user.User, links []string) {
	// Go back to menu mode in any case
	if err := b.store.SetMode(u.ID, brain.ModeMenu); err != nil {
		b.err.Println(err)
	}

	var queued int
	phrases, files, userErr, err := b.extractPhrases(u, links)
	if err == nil && userErr == "" {
		queued, err = b.store.QueueImport(u.ID, phrases)
	}
	if err != nil || userErr != "" || files == "" {
		if userErr == "" {
			userErr = u.Msg.Menu
		}
		b.send(u.ID, userErr, u.Rpl.MenuMode, err)
		return
	}

	// Ask for confirmation
	existing := len(phrases) - queued
	if existing == 0 {
		msg := fmt.Sprintf(u.Msg.ImportPrompt, queued, files)
		b.send(u.ID, msg, u.Rpl.Import, nil)
	} else if queued == 0 {
		msg := fmt.Sprintf(u.Msg.ImportNone+"\n\n"+u.Msg.Menu, existing, files)
		b.send(u.ID, msg, u.Rpl.MenuMode, nil)
	} else {
		msg := fmt.Sprintf(u.Msg.ImportPromptIgnore, queued, existing, files)
		b.send(u.ID, msg, u.Rpl.Import, nil)
	}
}

// Helper to handle links sent to Slangbrain.
// It doesn't matter if the links come from file uploads, sharing, referral links or inside messages.
//
// This function doesn't send anything back to the user,
// but returns information that should be used to generate a reply.
//
// For now we only handle .csv, .txt and .tsv files.
// .txt and .tsv files are handled with the CSV parser, only that they use a tab instead of a comma as separator.
// For all other links admins are notified to look into them manually.
//
// Returns a list of phrases that afterwards can be queued or imported directly,
// a formatted list of the imported files (f.e. "Japanese.csv, German.txt and English.tsv"),
// an error message to be sent to the user
// and a possible application error that needs to be handled.
//
// It is possible to have user error but no application error and also the other way around.
func (b Bot) extractPhrases(u user.User, links []string) ([]brain.Phrase, string, string, error) {
	// Collect files from links
	var files []struct{ Name, URL, Ext string }
	for _, link := range links {
		// Parse URL
		f, err := url.ParseRequestURI(link)
		if err != nil {
			b.err.Printf("[id=%d,url=%s] failed to parse URL: %v", u.ID, link, err)
			continue
		}

		// Notify admin for unsupported files
		ext := strings.ToLower(path.Ext(f.Path))
		if ext != ".csv" && ext != ".txt" && ext != ".tsv" {
			if b.feedback != nil {
				b.feedback <- Feedback{ChatID: u.ID, Username: u.Name(), Message: fmt.Sprintf("[unhandled link: %s]", link)}
			} else {
				b.err.Printf("[id=%d] unhandled link from %s: %s", u.ID, u.Name(), link)
			}
			continue
		}

		files = append(files, struct{ Name, URL, Ext string }{path.Base(f.Path), link, ext})
	}

	if len(files) == 0 {
		return nil, "", "", nil
	}

	// Get contents and parse files
	var allRecords [][]string
	var fileNames []string
	for _, file := range files {
		req, err := http.NewRequest("GET", file.URL, nil)
		if err != nil {
			return nil, "", "", fmt.Errorf("failed to create request to file %s: %v", file.URL, err)
		}

		res, err := b.do(req)
		if err != nil {
			return nil, "", "", fmt.Errorf("failed to get file %s: %v", file.URL, err)
		}
		defer func(f string) {
			if err := res.Body.Close(); err != nil {
				b.err.Printf("failed to close body for request to %s: %v", f, err)
			}
		}(file.URL)

		// Separate .tsv and .txt files by tab
		csvReader := csv.NewReader(res.Body)
		if file.Ext == "tsv" || file.Ext == ".txt" {
			csvReader.Comma = '\t'
		}

		records, err := csvReader.ReadAll()
		if err != nil {
			msg := fmt.Sprintf(u.Msg.ImportErrParse, file.Name, err)
			return nil, "", msg, fmt.Errorf("failed to parse file %s: %v", file.URL, err)
		}

		if len(records) == 0 {
			continue
		}
		// Check formatting
		if cols := len(records[0]); cols < 2 {
			return nil, "", fmt.Sprintf(u.Msg.ImportErrCols, file.Name, cols), nil
		}

		fileNames = append(fileNames, file.Name)
		allRecords = append(allRecords, records...)
	}

	if len(allRecords) == 0 {
		return nil, "", u.Msg.ImportEmpty, nil
	}

	// Check for duplicates
	var phrases []brain.Phrase
	for _, r := range allRecords {
		p := brain.Phrase{
			Phrase:      strings.TrimSpace(r[0]),
			Explanation: strings.TrimSpace(r[1]),
		}

		// Merge, if duplicate
		for _, prev := range phrases {
			if p.Explanation == prev.Explanation {
				prev.Phrase += " (" + p.Phrase + ")"
				continue
			}
		}

		phrases = append(phrases, p)
	}

	// List file names
	fileList := fileNames[0]
	if l := len(fileNames); l > 1 {
		fileList = strings.Join(fileNames[:l-1], ", ") + " " + u.Msg.And + " " + fileNames[l-1]
	}

	// Queue import
	return phrases, fileList, "", nil
}
