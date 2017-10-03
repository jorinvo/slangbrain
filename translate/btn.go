package translate

import (
	"strings"

	"qvl.io/fbot"
)

// Btn contains all button sets that can be sent to a user.
// They are already localized for one language.
type Btn struct {
	Help func(string) []fbot.Button
}

func newBtn(l labels, serverURL string) Btn {
	homepage := fbot.LinkButton(l.Homepage, l.BlogURL)
	normURL := strings.TrimSuffix(serverURL, "/")
	manager := normURL + "/webview/manage/"
	exporter := normURL + "/api/phrases.csv?token="

	return Btn{
		Help: func(token string) []fbot.Button {
			// Disable manage link if no location given
			if serverURL == "" || token == "" {
				return []fbot.Button{homepage}
			}
			return []fbot.Button{
				fbot.URLButton(l.Manage, manager+token),
				fbot.LinkButton(l.Export, exporter+token),
				homepage,
			}
		},
	}
}
