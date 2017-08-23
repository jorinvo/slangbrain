package translate

import (
	"strings"

	"github.com/jorinvo/slangbrain/fbot"
	"github.com/jorinvo/slangbrain/payload"
)

// Btn contains all button sets that can be sent to a user.
// They are already localized for one language.
type Btn struct {
	Help func(string) []fbot.Button
}

func newBtn(l labels, serverURL string) Btn {
	importHelp := fbot.PayloadButton(l.ImportHelp, payload.ImportHelp)
	normURL := strings.TrimSuffix(serverURL, "/")
	manager := normURL + "/webview/manage/"
	exporter := normURL + "/api/phrases.csv?token="

	return Btn{
		Help: func(token string) []fbot.Button {
			// Disable manage link if no location given
			if serverURL == "" || token == "" {
				return []fbot.Button{importHelp}
			}
			return []fbot.Button{
				fbot.URLButton(l.Manage, manager+token),
				importHelp,
				fbot.LinkButton(l.Export, exporter+token),
			}
		},
	}
}
