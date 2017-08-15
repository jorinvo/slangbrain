package translate

import (
	"github.com/jorinvo/slangbrain/fbot"
	"github.com/jorinvo/slangbrain/payload"
)

// Btn contains all button sets that can be sent to a user.
// They are already localized for one language.
type Btn struct {
	HelpEnable,
	HelpDisable []fbot.Button
}

func newBtn(l labels) Btn {
	// manage := fbot.URLButton(l.Manage, "https://fbot.slangbrain.com/webview/manage")
	feedback := fbot.PayloadButton(l.SendFeedback, payload.Feedback)

	return Btn{
		HelpEnable: []fbot.Button{
			// manage,
			fbot.PayloadButton(l.EnableNotifications, payload.Subscribe),
			feedback,
		},
		HelpDisable: []fbot.Button{
			// manage,
			fbot.PayloadButton(l.DisableNotifications, payload.Unsubscribe),
			feedback,
		},
	}
}
