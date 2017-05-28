package messenger

import (
	"fmt"
	"math"
	"time"

	"github.com/jorinvo/messenger"
	"github.com/jorinvo/slangbrain/brain"
)

func (b bot) startStudy(chatID int64) (string, []messenger.QuickReply, error) {
	study, err := b.store.GetStudy(chatID)
	if err != nil {
		return messageErr, buttonsStudyMode, err
	}
	if study.Total == 0 {
		if err = b.store.SetMode(chatID, brain.ModeIdle); err != nil {
			return messageErr, buttonsStudyMode, err
		}
		msg := messageStudyEmpty
		if study.Next > 0 {
			msg = fmt.Sprintf(messageStudyDone, formatDuration(study.Next))
		}
		return msg, buttonsIdleMode, nil
	}
	switch study.Mode {
	case brain.GuessExplanation:
		return fmt.Sprintf(messageStudyQuestion, study.Phrase), buttonsShow, nil
	case brain.GuessPhrase:
		return fmt.Sprintf(messageStudyQuestion, study.Explanation), buttonsShow, nil
	default:
		return messageErr, nil, fmt.Errorf("unknown study mode %v (%v)", study.Mode, study)
	}
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
