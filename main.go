package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/jorinvo/slangbrain/api"
	"github.com/jorinvo/slangbrain/brain"
	"github.com/jorinvo/slangbrain/messenger"
	"github.com/jorinvo/slangbrain/messenger/webview"
	"github.com/jorinvo/slangbrain/slack"
	"github.com/jorinvo/slangbrain/translate"
)

const cliUsage = `Slangbrain Messenger bot

Usage: %s [flags]

Slangbrain uses BoltDB as a database.
Data is stored in a single file. No external system is needed.
However, only one application can access the database at a time.

Slangbrain starts a server to serve a webhook handler at /webhook that can be registered as a Messenger bot.
The server is HTTP only and a proxy server should be used to make the bot available on
a public domain, preferably HTTPS only.

Certain features are better done with a custom web view.
They are rendered at /webview.

/slack can be registered as Slack Outgoing Webhook.
When users send feedback to the bot, the messages are forwarded to Slack
and admin replies in Slack are send back to the users.

/backup provides an endpoint to fetch backups of the database.

Flags:
`

var version = "development"

func main() {
	errorLogger := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
	infoLogger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)

	versionFlag := flag.Bool("version", false, "Print the version of the binary.")
	db := flag.String("db", "", "Required. Path to BoltDB file. Will be created if non-existent.")
	port := flag.Int("port", 8080, "Port Facebook webhook listens on.")
	verifyToken := flag.String("verify", "", "Required. Messenger bot verify token.")
	token := flag.String("token", "", "Required. Messenger bot token.")
	secret := flag.String("secret", "", "Required. Facebook app secret.")
	slackHook := flag.String("slackhook", "", "Required. URL of Slack Incoming Webhook. Used to send user messages to admin.")
	slackToken := flag.String("slacktoken", "", "Token for Slack Outgoing Webhook. Used to send admin answers to user messages.")
	backupAuth := flag.String("backupauth", "", "/backup basic auth in the form user:pasword. If empty, /backup is deactivated.")
	baseURL := flag.String("url", "https://fbot.slangbrain.com", "Overwrite the base URL. Used for linking webviews.")
	noSetup := flag.Bool("nosetup", false, "Skip sending setup instructions to Facebook")

	// Parse and validate flags
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, cliUsage, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if *versionFlag {
		infoLogger.Println("Slangbrain", version)
		os.Exit(0)
	}

	if *db == "" {
		errorLogger.Println("Flag -db is required")
		os.Exit(1)
	}
	if *token == "" {
		errorLogger.Println("flag -token is required")
		os.Exit(1)
	}
	if *secret == "" {
		errorLogger.Println("flag -secret is required")
		os.Exit(1)
	}
	if *verifyToken == "" {
		errorLogger.Println("flag -verify is required")
		os.Exit(1)
	}
	if *slackHook == "" {
		errorLogger.Println("flag -slackhook is required")
		os.Exit(1)
	}

	// Setup database
	store, err := brain.New(*db)
	if err != nil {
		errorLogger.Fatalln("failed to create store:", err)
	}
	defer func() {
		if err = store.Close(); err != nil {
			errorLogger.Println(err)
		}
	}()
	infoLogger.Printf("Database initialized: %s", *db)

	translator := translate.New(strings.TrimSuffix(*baseURL, "/") + "/webview/manage/")

	// Listen to system events for graceful shutdown
	shutdownSignals := make(chan os.Signal, 1)
	signal.Notify(shutdownSignals, os.Interrupt)

	// Start Facebook webhook server
	feedback := make(chan messenger.Feedback)
	opts := []func(*messenger.Bot){
		messenger.Verify(*verifyToken),
		messenger.LogInfo(infoLogger),
		messenger.LogErr(errorLogger),
		messenger.GetFeedback(feedback),
		messenger.Notify,
		messenger.MessageDelay(2 * time.Second),
		messenger.Translate(translator),
	}
	if !*noSetup {
		opts = append(opts, messenger.Setup)
	}
	webhookHandler, err := messenger.New(
		store,
		*token,
		*secret,
		opts...,
	)
	if err != nil {
		errorLogger.Fatalln("failed to start messenger:", err)
	}

	slackHandler := slack.New(
		store,
		*slackHook,
		slack.Reply(*slackToken, webhookHandler.SendMessage),
		slack.LogErr(errorLogger),
	)
	go func() {
		for f := range feedback {
			slackHandler.HandleMessage(f.ChatID, f.Username, f.Message)
		}
	}()

	backupHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "invalid method: "+r.Method, http.StatusMethodNotAllowed)
			return
		}
		if u, p, ok := r.BasicAuth(); !ok || u+":"+p != *backupAuth {
			http.Error(w, "failed basic auth", http.StatusUnauthorized)
			return
		}
		store.BackupTo(w)
	})

	apiHandler := api.Phrases(store, errorLogger)
	webviewHandler := webview.New(store, errorLogger, translator, "/api/")

	mux := http.NewServeMux()
	mux.Handle("/webhook", webhookHandler)
	mux.Handle("/api/phrases/", http.StripPrefix("/api/phrases/", apiHandler))
	mux.Handle("/webview/manage/", http.StripPrefix("/webview/manage/", webviewHandler))
	mux.Handle("/slack", slackHandler)
	if *backupAuth != "" {
		mux.Handle("/backup", backupHandler)
	}

	mAddr := "localhost:" + strconv.Itoa(*port)
	messengerServer := &http.Server{Addr: mAddr, Handler: mux}

	go func() {
		infoLogger.Printf("Server running at %s", mAddr)
		if err = messengerServer.ListenAndServe(); err != nil {
			errorLogger.Fatalln("failed to start server:", err)
		}
	}()

	// Wait for shutdown
	<-shutdownSignals
	infoLogger.Println("Waiting for connections before shutting down server.")
	if err = messengerServer.Shutdown(context.Background()); err != nil {
		errorLogger.Fatalln("failed to shutdown gracefully:", err)
	}
	infoLogger.Println("Server gracefully stopped.")
}
