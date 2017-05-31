package messenger

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/jorinvo/messenger"
	"github.com/jorinvo/slangbrain/brain"
)

// Everything that is not in the unicode character classes
// for letters or numeric values
// See: http://www.fileformat.info/info/unicode/category/index.htm
var specialChars = regexp.MustCompile(`[^\p{Ll}\p{Lm}\p{Lo}\p{Lu}\p{Nd}\p{Nl}\p{No}]`)
var inParantheses = regexp.MustCompile(`\(.*?\)`)

type replySender func(string, []messenger.QuickReply, error)

// Only handling the Get Started button here
func (b bot) PostbackHandler(p messenger.PostBack, r *messenger.Response) {
	if p.Payload != payloadGetStarted {
		return
	}

	b.messageWelcome(p.Sender.ID)
}

func (b bot) ReadHandler(read messenger.Read, res *messenger.Response) {
	if err := b.store.SetRead(res.To.ID, read.Watermark()); err != nil {
		b.log.Println(err)
	}
}

func (b bot) MessageHandler(m messenger.Message, r *messenger.Response) {
	if m.IsEcho {
		return
	}

	b.trackActivity(m.Sender.ID, m.Time)

	if m.QuickReply != nil {
		b.handleQuickReplies(m.Sender.ID, m.QuickReply.Payload)
		return
	}
	b.handleMessages(m.Sender.ID, m.Text)
}

func (b bot) handleMessages(id int64, msg string) {
	mode, err := b.store.GetMode(id)
	if err != nil {
		b.send(id, messageErr, buttonsMenuMode, fmt.Errorf("failed to get mode for id %v: %v", id, err))
		return
	}
	switch mode {
	case brain.ModeStudy:
		study, err := b.store.GetStudy(id)
		if err != nil {
			b.send(id, messageErr, buttonsStudyMode, fmt.Errorf("failed to get study: %v", err))
			return
		}
		// Score user unput and pick appropriate reply
		var score brain.Score = brain.ScoreGood
		m1 := messageStudyCorrect
		if normPhrase(msg) != normPhrase(study.Phrase) {
			score = brain.ScoreBad
			m1 = fmt.Sprintf(messageStudyWrong, study.Phrase)
		}
		b.send(id, m1, nil, nil)
		b.send(b.scoreAndStudy(id, score))

	case brain.ModeAdd:
		parts := strings.SplitN(strings.TrimSpace(msg), "\n", 2)
		phrase := strings.TrimSpace(parts[0])
		if phrase == "" {
			b.send(id, messagePhraseEmpty, buttonsAddMode, nil)
			return
		}
		if len(parts) == 1 {
			b.send(id, messageExplanationEmpty, buttonsAddMode, nil)
			return
		}
		explanation := strings.TrimSpace(parts[1])
		// Check for existing explanation
		p, err := b.store.FindPhrase(id, func(p brain.Phrase) bool {
			return p.Explanation == explanation
		})
		if err != nil {
			b.send(id, messageErr, nil, fmt.Errorf("failed to lookup phrase: %v", err))
			return
		}
		if p.Phrase != "" {
			b.send(id, fmt.Sprintf(messageExplanationExists, p.Phrase, p.Explanation), buttonsAddMode, nil)
			return
		}
		// Save phrase
		if err = b.store.AddPhrase(id, phrase, explanation); err != nil {
			b.send(id, messageErr, buttonsAddMode, fmt.Errorf("failed to save phrase: %v", err))
			return
		}
		b.send(id, fmt.Sprintf(messageAddDone, phrase, explanation), nil, nil)
		b.send(id, messageAddNext, buttonsAddMode, nil)

	case brain.ModeGetStarted:
		b.messageWelcome(id)

	default:
		b.send(b.messageStartMenu(id))
	}
}

func (b bot) handleQuickReplies(id int64, payload string) {
	switch payload {
	case payloadIdle:
		b.send(id, messageIdle, nil, nil)

	case payloadStartStudy:
		if err := b.store.SetMode(id, brain.ModeStudy); err != nil {
			b.send(id, messageErr, buttonsMenuMode, err)
			return
		}
		b.send(b.startStudy(id))

	case payloadStartAdd:
		if err := b.store.SetMode(id, brain.ModeAdd); err != nil {
			b.send(id, messageErr, buttonsMenuMode, err)
			return
		}
		b.send(id, messageStartAdd, buttonsAddMode, nil)

	case payloadShowHelp:
		b.send(id, messageHelp, buttonsHelp, nil)

	case payloadShowStudy:
		study, err := b.store.GetStudy(id)
		if err != nil {
			b.send(id, messageErr, buttonsShow, fmt.Errorf("failed to get study: %v", err))
			return
		}
		b.send(id, study.Phrase, buttonsScore, nil)

	case payloadScoreBad:
		b.send(b.scoreAndStudy(id, brain.ScoreBad))

	case payloadScoreOk:
		b.send(b.scoreAndStudy(id, brain.ScoreOK))

	case payloadScoreGood:
		b.send(b.scoreAndStudy(id, brain.ScoreGood))

	case payloadDelete:
		b.send(id, messageConfirmDelete, buttonsConfirmDelete, nil)

	case payloadConfirmDelete:
		if err := b.store.DeleteStudyPhrase(id); err != nil {
			b.send(id, messageErr, nil, nil)
		} else {
			b.send(id, messageDeleted, nil, nil)
		}
		b.send(b.startStudy(id))

	case payloadCancelDelete:
		b.send(id, messageCancelDelete, nil, nil)
		b.send(b.startStudy(id))

	case payloadStartMenu:
		fallthrough
	default:
		b.send(b.messageStartMenu(id))
	}
}

func (b bot) trackActivity(id int64, t time.Time) {
	if err := b.store.SetActivity(id, t); err != nil {
		b.log.Println(err)
	}
}

func (b bot) messageStartMenu(id int64) (int64, string, []messenger.QuickReply, error) {
	if err := b.store.SetMode(id, brain.ModeMenu); err != nil {
		return id, messageErr, buttonsMenuMode, err
	}
	return id, messageStartMenu, buttonsMenuMode, nil
}

func (b bot) messageWelcome(id int64) {
	p, err := b.client.ProfileByID(id)
	if err != nil {
		b.log.Printf("failed to get profile for %d: %v", id, err)
	}
	b.send(id, fmt.Sprintf(messageWelcome, p.FirstName), nil, nil)
	b.send(id, messageWelcome2, nil, b.store.SetMode(id, brain.ModeAdd))
}

func (b bot) startStudy(id int64) (int64, string, []messenger.QuickReply, error) {
	study, err := b.store.GetStudy(id)
	if err != nil {
		return id, messageErr, buttonsStudyMode, err
	}
	// No studies ready
	if study.Total == 0 {
		// Go to menu mode
		if err = b.store.SetMode(id, brain.ModeMenu); err != nil {
			return id, messageErr, buttonsStudyMode, err
		}
		// Display time until next study is ready or there are not studies yet
		if study.Next == 0 {
			return id, messageStudyEmpty, buttonsStudyEmpty, nil
		}
		return id, fmt.Sprintf(messageStudyDone, formatDuration(study.Next)), buttonsMenuMode, nil
	}
	// Send study to user
	return id, fmt.Sprintf(messageStudyQuestion, study.Total, study.Explanation), buttonsShow, nil
}

func (b bot) scoreAndStudy(id int64, score brain.Score) (int64, string, []messenger.QuickReply, error) {
	err := b.store.ScoreStudy(id, score)
	if err != nil {
		return id, messageErr, buttonsStudyMode, err
	}
	return b.startStudy(id)
}

// Send replies and log errors
func (b bot) send(id int64, reply string, buttons []messenger.QuickReply, err error) {
	if err != nil {
		b.log.Println(err)
	}
	if err = b.client.SendWithReplies(messenger.Recipient{ID: id}, reply, buttons); err != nil {
		b.log.Println("failed to send message:", err)
	}
}

// Format like "X hour[s] X minute[s]".
// Returns empty string for negativ durations.
func formatDuration(d time.Duration) string {
	// Precision in minutes
	d = time.Duration(math.Ceil(float64(d)/float64(time.Minute))) * time.Minute
	s := ""
	h := d / time.Hour
	m := (d - h*time.Hour) / time.Minute
	if h > 1 {
		s += fmt.Sprintf("%d", h) + " hours "
	} else if h == 1 {
		s += "1 hour "
	}
	if m > 1 {
		s += fmt.Sprintf("%d", m) + " minutes"
	} else if m > 0 {
		s += "1 minute"
	} else if s != "" {
		// No minutes, only hours, remove trailing space
		s = s[:len(s)-1]
	}
	return s
}
func normPhrase(s string) string {
	s = inParantheses.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	return specialChars.ReplaceAllString(s, "")
}
