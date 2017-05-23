package messenger

import (
	"log"
	"net/http"

	"github.com/jorinvo/messenger"
	"github.com/jorinvo/slangbrain/brain"
	"github.com/pkg/errors"
)

const (
	payloadAdd        = "PAYLOAD_ADD"
	payloadStudy      = "PAYLOAD_STUDY"
	payloadGetStarted = "PAYLOAD_GETSTARTED"
	payloadShow       = "PAYLOAD_SHOW"
	payloadScoreBad   = "PAYLOAD_SCOREBAD"
	payloadScoreOk    = "PAYLOAD_SCOREOK"
	payloadScoreGood  = "PAYLOAD_SCOREGOOD"
)

const (
	messageAdd = `Please send me a phrase and its explanation.
Separate them with a linebreak.`
	messageWelcome = `Welcome to Slangbrain!
...

` + messageAdd
	messageErr       = "Sorry, something went wrong."
	messageStudyDone = "Congrats, you finished all your studies for now!"
	messageButtons   = `Do you know what this means?

%s`
	greeting = "Welcome to Slangebrain!"
)

var (
	buttonsShow  = []messenger.QuickReply{button("show", payloadShow)}
	buttonsScore = []messenger.QuickReply{
		button("don't know", payloadScoreBad),
		button("soso", payloadScoreOk),
		button("easy", payloadScoreGood),
	}
)

// Config ...
type Config struct {
	Log         *log.Logger
	Token       string
	VerifyToken string
	Store       brain.Store
	Init        bool
}

type bot struct {
	store brain.Store
	log   *log.Logger
}

// Run ...
func Run(config Config) (http.Handler, error) {
	b := bot{
		store: config.Store,
		log:   config.Log,
	}

	client := messenger.New(messenger.Options{
		Verify:      true,
		VerifyToken: config.VerifyToken,
		Token:       config.Token,
	})

	if config.Init {

		err := client.SetGreeting([]messenger.Greeting{
			{Locale: "default", Text: greeting},
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to set greeting")
		}
		config.Log.Printf("Greeting set to: %s", greeting)

		err = client.GetStarted(payloadGetStarted)
		if err != nil {
			return nil, errors.Wrap(err, "failed to enable Get Started button")
		}
		config.Log.Printf("Get Started button activated")

		// Set menu
		err = client.PeristentMenu([]messenger.LocalizedMenu{{
			Locale:                "default",
			ComposerInputDisabled: true,
			Items: []messenger.MenuItem{
				menuItem("Add Phrases", payloadAdd),
				menuItem("Study", payloadStudy),
			},
		}})
		if err != nil {
			return nil, errors.Wrap(err, "failed to enable set menu")
		}
		config.Log.Printf("Persistent Menu loaded")

	}

	client.HandleMessage(b.MessageHandler)
	client.HandlePostBack(b.PostbackHandler)

	return client.Handler(), nil
}

func button(title, payload string) messenger.QuickReply {
	return messenger.QuickReply{
		ContentType: "text",
		Title:       title,
		Payload:     payload,
	}
}

func menuItem(title, payload string) messenger.MenuItem {
	return messenger.MenuItem{
		Type:    "postback",
		Title:   title,
		Payload: payload,
	}
}
