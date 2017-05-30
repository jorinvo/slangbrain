package messenger

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jorinvo/messenger"
	"github.com/jorinvo/slangbrain/brain"
)

// Config ...
type Config struct {
	Log            *log.Logger
	Token          string
	VerifyToken    string
	Store          brain.Store
	NotifyInterval time.Duration
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

	if err := client.GetStarted(payloadGetStarted); err != nil {
		return nil, fmt.Errorf("failed to enable Get Started button: %v", err)
	}
	config.Log.Printf("Get Started button activated")

	client.HandlePostBack(b.PostbackHandler)
	client.HandleRead(b.ReadHandler)
	client.HandleMessage(b.MessageHandler)

	if config.NotifyInterval > 0 {
		config.Log.Println("Notifications enabled")
		go func() {
			for range time.Tick(config.NotifyInterval) {
				config.Log.Println("Sending notifications")
				dueStudies, err := config.Store.GetDueStudies()
				if err != nil {
					config.Log.Println(err)
					return
				}
				now := time.Now()
				for chatID, count := range dueStudies {
					profile, err := client.ProfileByID(chatID)
					if err != nil {
						config.Log.Printf("Failed to get profile for %d: %v", chatID, err)
						continue
					}
					to := messenger.Recipient{ID: chatID}
					msg := fmt.Sprintf(messageStudiesDue, profile.FirstName, count)
					if err = client.SendWithReplies(to, msg, buttonsStudiesDue); err != nil {
						config.Log.Printf("Failed to notify user %d: %v", chatID, err)
					}
					b.trackActivity(chatID, now)
				}
			}
		}()
	}

	return client.Handler(), nil
}
