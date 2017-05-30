package messenger

import (
	"fmt"
	"log"
	"net/http"

	"github.com/jorinvo/messenger"
	"github.com/jorinvo/slangbrain/brain"
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
