package messenger

import (
	"fmt"
	"time"

	"github.com/jorinvo/slangbrain/brain"
	"github.com/jorinvo/slangbrain/payload"
	"github.com/jorinvo/slangbrain/user"
)

func (b Bot) handlePayload(u user.User, p, referral string) {
	isDuplicate, err := b.store.IsDuplicate(u.ID, p)
	if err != nil {
		b.err.Println(err)
	}
	if isDuplicate {
		b.info.Printf("[id=%d,p=%s] same payload sent twice in a row", u.ID, p)
		return
	}

	switch p {
	case payload.GetStarted:
		b.messageWelcome(u, referral)

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
		replies := u.Rpl.HelpSubscribe
		if isSubscribed {
			replies = u.Rpl.HelpUnsubscribe
		}
		buttons := u.Btn.Help(token)
		if err = b.client.SendWithButtons(u.ID, u.Msg.Help, replies, buttons); err != nil {
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
		b.send(b.scoreAndStudy(u, -2))

	case payload.ScoreOk:
		b.send(b.scoreAndStudy(u, 0))

	case payload.ScoreGood:
		b.send(b.scoreAndStudy(u, 1))

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

	case payload.ImportHelp:
		b.send(u.ID, u.Msg.ImportHelp1, nil, nil)
		time.Sleep(b.messageDelay)
		b.send(u.ID, u.Msg.ImportHelp2, u.Rpl.ImportHelp, nil)

	case payload.ConfirmImport:
		count, err := b.store.ApplyImport(u.ID)
		if err != nil {
			b.send(u.ID, u.Msg.Error, u.Rpl.MenuMode, err)
			return
		}
		b.send(u.ID, fmt.Sprintf(u.Msg.ImportConfirm, count)+"\n\n"+u.Msg.Menu, u.Rpl.MenuMode, nil)

	case payload.CancelImport:
		b.send(u.ID, u.Msg.ImportCancel+"\n\n"+u.Msg.Menu, u.Rpl.MenuMode, b.store.ClearImport(u.ID))

	case payload.Menu:
		fallthrough
	default:
		b.send(b.messageStartMenu(u))
	}
}
