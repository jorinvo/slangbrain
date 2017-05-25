package messenger

import (
	"strings"

	"github.com/jorinvo/messenger"
	"github.com/jorinvo/slangbrain/brain"
	"github.com/pkg/errors"
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

	switch mode {
	case brain.ModeStudy:
		reply, buttons, err := b.messageStudy(m)
		if err != nil {
			b.log.Println(err)
		}
		if err = r.TextWithReplies(reply, buttons); err != nil {
			b.log.Println("failed to send message:", err)
		}

	case brain.ModeAdd:
		reply, err := b.messageAdd(m)
		if err != nil {
			b.log.Println(err)
		}
		if err = r.Text(reply); err != nil {
			b.log.Println("failed to send message:", err)
		}
	}
}

func (b bot) messageAdd(m messenger.Message) (string, error) {
	parts := strings.SplitN(m.Text, "\n", 2)
	if len(parts) == 1 {
		return messageErrExplanation, errors.Errorf("explanation missing: %s", m.Text)
	}
	phrase := strings.TrimSpace(parts[0])
	explanation := strings.TrimSpace(parts[1])
	err := b.store.AddPhrase(m.Sender.ID, phrase, explanation)
	// TODO: keep user updated
	if err != nil {
		return messageErr, errors.Wrap(err, "failed to save phrase")
	}
	return "Phrase saved. Add next one.", nil
}

func (b bot) messageStudy(m messenger.Message) (string, []messenger.QuickReply, error) {
	if m.QuickReply == nil {
		return "Currently only quick replies are supported.", nil, errors.Errorf("not a quick reply: %v", m)
	}

	switch m.QuickReply.Payload {
	case payloadShow:
		study, err := b.store.GetStudy(m.Sender.ID)
		if err != nil {
			return messageErr, nil, errors.Wrap(err, "failed to show study")
		}
		switch study.Mode {
		case brain.ButtonsExplanation:
			return study.Explanation, buttonsScore, nil
		default:
			return messageErr, nil, errors.New("unknown study mode")
		}

	case payloadScoreBad:
		return b.scoreAndStudy(m.Sender.ID, brain.ScoreBad)
	case payloadScoreOk:
		return b.scoreAndStudy(m.Sender.ID, brain.ScoreOK)
	case payloadScoreGood:
		return b.scoreAndStudy(m.Sender.ID, brain.ScoreGood)
	default:
		return messageErr, nil, errors.Errorf("unknown payload: %s", m.QuickReply.Payload)
	}
}
