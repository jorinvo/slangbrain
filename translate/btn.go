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
	HelpDisable,
	HelpEnable,
	Feedback,
	AddMode,
	StudyMode,
	Show,
	Score,
	StudyEmpty,
	StudiesDue,
	ConfirmDelete []fbot.Button
}

// button labels have a 20 char limit
type buttonLabels struct {
	StudyDone,
	Study,
	Add,
	Done,
	Help,
	SubscribeConfirm,
	SubscribeDeny,
	DisableNotifications,
	EnableNotifications,
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
		buttonStudyDone = fbot.Button{Text: b.StudyDone, Payload: payload.Menu}
		buttonStudy     = fbot.Button{Text: iconStudy + " " + b.Study, Payload: payload.Study}
		buttonAdd       = fbot.Button{Text: iconAdd + " " + b.Add, Payload: payload.Add}
		buttonDone      = fbot.Button{Text: iconDone + " " + b.Done, Payload: payload.Idle}
		buttonHelp      = fbot.Button{Text: iconHelp + " " + b.Help, Payload: payload.Help}
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
			fbot.Button{Text: b.SubscribeDeny, Payload: payload.DenySubscribe},
		},
		HelpDisable: []fbot.Button{
			fbot.Button{Text: b.DisableNotifications, Payload: payload.Unsubscribe},
			fbot.Button{Text: b.SendFeedback, Payload: payload.Feedback},
			fbot.Button{Text: b.QuitHelp, Payload: payload.Menu},
		},
		HelpEnable: []fbot.Button{
			fbot.Button{Text: b.EnableNotifications, Payload: payload.Subscribe},
			fbot.Button{Text: b.SendFeedback, Payload: payload.Feedback},
			fbot.Button{Text: b.QuitHelp, Payload: payload.Menu},
		},
		Feedback: []fbot.Button{
			fbot.Button{Text: iconDelete + " " + b.CancelFeedback, Payload: payload.Menu},
		},
		AddMode: []fbot.Button{
			fbot.Button{Text: b.StopAdding, Payload: payload.Menu},
		},
		StudyMode: []fbot.Button{
			buttonStudyDone,
		},
		Show: []fbot.Button{
			buttonDelete,
			buttonStudyDone,
			fbot.Button{Text: iconShow + " " + b.ShowPhrase, Payload: payload.ShowPhrase},
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
			fbot.Button{Text: b.StudyNotNow, Payload: payload.Menu},
		},
		ConfirmDelete: []fbot.Button{
			fbot.Button{Text: iconDelete + " " + b.ConfirmDelete, Payload: payload.ConfirmDelete},
			fbot.Button{Text: b.CancelDelete, Payload: payload.CancelDelete},
		},
	}
}
