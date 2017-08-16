package messenger

import (
	"fmt"

	"github.com/jorinvo/slangbrain/brain"
	"github.com/jorinvo/slangbrain/payload"
	"github.com/jorinvo/slangbrain/user"
)

func (b Bot) handlePayload(u user.User, p string) {
	switch p {
	case payload.GetStarted:
		b.messageWelcome(u)

	case payload.Idle:
		b.send(u.ID, u.Msg.Idle, nil, nil)

	case payload.Study:
		if err := b.store.SetMode(u.ID, brain.ModeStudy); err != nil {
			b.send(u.ID, u.Msg.Error, u.Rpl.MenuMode, err)
			return
		}
		b.send(b.startStudy(u))

	case payload.Add:
		if err := b.store.SetMode(u.ID, brain.ModeAdd); err != nil {
			b.send(u.ID, u.Msg.Error, u.Rpl.MenuMode, err)
			return
		}
		b.send(u.ID, u.Msg.Add, u.Rpl.AddMode, nil)

	case payload.Help:
		// Generate manage token in case user clicks the manage button
		token, err := b.store.GenerateToken(u.ID)
		if err != nil {
			b.err.Println(err)
		}
		isSubscribed, err := b.store.IsSubscribed(u.ID)
		if err != nil {
			b.err.Println(err)
		}
		buttons := u.Btn.Help(isSubscribed, token)
		if err = b.client.SendWithButtons(u.ID, u.Msg.Help, u.Rpl.Help, buttons); err != nil {
			b.err.Println("failed to send message:", err)
		}

	case payload.ShowPhrase:
		study, err := b.store.GetStudy(u.ID)
		if err != nil {
			b.send(u.ID, u.Msg.Error, u.Rpl.Show, fmt.Errorf("failed to get study: %v", err))
			return
		}
		b.send(u.ID, study.Phrase, u.Rpl.Score, nil)

	case payload.ScoreBad:
		b.send(b.scoreAndStudy(u, -1))

	case payload.ScoreOk:
		b.send(b.scoreAndStudy(u, 0))

	case payload.ScoreGood:
		b.send(b.scoreAndStudy(u, 1))

	case payload.Delete:
		b.send(u.ID, u.Msg.ConfirmDelete, u.Rpl.ConfirmDelete, nil)

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
		b.send(u.ID, u.Msg.Subscribed+"\n\n"+u.Msg.Menu, u.Rpl.MenuMode, nil)

	case payload.Unsubscribe:
		if err := b.store.Unsubscribe(u.ID); err != nil {
			b.send(u.ID, u.Msg.Error, nil, nil)
			return
		}
		b.send(u.ID, u.Msg.ConfirmUnsubscribe+"\n\n"+u.Msg.Menu, u.Rpl.MenuMode, nil)

	case payload.DenySubscribe:
		b.send(u.ID, u.Msg.DenySubscribe+"\n\n"+u.Msg.Menu, u.Rpl.MenuMode, nil)

	case payload.Feedback:
		if err := b.store.SetMode(u.ID, brain.ModeFeedback); err != nil {
			b.send(u.ID, u.Msg.Error, u.Rpl.MenuMode, err)
			return
		}
		b.send(u.ID, u.Msg.Feedback, u.Rpl.Feedback, nil)

	case payload.Menu:
		fallthrough
	default:
		b.send(b.messageStartMenu(u))
	}
}
