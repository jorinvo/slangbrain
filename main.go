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

	"github.com/jorinvo/slangbrain/admin"
	"github.com/jorinvo/slangbrain/brain"
	"github.com/jorinvo/slangbrain/fbot"
	"github.com/jorinvo/slangbrain/messenger"
)

const cliUsage = `Slangebrain Messenger bot

Usage: %s [flags]

Slangbrain uses BoltDB as a database.
Data is stored in a single file. No external system is needed.
However, only one application can access the database at a time.

Slangbrain starts a server to serve a webhook handler that can be registered as a Messenger bot.
The server is HTTP only and a proxy server should be used to make the bot available on
a public domain, preferably HTTPS only.

An admin server runs on a separate port.
It should be proxied and secured via HTTPS + basic auth.
The admin server provides an endpoint to fetch backups of the database.
Further, it provides an endpoint that can be registered as Slack Outgoing Webhook.

When users send feedback to the bot, the messages are forwarded to Slack
and admin replies in Slack are send back to the users.


Flags:
`

func main() {
	errorLogger := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
	infoLogger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)

	db := flag.String("db", "", "Required. Path to BoltDB file. Will be created if non-existent.")
	port := flag.Int("port", 8080, "Port Facebook webhook listens on.")
	verifyToken := flag.String("verify", "", "Required. Messenger bot verify token.")
	token := flag.String("token", "", "Required. Messenger bot token.")
	slackHook := flag.String("slackhook", "", "Required. URL of Slack Incoming Webhook. Used to send user messages to admin.")
	slackToken := flag.String("slacktoken", "", "Token for Slack Outgoing Webhook. Used to send admin answers to user messages.")
	adminPort := flag.Int("admin", 8081, "Port admin interface listens on.")
	notifyInterval := flag.Duration("notify", 0, "Interval of sending user notifications.")

	// Parse and validate flags
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, cliUsage, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if *db == "" {
		errorLogger.Println("Flag -db is required.")
		os.Exit(1)
	}
	if *token == "" {
		errorLogger.Println("Flag -token is required.")
		os.Exit(1)
	}
	if *verifyToken == "" {
		errorLogger.Println("Flag -verify is required.")
		os.Exit(1)
	}
	if *slackHook == "" {
		errorLogger.Println("Flag -slackhook is required.")
		os.Exit(1)
	}

	// Setup database
	store, err := brain.New(*db)
	if err != nil {
		errorLogger.Fatalln("failed to create store:", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			errorLogger.Println(err)
		}
	}()
	infoLogger.Printf("Database initialized: %s", *db)

	// Setup Facebook API client
	client := fbot.New(*token, *verifyToken)

	// Listen to system events for graceful shutdown
	shutdownSignals := make(chan os.Signal)
	signal.Notify(shutdownSignals, os.Interrupt)

	// Start admin server
	adminHandler := admin.New(store, *slackHook, *slackToken, errorLogger, client)
	aAddr := "localhost:" + strconv.Itoa(*adminPort)
	adminServer := &http.Server{Addr: aAddr, Handler: adminHandler}
	go func() {
		infoLogger.Printf("Admin server running at %s", aAddr)
		if err := adminServer.ListenAndServe(); err != nil {
			errorLogger.Fatalln("failed to start server:", err)
		}
	}()

	// Start Facebook webhook server
	handler, err := messenger.Run(messenger.Config{
		ErrorLogger:    errorLogger,
		InfoLogger:     infoLogger,
		Client:         client,
		Store:          store,
		NotifyInterval: *notifyInterval,
		MessageHandler: adminHandler.HandleMessage,
	})
	if err != nil {
		log.Fatalln("failed to start messenger:", err)
	}
	mAddr := "localhost:" + strconv.Itoa(*port)
	messengerServer := &http.Server{Addr: mAddr, Handler: handler}
	go func() {
		infoLogger.Printf("Messenger webhook server running at %s", mAddr)
		if err := messengerServer.ListenAndServe(); err != nil {
			errorLogger.Fatalln("failed to start server:", err)
		}
	}()

	// Wait for shutdown
	<-shutdownSignals
	infoLogger.Println("Waiting for connections before shutting down server.")
	if err = messengerServer.Shutdown(context.Background()); err != nil {
		errorLogger.Fatalln("failed to shutdown gracefully:", err)
	}
	if err = adminServer.Shutdown(context.Background()); err != nil {
		errorLogger.Fatalln("failed to shutdown gracefully:", err)
	}
	infoLogger.Println("Server gracefully stopped.")
}
