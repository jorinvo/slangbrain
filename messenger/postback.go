package messenger

import (
	"fmt"

	"github.com/jorinvo/messenger"
	"github.com/jorinvo/slangbrain/brain"
	"github.com/pkg/errors"
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

func (b bot) study(chatID int64) (string, []messenger.QuickReply, error) {
	study, err := b.store.GetStudy(chatID)
	if err != nil {
		return messageErr, nil, errors.Wrapf(err, "failed to study with id %v", chatID)
	}
	if study.Total == 0 {
		return messageStudyDone, nil, err
	}
	switch study.Mode {
	case brain.ButtonsExplanation:
		return fmt.Sprintf(messageButtons, study.Phrase), buttonsShow, err
	default:
		return messageErr, nil, errors.Wrapf(err, "unknown study mode %v", study.Mode)
	}
}

func (b bot) scoreAndStudy(chatID int64, score brain.Score) (string, []messenger.QuickReply, error) {
	err := b.store.ScoreStudy(chatID, score)
	if err != nil {
		return messageErr, nil, errors.Wrapf(err, "failed to score study with id %v", chatID)
	}
	return b.study(chatID)
}
