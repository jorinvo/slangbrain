package messenger

import (
	"fmt"
	"log"
	"net/http"

	"github.com/jorinvo/messenger"
	"github.com/jorinvo/slangbrain/brain"
)

const (
	payloadStartIdle  = "PAYLOAD_STARTIDLE"
	payloadStartAdd   = "PAYLOAD_STARTADD"
	payloadStartStudy = "PAYLOAD_STARTSTUDY"
	payloadGetStarted = "PAYLOAD_GETSTARTED"
	payloadShow       = "PAYLOAD_SHOW"
	payloadScoreBad   = "PAYLOAD_SCOREBAD"
	payloadScoreOk    = "PAYLOAD_SCOREOK"
	payloadScoreGood  = "PAYLOAD_SCOREGOOD"
)

const (
	messageStartIdle = "What would you like to do next?"
	messageStartAdd  = `Please send me a phrase and its explanation.
Separate them with a linebreak.`
	messageWelcome = `Welcome to Slangbrain!
...

` + messageStartAdd
	messageErr            = "Sorry, something went wrong."
	messageErrExplanation = "The phrase is missing an explanation. Please send it again with explanation."
	messageStudyDone      = "Congrats, you finished all your studies for now!"
	messageStudyQuestion  = `Do you know what this means?

%s`
	messagePhraseExists = `You already saved this phrase before:
%s
%s
`
	messageExplanationExists = `You already saved a phrase with the same explanation:
%s
%s
`
	greeting = "Welcome to Slangebrain!"
)

var (
	buttonStudyDone = button("done studying", payloadStartIdle)
)

var (
	buttonsIdleMode = []messenger.QuickReply{
		button("study", payloadStartStudy),
		button("add phrases", payloadStartAdd),
	}
	buttonsAddMode = []messenger.QuickReply{
		button("done adding", payloadStartIdle),
	}
	buttonsStudyMode = []messenger.QuickReply{
		buttonStudyDone,
	}
	buttonsShow = []messenger.QuickReply{
		buttonStudyDone,
		button("show", payloadShow),
	}
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

	err := client.SetGreeting([]messenger.Greeting{
		{Locale: "default", Text: greeting},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set greeting: %v", err)
	}
	config.Log.Printf("Greeting set to: %s", greeting)

	err = client.GetStarted(payloadGetStarted)
	if err != nil {
		return nil, fmt.Errorf("failed to enable Get Started button: %v", err)
	}
	config.Log.Printf("Get Started button activated")

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
