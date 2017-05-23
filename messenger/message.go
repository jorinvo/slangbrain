package messenger

import (
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
		return
	}

	if mode == brain.ModeStudy {
		msg := messageErr
		var buttons []messenger.QuickReply
		if m.QuickReply == nil {
			return
		}
		switch m.QuickReply.Payload {
		case payloadShow:
			study, err := b.store.GetStudy(m.Sender.ID)
			if err != nil {
				b.log.Println("failed to show study:", err)
				break
			}
			switch study.Mode {
			case brain.ButtonsExplanation:
				msg = study.Explanation
				buttons = buttonsScore
			}
		case payloadScoreBad:
			msg, buttons, err = b.scoreAndStudy(m.Sender.ID, brain.ScoreBad)
			if err != nil {
				b.log.Println("failed to continue studies:", err)
			}
		case payloadScoreOk:
			msg, buttons, err = b.scoreAndStudy(m.Sender.ID, brain.ScoreOK)
			if err != nil {
				b.log.Println("failed to continue studies:", err)
			}
		case payloadScoreGood:
			msg, buttons, err = b.scoreAndStudy(m.Sender.ID, brain.ScoreGood)
			if err != nil {
				b.log.Println("failed to continue studies:", err)
			}
		}

		err = r.TextWithReplies(msg, buttons)
		if err != nil {
			b.log.Println("failed to send message:", err)
		}
	}

	if mode == brain.ModeAdd {
		parts := strings.SplitN(m.Text, "\n", 2)
		phrase := strings.TrimSpace(parts[0])
		explanation := ""
		if len(parts) > 1 {
			explanation = strings.TrimSpace(parts[1])
		}

		err = b.store.AddPhrase(m.Sender.ID, phrase, explanation)
		if err != nil {
			b.log.Println("failed to save phrase:", err)
			// TODO: keep user updated
			return
		}

		err = r.Text("Phrase saved. Add next one.")
		if err != nil {
			b.log.Println("failed to send message:", err)
		}
	}
}
