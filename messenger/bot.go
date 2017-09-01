package messenger

import (
	"net/url"

	"github.com/jorinvo/slangbrain/brain"
	"github.com/jorinvo/slangbrain/fbot"
	"github.com/jorinvo/slangbrain/user"
)

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

	u := user.Get(e.ChatID, b.store, b.err, b.content, b.client.GetProfile)

	if e.Type == fbot.EventReferral {
		ref, err := url.QueryUnescape(e.Ref)
		if err != nil {
			b.err.Printf("[id=%d] failed to unescape ref %s: %v", u.ID, e.Ref, err)
			return
		}
		if links := getLinks(ref); links != nil {
			b.handleLinks(u, links)
			return
		}
		b.err.Printf("[id=%d] got unhandled ref: %s", u.ID, e.Ref)
		return
	}

	if e.Type == fbot.EventPayload {
		b.handlePayload(u, e.Payload, e.Ref)
		b.scheduleNotify(e.ChatID)
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
		b.scheduleNotify(e.ChatID)
		return
	}

	if e.Type == fbot.EventAttachment {
		b.handleAttachments(u, e.Attachments)
		return
	}

	b.err.Printf("unhandled event: %#v", e)
}
