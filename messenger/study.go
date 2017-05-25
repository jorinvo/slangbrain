package messenger

import (
	"fmt"

	"github.com/jorinvo/messenger"
	"github.com/jorinvo/slangbrain/brain"
	"github.com/pkg/errors"
)

func (b bot) study(chatID int64) (string, []messenger.QuickReply, error) {
	study, err := b.store.GetStudy(chatID)
	if err != nil {
		return messageErr, nil, errors.Wrapf(err, "failed to study with id %v", chatID)
	}
	if study.Total == 0 {
		return messageStudyDone, nil, nil
	}
	switch study.Mode {
	case brain.ButtonsExplanation:
		return fmt.Sprintf(messageButtons, study.Phrase), buttonsShow, nil
	default:
		return messageErr, nil, errors.Errorf("unknown study mode %v (%v)", study.Mode, study)
	}
}

func (b bot) scoreAndStudy(chatID int64, score brain.Score) (string, []messenger.QuickReply, error) {
	err := b.store.ScoreStudy(chatID, score)
	if err != nil {
		return messageErr, nil, errors.Wrapf(err, "failed to score study with id %v", chatID)
	}
	return b.study(chatID)
}
