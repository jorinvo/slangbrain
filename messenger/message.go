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

	if m.QuickReply != nil && m.QuickReply.Payload == payloadStartIdle {
		fn = b.messageStartIdle
	} else if m.QuickReply != nil && m.QuickReply.Payload == payloadStartStudy {
		fn = b.messageStartStudy
	} else if m.QuickReply != nil && m.QuickReply.Payload == payloadStartAdd {
		fn = b.messageStartAdd
	} else if mode == brain.ModeStudy {
		fn = b.messageStudy
	} else if mode == brain.ModeAdd {
		fn = b.messageAdd
	} else if mode == brain.ModeGetStarted {
		fn = b.messageGetStarted
	} else {
		fn = b.messageStartIdle
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

func (b bot) messageStartIdle(m messenger.Message) (string, []messenger.QuickReply, error) {
	err := b.store.SetMode(m.Sender.ID, brain.ModeIdle)
	if err != nil {
		return messageErr, buttonsIdleMode, err
	}
	return messageStartIdle, buttonsIdleMode, nil
}

func (b bot) messageStartAdd(m messenger.Message) (string, []messenger.QuickReply, error) {
	err := b.store.SetMode(m.Sender.ID, brain.ModeAdd)
	if err != nil {
		return messageErr, buttonsIdleMode, err
	}
	return messageStartAdd, buttonsAddMode, nil
}

func (b bot) messageStartStudy(m messenger.Message) (string, []messenger.QuickReply, error) {
	err := b.store.SetMode(m.Sender.ID, brain.ModeStudy)
	if err != nil {
		return messageErr, buttonsIdleMode, err
	}
	return b.startStudy(m.Sender.ID)
}

func (b bot) messageAdd(m messenger.Message) (string, []messenger.QuickReply, error) {
	parts := strings.SplitN(m.Text, "\n", 2)
	if len(parts) == 1 {
		return messageErrExplanation, buttonsAddMode, fmt.Errorf("explanation missing: %s", m.Text)
	}
	phrase := strings.TrimSpace(parts[0])
	explanation := strings.TrimSpace(parts[1])
	// Check for existing phrases
	p, err := b.store.FindPhrase(m.Sender.ID, func(p brain.Phrase) bool {
		return p.Phrase == phrase
	})
	if err != nil {
		return messageErr, nil, fmt.Errorf("failed to lookup phrase: %v", err)
	}
	if p.Phrase != "" {
		return fmt.Sprintf(messagePhraseExists, p.Phrase, p.Explanation), buttonsAddMode, nil
	}
	// Check for existing explanation
	p, err = b.store.FindPhrase(m.Sender.ID, func(p brain.Phrase) bool {
		return p.Explanation == explanation
	})
	if err != nil {
		return messageErr, nil, fmt.Errorf("failed to lookup phrase: %v", err)
	}
	if p.Phrase != "" {
		return fmt.Sprintf(messageExplanationExists, p.Phrase, p.Explanation), buttonsAddMode, nil
	}
	// Save phrase
	err = b.store.AddPhrase(m.Sender.ID, phrase, explanation)
	// TODO: keep user updated
	if err != nil {
		return messageErr, buttonsAddMode, fmt.Errorf("failed to save phrase: %v", err)
	}
	// count, err := b.store.CountStudies(m.Sender.ID)
	// var buttons []messenger.QuickReply
	// if err == nil && count > 0 {
	// 	buttons = []messenger.QuickReply{
	// 		button(fmt.Sprintf("study (%d)", count), payloadStartStudy),
	// 	}
	// }
	// Fail silent when counting errors
	return fmt.Sprintf(messageAddDone, phrase, explanation), buttonsAddMode, err
}

func (b bot) messageStudy(m messenger.Message) (string, []messenger.QuickReply, error) {
	if m.QuickReply == nil {
		return "Currently only quick replies are supported.", buttonsStudyMode, fmt.Errorf("not a quick reply: %v", m)
	}

	switch m.QuickReply.Payload {
	case payloadShow:
		study, err := b.store.GetStudy(m.Sender.ID)
		if err != nil {
			return messageErr, buttonsShow, fmt.Errorf("failed to show study: %v", err)
		}
		switch study.Mode {
		case brain.GuessExplanation:
			return study.Explanation, buttonsScore, nil
		case brain.GuessPhrase:
			return study.Phrase, buttonsScore, nil
		default:
			return messageErr, buttonsStudyMode, fmt.Errorf("SHOULD NEVER HAPPEN: unknown study mode: %v (%v)", study.Mode, study)
		}

	case payloadScoreBad:
		return b.scoreAndStudy(m.Sender.ID, brain.ScoreBad)
	case payloadScoreOk:
		return b.scoreAndStudy(m.Sender.ID, brain.ScoreOK)
	case payloadScoreGood:
		return b.scoreAndStudy(m.Sender.ID, brain.ScoreGood)
	default:
		return messageErr, buttonsStudyMode, fmt.Errorf("unknown payload: %s", m.QuickReply.Payload)
	}
}

func (b bot) messageGetStarted(m messenger.Message) (string, []messenger.QuickReply, error) {
	// For now, start in add mode.
	// Later there might be a better introduction for users.
	err := b.store.SetMode(m.Sender.ID, brain.ModeAdd)
	return messageWelcome, nil, err
}
