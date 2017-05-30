package messenger

import (
	"fmt"
	"log"
	"net/http"

	"github.com/jorinvo/messenger"
	"github.com/jorinvo/slangbrain/brain"
)

const (
	payloadIdle       = "PAYLOAD_IDLE"
	payloadShowHelp   = "PAYLOAD_SHOWHELP"
	payloadStartMenu  = "PAYLOAD_STARTMENU"
	payloadStartAdd   = "PAYLOAD_STARTADD"
	payloadStartStudy = "PAYLOAD_STARTSTUDY"
	payloadGetStarted = "PAYLOAD_GETSTARTED"
	payloadShowStudy  = "PAYLOAD_SHOWSTUDY"
	payloadScoreBad   = "PAYLOAD_SCOREBAD"
	payloadScoreOk    = "PAYLOAD_SCOREOK"
	payloadScoreGood  = "PAYLOAD_SCOREGOOD"
)

const (
	messageStartMenu = `What would you like to do next?
Please use the buttons below.`
	messageHelp = `this is help
`
	messageIdle     = "Good, just send me a message to continue with your studies."
	messageStartAdd = `Please send me a phrase and its explanation.
Separate them with a linebreak.`
	messageWelcome = `Welcome!
Slangebrain is here to help you with your language studies.
Whenever you pick up a new phrase, just add it to your Slangebrain and remember it forever.
Master the language you encounter in your every day life instead of being limited to a textbook.`
	messageWelcome2 = `You begin by adding phrases and after Slangbrain will test your memories in a natural schedule.

` + messageStartAdd
	messageErr              = "Sorry, something went wrong."
	messageExplanationEmpty = "The phrase is missing an explanation. Please send it again with explanation."
	messagePhraseEmpty      = "Please send a phrase."
	messageStudyDone        = `Congrats, you finished all your studies for now!
Come back in %s.`
	messageStudyCorrect = "Nice, your answer was correct!"
	messageStudyWrong   = `Sorry, the correct version is:

%s

`
	messageStudyEmpty = `You have added no phrases yet.
Click the button below and get started.`
	messageStudyQuestion = `Do you remember how to say this?

%s

Use the buttons or type the phrase.`
	messageExplanationExists = `You already saved a phrase with the same explanation:
%s
%s

Please send it again with an explanation you can distinguish from the existing one.`
	messageAddDone = `Saved phrase:
%s

With explanation:
%s`
	messageAddNext = "Add next phrase."
	greeting       = `Slangbrain helps you with our language Studies.
Master the language you encounter in your every day life instead of being limited to a textbook.`
)

var (
	buttonStudyDone = button("done studying", payloadStartMenu)
	// Teacher emoji
	buttonStudy = button("\U0001F468\U0000200D\U0001F3EB study", payloadStartStudy)
	// Plus sign emoji
	buttonAdd = button("\u2795 phrases", payloadStartAdd)
	// Waving hand emoji
	buttonDone = button("\u2714 done", payloadIdle)
)

var (
	buttonsMenuMode = []messenger.QuickReply{
		buttonStudy,
		buttonAdd,
		button("\u2753 help", payloadShowHelp),
		buttonDone,
	}
	buttonsHelp = []messenger.QuickReply{
		buttonStudy,
		buttonAdd,
		buttonDone,
	}
	buttonsAddMode = []messenger.QuickReply{
		button("stop adding", payloadStartMenu),
	}
	buttonsStudyMode = []messenger.QuickReply{
		buttonStudyDone,
	}
	buttonsShow = []messenger.QuickReply{
		buttonStudyDone,
		button("\U0001F449 show phrase", payloadShowStudy),
	}
	buttonsScore = []messenger.QuickReply{
		// Ok thumb down emoji
		button("\U0001F44E didn't know", payloadScoreBad),
		button("soso", payloadScoreOk),
		// Ok hand emoji
		button("\U0001F44C easy", payloadScoreGood),
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
	store  brain.Store
	log    *log.Logger
	client *messenger.Messenger
}

// Run ...
func Run(config Config) (http.Handler, error) {
	client := messenger.New(messenger.Options{
		Verify:      true,
		VerifyToken: config.VerifyToken,
		Token:       config.Token,
	})

	b := bot{
		store:  config.Store,
		log:    config.Log,
		client: client,
	}

	err := client.SetGreeting([]messenger.Greeting{
		{Locale: "default", Text: greeting},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set greeting: %v", err)
	}
	config.Log.Println("Greeting set")

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
