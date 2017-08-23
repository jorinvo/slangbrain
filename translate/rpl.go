package translate

import (
	"github.com/jorinvo/slangbrain/fbot"
	"github.com/jorinvo/slangbrain/payload"
)

// Rpl contains all reply sets that can be sent to a user.
// They are already localized for one language.
type Rpl struct {
	MenuMode,
	Subscribe,
	HelpSubscribe,
	HelpUnsubscribe,
	Feedback,
	AddMode,
	StudyMode,
	Show,
	Score,
	StudyEmpty,
	StudiesDue,
	ConfirmDelete,
	ImportHelp,
	Import []fbot.Reply
}

func newRpl(l labels) Rpl {
	var (
		studyDone    = fbot.Reply{Text: l.StudyDone, Payload: payload.Menu}
		study        = fbot.Reply{Text: iconStudy + " " + l.Study, Payload: payload.Study}
		add          = fbot.Reply{Text: iconAdd + " " + l.Add, Payload: payload.Add}
		done         = fbot.Reply{Text: iconDone + " " + l.Done, Payload: payload.Idle}
		help         = fbot.Reply{Text: iconHelp + " " + l.Help, Payload: payload.Help}
		deletePhrase = fbot.Reply{Text: iconDelete, Payload: payload.Delete}
	)

	return Rpl{
		MenuMode: []fbot.Reply{
			study,
			add,
			help,
			done,
		},
		Subscribe: []fbot.Reply{
			fbot.Reply{Text: iconGood + " " + l.SubscribeConfirm, Payload: payload.Subscribe},
			fbot.Reply{Text: l.SubscribeDeny, Payload: payload.DenySubscribe},
		},
		HelpSubscribe: []fbot.Reply{
			fbot.Reply{Text: l.EnableNotifications, Payload: payload.Subscribe},
			fbot.Reply{Text: l.SendFeedback, Payload: payload.Feedback},
			fbot.Reply{Text: l.QuitHelp, Payload: payload.Menu},
		},
		HelpUnsubscribe: []fbot.Reply{
			fbot.Reply{Text: l.DisableNotifications, Payload: payload.Unsubscribe},
			fbot.Reply{Text: l.SendFeedback, Payload: payload.Feedback},
			fbot.Reply{Text: l.QuitHelp, Payload: payload.Menu},
		},
		Feedback: []fbot.Reply{
			fbot.Reply{Text: iconDelete + " " + l.CancelFeedback, Payload: payload.Menu},
		},
		AddMode: []fbot.Reply{
			fbot.Reply{Text: l.StopAdding, Payload: payload.Menu},
		},
		StudyMode: []fbot.Reply{
			studyDone,
		},
		Show: []fbot.Reply{
			deletePhrase,
			studyDone,
			fbot.Reply{Text: iconShow + " " + l.ShowPhrase, Payload: payload.ShowPhrase},
		},
		Score: []fbot.Reply{
			deletePhrase,
			fbot.Reply{Text: iconBad + " " + l.ScoreBad, Payload: payload.ScoreBad},
			fbot.Reply{Text: iconOK, Payload: payload.ScoreOk},
			fbot.Reply{Text: iconGood + " " + l.ScoreGood, Payload: payload.ScoreGood},
		},
		StudyEmpty: []fbot.Reply{
			add,
		},
		StudiesDue: []fbot.Reply{
			study,
			fbot.Reply{Text: l.StudyNotNow, Payload: payload.Menu},
		},
		ConfirmDelete: []fbot.Reply{
			fbot.Reply{Text: iconDelete + " " + l.ConfirmDelete, Payload: payload.ConfirmDelete},
			fbot.Reply{Text: l.CancelDelete, Payload: payload.CancelDelete},
		},
		ImportHelp: []fbot.Reply{
			fbot.Reply{Text: l.CloseImportHelp, Payload: payload.Menu},
		},
		Import: []fbot.Reply{
			fbot.Reply{Text: iconGood + " " + l.ConfirmImport, Payload: payload.ConfirmImport},
			fbot.Reply{Text: l.CancelImport, Payload: payload.CancelImport},
		},
	}
}
