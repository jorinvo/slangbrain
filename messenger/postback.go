package messenger

import (
	"github.com/jorinvo/messenger"
	"github.com/jorinvo/slangbrain/brain"
)

// Only handling the Get Started button here
func (b bot) PostbackHandler(p messenger.PostBack, r *messenger.Response) {
	b.log.Println("postback", p.Payload)

	if p.Payload != payloadGetStarted {
		return
	}

	err := b.store.SetMode(p.Sender.ID, brain.ModeAdd)
	if err != nil {
		b.log.Println("failed to set mode:", err)
	}

	err = r.Text(messageWelcome)
	if err != nil {
		b.log.Println("failed to send message:", err)
	}
}
