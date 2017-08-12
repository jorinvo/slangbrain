package translate

import (
	"github.com/jorinvo/slangbrain/fbot"
	"github.com/jorinvo/slangbrain/payload"
)

// Btn contains all button sets that can be sent to a user.
// They are already localized for one language.
type Btn struct {
	MenuMode,
	Subscribe,
	Help,
	Feedback,
	AddMode,
	StudyMode,
	Show,
	Score,
	StudyEmpty,
	StudiesDue,
	ConfirmDelete []fbot.Button
}

type buttonLabels struct {
	StudyDone,
	Study,
	Add,
	Done,
	Help,
	SubscribeConfirm,
	SubscribeDeny,
	StopNotifications,
	SendFeedback,
	QuitHelp,
	CancelFeedback,
	StopAdding,
	ShowPhrase,
	ScoreBad,
	ScoreGood,
	StudyNotNow,
	ConfirmDelete,
	CancelDelete string
}

func newBtn(b buttonLabels) Btn {
	var (
		buttonStudyDone = fbot.Button{Text: b.StudyDone, Payload: payload.StartMenu}
		buttonStudy     = fbot.Button{Text: iconStudy + " " + b.Study, Payload: payload.StartStudy}
		buttonAdd       = fbot.Button{Text: iconAdd + " " + b.Add, Payload: payload.StartAdd}
		buttonDone      = fbot.Button{Text: iconDone + " " + b.Done, Payload: payload.Idle}
		buttonHelp      = fbot.Button{Text: iconHelp + " " + b.Help, Payload: payload.ShowHelp}
		buttonDelete    = fbot.Button{Text: iconDelete, Payload: payload.Delete}
	)

	return Btn{
		MenuMode: []fbot.Button{
			buttonStudy,
			buttonAdd,
			buttonHelp,
			buttonDone,
		},
		Subscribe: []fbot.Button{
			fbot.Button{Text: iconGood + " " + b.SubscribeConfirm, Payload: payload.Subscribe},
			fbot.Button{Text: b.SubscribeDeny, Payload: payload.NoSubscription},
		},
		Help: []fbot.Button{
			fbot.Button{Text: b.StopNotifications, Payload: payload.Unsubscribe},
			fbot.Button{Text: b.SendFeedback, Payload: payload.Feedback},
			fbot.Button{Text: b.QuitHelp, Payload: payload.StartMenu},
		},
		Feedback: []fbot.Button{
			fbot.Button{Text: iconDelete + " " + b.CancelFeedback, Payload: payload.StartMenu},
		},
		AddMode: []fbot.Button{
			fbot.Button{Text: b.StopAdding, Payload: payload.StartMenu},
		},
		StudyMode: []fbot.Button{
			buttonStudyDone,
		},
		Show: []fbot.Button{
			buttonDelete,
			buttonStudyDone,
			fbot.Button{Text: iconShow + " " + b.ShowPhrase, Payload: payload.ShowStudy},
		},
		Score: []fbot.Button{
			buttonDelete,
			fbot.Button{Text: iconBad + " " + b.ScoreBad, Payload: payload.ScoreBad},
			fbot.Button{Text: iconOK, Payload: payload.ScoreOk},
			fbot.Button{Text: iconGood + " " + b.ScoreGood, Payload: payload.ScoreGood},
		},
		StudyEmpty: []fbot.Button{
			buttonAdd,
		},
		StudiesDue: []fbot.Button{
			buttonStudy,
			fbot.Button{Text: b.StudyNotNow, Payload: payload.StartMenu},
		},
		ConfirmDelete: []fbot.Button{
			fbot.Button{Text: iconDelete + " " + b.ConfirmDelete, Payload: payload.ConfirmDelete},
			fbot.Button{Text: b.CancelDelete, Payload: payload.CancelDelete},
		},
	}
}
