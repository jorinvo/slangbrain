package messenger

import (
	"regexp"

	"github.com/jorinvo/slangbrain/brain"
	"github.com/jorinvo/slangbrain/fbot"
	"github.com/jorinvo/slangbrain/user"
)

// Everything that is not in the unicode character classes
// for letters or numeric values
// See: http://www.fileformat.info/info/unicode/category/index.htm
var specialChars = regexp.MustCompile(`[^\p{Ll}\p{Lm}\p{Lo}\p{Lu}\p{Nd}\p{Nl}\p{No}\p{Mn}]`)

var inParantheses = regexp.MustCompile(`\(.*?\)`)

// HandleEvent handles a Messenger event.
func (b Bot) HandleEvent(e fbot.Event) {
	if e.Type == fbot.EventError {
		b.err.Println(e.Text)
		return
	}

	if e.Type == fbot.EventRead {
		if err := b.store.SetRead(e.ChatID, e.Time); err != nil {
			b.err.Println(err)
		}
		return
	}

	b.scheduleNotify(e.ChatID)

	u := user.Get(e.ChatID, b.store, b.err, b.content, b.client.GetProfile)

	if e.Type == fbot.EventPayload {
		b.handlePayload(u, e.Payload)
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

	if e.Type == fbot.EventMessage {
		b.handleMessage(u, e.Text)
		return
	}

	if e.Type == fbot.EventAttachment {
		b.handleAttachments(u, e.Attachments)
		return
	}

	b.err.Printf("unhandled event: %#v", e)
}
