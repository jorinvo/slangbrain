package messenger

import (
	"fmt"

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
		return messageStudyDone, buttonsIdleMode, nil
	}
	switch study.Mode {
	case brain.GuessExplanation:
		return fmt.Sprintf(messageStudyQuestion, study.Phrase), buttonsShow, nil
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
