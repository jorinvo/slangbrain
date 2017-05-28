package messenger

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jorinvo/messenger"
	"github.com/jorinvo/slangbrain/brain"
)

// Everything that is not in the unicode character classes
// for letters or numeric values
// See: http://www.fileformat.info/info/unicode/category/index.htm
var specialChars = regexp.MustCompile(`[^\p{Ll}\p{Lm}\p{Lo}\p{Lu}\p{Nd}\p{Nl}\p{No}]`)

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
		// Start idle mode
		fn = b.messageStartIdle
	} else if m.QuickReply != nil && m.QuickReply.Payload == payloadStartStudy {
		// Start study mode
		fn = b.messageStartStudy
	} else if m.QuickReply != nil && m.QuickReply.Payload == payloadStartAdd {
		// Start add mode
		fn = b.messageStartAdd
	} else if mode == brain.ModeStudy {
		// Handle quick replies and messages for study mode
		fn = b.messageStudy
	} else if mode == brain.ModeAdd {
		// Handle quick replies and messages for add mode
		fn = b.messageAdd
	} else if mode == brain.ModeGetStarted {
		// Get started when no mode was set
		fn = b.messageGetStarted
	} else {
		// If something goes wrong fall back to idle mode
		fn = b.messageStartIdle
	}

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
	// Check for existing explanation
	p, err := b.store.FindPhrase(m.Sender.ID, func(p brain.Phrase) bool {
		return p.Explanation == explanation
	})
	if err != nil {
		return messageErr, nil, fmt.Errorf("failed to lookup phrase: %v", err)
	}
	if p.Phrase != "" {
		return fmt.Sprintf(messageExplanationExists, p.Phrase, p.Explanation), buttonsAddMode, nil
	}
	// Save phrase
	if err = b.store.AddPhrase(m.Sender.ID, phrase, explanation); err != nil {
		return messageErr, buttonsAddMode, fmt.Errorf("failed to save phrase: %v", err)
	}
	return fmt.Sprintf(messageAddDone, phrase, explanation), buttonsAddMode, nil
}

func (b bot) messageStudy(m messenger.Message) (string, []messenger.QuickReply, error) {
	id := m.Sender.ID
	// Handle message
	if m.QuickReply == nil {
		study, err := b.store.GetStudy(id)
		if err != nil {
			return messageErr, buttonsStudyMode, fmt.Errorf("failed to get study: %v", err)
		}
		// Ignore special chars
		userReply := specialChars.ReplaceAllString(strings.TrimSpace(m.Text), "")
		phrase := specialChars.ReplaceAllString(study.Phrase, "")
		// Score user reply and pick appropriate reply
		var score brain.Score
		var m1 string
		if userReply == phrase {
			score = brain.ScoreGood
			m1 = messageStudyCorrect
		} else {
			score = brain.ScoreBad
			m1 = fmt.Sprintf(messageStudyWrong, study.Phrase)
		}
		m2, b, err := b.scoreAndStudy(id, score)
		if err != nil {
			return m2, b, err
		}
		return m1 + m2, b, nil
	}
	// Handle quick replies
	switch m.QuickReply.Payload {
	case payloadShow:
		study, err := b.store.GetStudy(id)
		if err != nil {
			return messageErr, buttonsShow, fmt.Errorf("failed to get study: %v", err)
		}
		return study.Phrase, buttonsScore, nil
	case payloadScoreBad:
		return b.scoreAndStudy(id, brain.ScoreBad)
	case payloadScoreOk:
		return b.scoreAndStudy(id, brain.ScoreOK)
	case payloadScoreGood:
		return b.scoreAndStudy(id, brain.ScoreGood)
	default:
		return messageErr, buttonsStudyMode, fmt.Errorf("unknown payload: %s", m.QuickReply.Payload)
	}
}

func (b bot) messageGetStarted(m messenger.Message) (string, []messenger.QuickReply, error) {
	err := b.store.SetMode(m.Sender.ID, brain.ModeAdd)
	return messageWelcome, nil, err
}
