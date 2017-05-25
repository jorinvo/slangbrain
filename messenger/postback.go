package messenger

import (
	"github.com/jorinvo/messenger"
	"github.com/jorinvo/slangbrain/brain"
)

func (b bot) PostbackHandler(p messenger.PostBack, r *messenger.Response) {
	var err error
	msg := messageErr
	var buttons []messenger.QuickReply

	b.log.Println("postback", p.Payload)

	switch p.Payload {
	case payloadGetStarted:
		err = b.store.SetMode(p.Sender.ID, brain.ModeAdd)
		if err != nil {
			b.log.Println("failed to set mode:", err)
			break
		}
		msg = messageWelcome
	case payloadAdd:
		err = b.store.SetMode(p.Sender.ID, brain.ModeAdd)
		if err != nil {
			b.log.Println("failed to set mode:", err)
			break
		}
		msg = messageAdd
	case payloadStudy:
		err = b.store.SetMode(p.Sender.ID, brain.ModeStudy)
		if err != nil {
			b.log.Println("failed to set mode:", err)
			break
		}
		msg, buttons, err = b.study(p.Sender.ID)
		if err != nil {
			b.log.Println("failed to study:", err)
		}
	}

	err = r.TextWithReplies(msg, buttons)
	if err != nil {
		b.log.Println("failed to send message:", err)
	}
}
