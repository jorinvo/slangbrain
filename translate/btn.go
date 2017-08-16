package translate

import (
	"github.com/jorinvo/slangbrain/fbot"
	"github.com/jorinvo/slangbrain/payload"
)

// Btn contains all button sets that can be sent to a user.
// They are already localized for one language.
type Btn struct {
	Help func(bool, string) []fbot.Button
}

func newBtn(l labels, managerLocation string) Btn {
	feedback := fbot.PayloadButton(l.SendFeedback, payload.Feedback)

	return Btn{
		Help: func(isSubscribed bool, manageToken string) []fbot.Button {
			var notificationToggle fbot.Button
			if isSubscribed {
				notificationToggle = fbot.PayloadButton(l.DisableNotifications, payload.Unsubscribe)
			} else {
				notificationToggle = fbot.PayloadButton(l.EnableNotifications, payload.Subscribe)
			}
			// Disable manage link if no location given
			if managerLocation == "" || manageToken == "" {
				return []fbot.Button{
					notificationToggle,
					feedback,
				}
			}
			return []fbot.Button{
				fbot.URLButton(l.Manage, managerLocation+manageToken),
				notificationToggle,
				feedback,
			}
		},
	}
}
