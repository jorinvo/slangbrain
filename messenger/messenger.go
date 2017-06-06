package messenger

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jorinvo/slangbrain/brain"
	"github.com/jorinvo/slangbrain/fbot"
)

// Config ...
type Config struct {
	ErrorLogger    *log.Logger
	InfoLogger     *log.Logger
	Store          brain.Store
	Client         fbot.Client
	NotifyInterval time.Duration
	MessageHandler func(int64, string, string) error
}

type bot struct {
	store          brain.Store
	err            *log.Logger
	info           *log.Logger
	client         fbot.Client
	messageHandler func(int64, string, string) error
}

// Run ...
func Run(config Config) (http.Handler, error) {
	b := bot{
		store:          config.Store,
		err:            config.ErrorLogger,
		info:           config.InfoLogger,
		client:         config.Client,
		messageHandler: config.MessageHandler,
	}

	if err := config.Client.SetGreetings(map[string]string{"default": greeting}); err != nil {
		return nil, fmt.Errorf("failed to set greeting: %v", err)
	}
	config.InfoLogger.Println("Greeting set")

	if err := config.Client.SetGetStartedPayload(payloadGetStarted); err != nil {
		return nil, fmt.Errorf("failed to enable Get Started button: %v", err)
	}
	config.InfoLogger.Printf("Get Started button activated")

	if config.NotifyInterval > 0 {
		config.InfoLogger.Println("Notifications enabled")
		go func() {
			for range time.Tick(config.NotifyInterval) {
				config.InfoLogger.Println("Sending notifications")
				dueStudies, err := config.Store.GetDueStudies()
				if err != nil {
					config.ErrorLogger.Println(err)
					return
				}
				now := time.Now()
				for chatID, count := range dueStudies {
					p, err := config.Client.GetProfile(chatID)
					name := p.FirstName
					if err != nil {
						name = "there"
						config.ErrorLogger.Printf("failed to get profile for %d: %v", chatID, err)
					}
					msg := fmt.Sprintf(messageStudiesDue, name, count)
					if err = config.Client.Send(chatID, msg, buttonsStudiesDue); err != nil {
						config.ErrorLogger.Printf("failed to notify user %d: %v", chatID, err)
					}
					b.trackActivity(chatID, now)
				}
			}
		}()
	}

	return config.Client.Webhook(b.HandleEvent), nil
}
