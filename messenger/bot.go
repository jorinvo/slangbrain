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

type replySender func(string, []messenger.QuickReply, error)

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

func (b bot) ReadHandler(read messenger.Read, res *messenger.Response) {
	b.log.Printf("read message (%d)", res.To.ID)
	if err := b.store.SetRead(res.To.ID, read.Watermark()); err != nil {
		b.log.Println(err)
	}
}

func (b bot) MessageHandler(m messenger.Message, r *messenger.Response) {
	if m.IsEcho {
		return
	}

	b.logMessage(m)
	b.trackActivity(m.Sender.ID, m.Time)

	// Helper to send replies and log errors
	send := func(reply string, buttons []messenger.QuickReply, err error) {
		if err != nil {
			b.log.Println(err)
		}
		if err = r.TextWithReplies(reply, buttons); err != nil {
			b.log.Println("failed to send message:", err)
		}
	}

	if m.QuickReply != nil {
		b.handleQuickReplies(send, m.Sender.ID, m.QuickReply.Payload)
		return
	}
	b.handleMessages(send, m.Sender.ID, m.Text)
}

func (b bot) handleMessages(send replySender, id int64, msg string) {
	mode, err := b.store.GetMode(id)
	if err != nil {
		send(messageErr, buttonsMenuMode, fmt.Errorf("failed to get mode for id %v: %v", id, err))
		return
	}
	switch mode {
	case brain.ModeStudy:
		study, err := b.store.GetStudy(id)
		if err != nil {
			send(messageErr, buttonsStudyMode, fmt.Errorf("failed to get study: %v", err))
			return
		}
		// Score user unput and pick appropriate reply
		var score brain.Score = brain.ScoreGood
		m1 := messageStudyCorrect
		if normPhrase(msg) != normPhrase(study.Phrase) {
			score = brain.ScoreBad
			m1 = fmt.Sprintf(messageStudyWrong, study.Phrase)
		}
		send(m1, nil, nil)
		send(b.scoreAndStudy(id, score))

	case brain.ModeAdd:
		parts := strings.SplitN(strings.TrimSpace(msg), "\n", 2)
		phrase := strings.TrimSpace(parts[0])
		if phrase == "" {
			send(messagePhraseEmpty, buttonsAddMode, nil)
			return
		}
		if len(parts) == 1 {
			send(messageExplanationEmpty, buttonsAddMode, nil)
			return
		}
		explanation := strings.TrimSpace(parts[1])
		// Check for existing explanation
		p, err := b.store.FindPhrase(id, func(p brain.Phrase) bool {
			return p.Explanation == explanation
		})
		if err != nil {
			send(messageErr, nil, fmt.Errorf("failed to lookup phrase: %v", err))
			return
		}
		if p.Phrase != "" {
			send(fmt.Sprintf(messageExplanationExists, p.Phrase, p.Explanation), buttonsAddMode, nil)
			return
		}
		// Save phrase
		if err = b.store.AddPhrase(id, phrase, explanation); err != nil {
			send(messageErr, buttonsAddMode, fmt.Errorf("failed to save phrase: %v", err))
			return
		}
		send(fmt.Sprintf(messageAddDone, phrase, explanation), nil, nil)
		send(messageAddNext, buttonsAddMode, nil)

	case brain.ModeGetStarted:
		send(messageWelcome, nil, b.store.SetMode(id, brain.ModeAdd))
		send(messageWelcome2, nil, nil)

	default:
		send(b.messageStartMenu(id))
	}
}

func (b bot) handleQuickReplies(send replySender, id int64, payload string) {
	switch payload {
	case payloadIdle:
		send(messageIdle, nil, nil)

	case payloadStartStudy:
		if err := b.store.SetMode(id, brain.ModeStudy); err != nil {
			send(messageErr, buttonsMenuMode, err)
			return
		}
		send(b.startStudy(id))

	case payloadStartAdd:
		if err := b.store.SetMode(id, brain.ModeAdd); err != nil {
			send(messageErr, buttonsMenuMode, err)
			return
		}
		send(messageStartAdd, buttonsAddMode, nil)

	case payloadShowHelp:
		send(messageHelp, buttonsHelp, nil)

	case payloadShowStudy:
		study, err := b.store.GetStudy(id)
		if err != nil {
			send(messageErr, buttonsShow, fmt.Errorf("failed to get study: %v", err))
			return
		}
		send(study.Phrase, buttonsScore, nil)

	case payloadScoreBad:
		send(b.scoreAndStudy(id, brain.ScoreBad))

	case payloadScoreOk:
		send(b.scoreAndStudy(id, brain.ScoreOK))

	case payloadScoreGood:
		send(b.scoreAndStudy(id, brain.ScoreGood))

	case payloadStartMenu:
		fallthrough
	default:
		send(b.messageStartMenu(id))
	}
}

func (b bot) logMessage(m messenger.Message) {
	logMsg := "message: "
	if m.QuickReply != nil {
		logMsg += "[" + m.QuickReply.Payload + "] "
	}
	b.log.Println(logMsg + m.Text)
}

func (b bot) trackActivity(chatID int64, t time.Time) {
	if err := b.store.SetActivity(chatID, t); err != nil {
		b.log.Println(err)
	}
}

func (b bot) messageStartMenu(chatID int64) (string, []messenger.QuickReply, error) {
	if err := b.store.SetMode(chatID, brain.ModeMenu); err != nil {
		return messageErr, buttonsMenuMode, err
	}
	return messageStartMenu, buttonsMenuMode, nil
}

func (b bot) startStudy(chatID int64) (string, []messenger.QuickReply, error) {
	study, err := b.store.GetStudy(chatID)
	if err != nil {
		return messageErr, buttonsStudyMode, err
	}
	// No studies ready
	if study.Total == 0 {
		// Go to menu mode
		if err = b.store.SetMode(chatID, brain.ModeMenu); err != nil {
			return messageErr, buttonsStudyMode, err
		}
		// Display time until next study is ready or there are not studies yet
		msg := messageStudyEmpty
		if study.Next > 0 {
			msg = fmt.Sprintf(messageStudyDone, formatDuration(study.Next))
		}
		return msg, buttonsMenuMode, nil
	}
	// Send study to user
	return fmt.Sprintf(messageStudyQuestion, study.Total, study.Explanation), buttonsShow, nil
}

func (b bot) scoreAndStudy(chatID int64, score brain.Score) (string, []messenger.QuickReply, error) {
	err := b.store.ScoreStudy(chatID, score)
	if err != nil {
		return messageErr, buttonsStudyMode, err
	}
	return b.startStudy(chatID)
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
	return specialChars.ReplaceAllString(strings.ToLower(strings.TrimSpace(s)), "")
}
