package messenger

import (
	"regexp"

	"github.com/jorinvo/slangbrain/brain"
	"github.com/jorinvo/slangbrain/common"
	"github.com/jorinvo/slangbrain/fbot"
	"github.com/jorinvo/slangbrain/translate"
)

// user stores relevent information for the current request.
// It is used for handling messages and payloads.
// A profile and content in the correct language are loaded.
type user struct {
	ID int64
	common.Profile
	translate.Content
}

// Everything that is not in the unicode character classes
// for letters or numeric values
// See: http://www.fileformat.info/info/unicode/category/index.htm
var specialChars = regexp.MustCompile(`[^\p{Ll}\p{Lm}\p{Lo}\p{Lu}\p{Nd}\p{Nl}\p{No}]`)

var inParantheses = regexp.MustCompile(`\(.*?\)`)

// HandleEvent handles a Messenger event.
func (b Bot) HandleEvent(e fbot.Event) {
	if e.Type == fbot.EventError {
		b.err.Println(e.Text)
		return
	}

	if e.Type == fbot.EventUnknown {
		b.err.Println("received unknown event:", e)
		return
	}

	if e.Type == fbot.EventRead {
		if err := b.store.SetRead(e.ChatID, e.Time); err != nil {
			b.err.Println(err)
		}
		return
	}

	b.scheduleNotify(e.ChatID)

	if e.Type == fbot.EventPayload {
		b.handlePayload(b.getUser(e.ChatID), e.Payload)
		return
	}

	if err := b.store.QueueMessage(e.MessageID); err != nil {
		if err == brain.ErrExists {
			b.info.Printf("Message already processed: %v", e.MessageID)
			return
		}
		b.err.Println("failed to save message ID:", err)
		return
	}

	b.handleMessage(b.getUser(e.ChatID), e.Text)
}
