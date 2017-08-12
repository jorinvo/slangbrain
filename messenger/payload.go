package messenger

import (
	"fmt"

	"github.com/jorinvo/slangbrain/brain"
	"github.com/jorinvo/slangbrain/payload"
)

func (b Bot) handlePayload(u user, p string) {
	switch p {
	case payload.GetStarted:
		b.messageWelcome(u)

	case payload.Idle:
		b.send(u.ID, u.Msg.Idle, nil, nil)

	case payload.StartStudy:
		if err := b.store.SetMode(u.ID, brain.ModeStudy); err != nil {
			b.send(u.ID, u.Msg.Error, u.Btn.MenuMode, err)
			return
		}
		b.send(b.startStudy(u))

	case payload.StartAdd:
		if err := b.store.SetMode(u.ID, brain.ModeAdd); err != nil {
			b.send(u.ID, u.Msg.Error, u.Btn.MenuMode, err)
			return
		}
		b.send(u.ID, u.Msg.Add, u.Btn.AddMode, nil)

	case payload.ShowHelp:
		isSubscribed, err := b.store.IsSubscribed(u.ID)
		if err != nil {
			b.err.Println(err)
		}
		buttons := u.Btn.Help
		if !isSubscribed {
			buttons = buttons[1:]
		}
		b.send(u.ID, u.Msg.Help, buttons, nil)

	case payload.ShowStudy:
		study, err := b.store.GetStudy(u.ID)
		if err != nil {
			b.send(u.ID, u.Msg.Error, u.Btn.Show, fmt.Errorf("failed to get study: %v", err))
			return
		}
		b.send(u.ID, study.Phrase, u.Btn.Score, nil)

	case payload.ScoreBad:
		b.send(b.scoreAndStudy(u, -1))

	case payload.ScoreOk:
		b.send(b.scoreAndStudy(u, 0))

	case payload.ScoreGood:
		b.send(b.scoreAndStudy(u, 1))

	case payload.Delete:
		b.send(u.ID, u.Msg.ConfirmDelete, u.Btn.ConfirmDelete, nil)

	case payload.ConfirmDelete:
		if err := b.store.DeleteStudyPhrase(u.ID); err != nil {
			b.send(u.ID, u.Msg.Error, nil, nil)
		} else {
			b.send(u.ID, u.Msg.Deleted, nil, nil)
		}
		b.send(b.startStudy(u))

	case payload.CancelDelete:
		b.send(u.ID, u.Msg.CancelDelete, nil, nil)
		b.send(b.startStudy(u))

	case payload.Subscribe:
		if err := b.store.Subscribe(u.ID); err != nil {
			b.send(u.ID, u.Msg.Error, nil, nil)
			return
		}
		b.send(u.ID, u.Msg.Subscribed, u.Btn.MenuMode, nil)

	case payload.Unsubscribe:
		if err := b.store.Unsubscribe(u.ID); err != nil {
			b.send(u.ID, u.Msg.Error, nil, nil)
			return
		}
		b.send(u.ID, u.Msg.ConfirmUnsubscribe, u.Btn.MenuMode, nil)

	case payload.NoSubscription:
		b.send(u.ID, u.Msg.Unsubscribed, u.Btn.MenuMode, nil)

	case payload.Feedback:
		if err := b.store.SetMode(u.ID, brain.ModeFeedback); err != nil {
			b.send(u.ID, u.Msg.Error, u.Btn.MenuMode, err)
			return
		}
		b.send(u.ID, u.Msg.Fedback, u.Btn.Feedback, nil)

	case payload.StartMenu:
		fallthrough
	default:
		b.send(b.messageStartMenu(u))
	}
}
