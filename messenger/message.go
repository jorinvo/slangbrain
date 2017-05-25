package messenger

import (
	"fmt"
	"strings"

	"github.com/jorinvo/messenger"
	"github.com/jorinvo/slangbrain/brain"
)

func (b bot) MessageHandler(m messenger.Message, r *messenger.Response) {
	if m.IsEcho {
		return
	}

	b.log.Println("message", m.QuickReply, m.Text)

	mode, err := b.store.GetMode(m.Sender.ID)
	if err != nil {
		b.log.Printf("failed to get mode for id %v: %v", m.Sender.ID, err)
		if err = r.Text(messageErr); err != nil {
			b.log.Println("failed to send message:", err)
		}
	}

	var fn func(messenger.Message) (string, []messenger.QuickReply, error)

	if m.QuickReply != nil && m.QuickReply.Payload == payloadStudy {
		fn = b.messageStartStudying
	} else if mode == brain.ModeStudy {
		fn = b.messageStudy
	} else if mode == brain.ModeAdd {
		fn = b.messageAdd
	}
	// TODO: else idle

	reply, buttons, err := fn(m)
	if err != nil {
		b.log.Println(err)
	}
	if err = r.TextWithReplies(reply, buttons); err != nil {
		b.log.Println("failed to send message:", err)
	}
}

func (b bot) messageStartStudying(m messenger.Message) (string, []messenger.QuickReply, error) {
	err := b.store.SetMode(m.Sender.ID, brain.ModeStudy)
	if err != nil {
		return messageErr, nil, fmt.Errorf("failed to set mode: %v", err)
	}
	return b.study(m.Sender.ID)
}

func (b bot) messageAdd(m messenger.Message) (string, []messenger.QuickReply, error) {
	parts := strings.SplitN(m.Text, "\n", 2)
	if len(parts) == 1 {
		return messageErrExplanation, nil, fmt.Errorf("explanation missing: %s", m.Text)
	}
	phrase := strings.TrimSpace(parts[0])
	explanation := strings.TrimSpace(parts[1])
	err := b.store.AddPhrase(m.Sender.ID, phrase, explanation)
	// TODO: keep user updated
	if err != nil {
		return messageErr, nil, fmt.Errorf("failed to save phrase: %v", err)
	}
	count, err := b.store.CountStudies(m.Sender.ID)
	var buttons []messenger.QuickReply
	if err == nil && count > 0 {
		buttons = []messenger.QuickReply{
			button(fmt.Sprintf("study (%d)", count), payloadStudy),
		}
	}
	// Fail silent when counting errors
	return "Phrase saved. Add next one.", buttons, err
}

func (b bot) messageStudy(m messenger.Message) (string, []messenger.QuickReply, error) {
	if m.QuickReply == nil {
		return "Currently only quick replies are supported.", nil, fmt.Errorf("not a quick reply: %v", m)
	}

	switch m.QuickReply.Payload {
	case payloadShow:
		study, err := b.store.GetStudy(m.Sender.ID)
		if err != nil {
			return messageErr, buttonsShow, fmt.Errorf("failed to show study: %v", err)
		}
		switch study.Mode {
		case brain.ButtonsExplanation:
			return study.Explanation, buttonsScore, nil
		default:
			return messageErr, nil, fmt.Errorf("SHOULD NEVER HAPPEN: unknown study mode: %v (%v)", study.Mode, study)
		}

	case payloadScoreBad:
		return b.scoreAndStudy(m.Sender.ID, brain.ScoreBad)
	case payloadScoreOk:
		return b.scoreAndStudy(m.Sender.ID, brain.ScoreOK)
	case payloadScoreGood:
		return b.scoreAndStudy(m.Sender.ID, brain.ScoreGood)
	default:
		return messageErr, nil, fmt.Errorf("unknown payload: %s", m.QuickReply.Payload)
	}
}
