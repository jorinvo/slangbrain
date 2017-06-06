package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"

	"github.com/jorinvo/slangbrain/admin"
	"github.com/jorinvo/slangbrain/brain"
	"github.com/jorinvo/slangbrain/messenger"
)

var version string

const cliUsage = `Slangebrain Messenger bot

Usage: %s [flags]

Flags:
`

func main() {
	errorLogger := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
	infoLogger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)

	versionFlag := flag.Bool("version", false, "Print binary version")
	// silent := flag.Bool("silent", false, "Suppress all output")
	db := flag.String("db", "", "required")
	port := flag.Int("port", 8080, "")
	verifyToken := flag.String("verify", "", "required unless import")
	token := flag.String("token", "", "required unless import")
	adminPort := flag.Int("admin", 8081, "")
	notifyInterval := flag.Duration("notify", 0, "")

	// Parse args
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, cliUsage, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if *versionFlag {
		fmt.Printf("ghbackup %s %s %s\n", version, runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}
	if *db == "" {
		errorLogger.Println("Flag -db is required.")
		os.Exit(1)
	}

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

	if *token == "" {
		errorLogger.Println("Flag -token is required.")
		os.Exit(1)
	}
	if *verifyToken == "" {
		errorLogger.Println("Flag -verify is required.")
		os.Exit(1)
	}

	handler, err := messenger.Run(messenger.Config{
		ErrorLogger:    errorLogger,
		InfoLogger:     infoLogger,
		Token:          *token,
		VerifyToken:    *verifyToken,
		Store:          store,
		NotifyInterval: *notifyInterval,
	})
	if err != nil {
		log.Fatalln("failed to start messenger:", err)
	}

	// Listen to system events
	shutdownSignals := make(chan os.Signal)
	// CTRL-C to shutdown
	signal.Notify(shutdownSignals, os.Interrupt)

	// Start server
	mAddr := "localhost:" + strconv.Itoa(*port)
	messengerServer := &http.Server{Addr: mAddr, Handler: handler}
	go func() {
		infoLogger.Printf("Messenger webhook server running at %s", mAddr)
		if err := messengerServer.ListenAndServe(); err != nil {
			errorLogger.Fatalln("failed to start server:", err)
		}
	}()

	adminHandler := admin.New(store, errorLogger)
	aAddr := "localhost:" + strconv.Itoa(*adminPort)
	adminServer := &http.Server{Addr: aAddr, Handler: adminHandler}
	go func() {
		infoLogger.Printf("Admin server running at %s", aAddr)
		if err := adminServer.ListenAndServe(); err != nil {
			errorLogger.Fatalln("failed to start server:", err)
		}
	}()

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
