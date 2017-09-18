package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/jorinvo/slangbrain/api"
	"github.com/jorinvo/slangbrain/brain"
	"github.com/jorinvo/slangbrain/messenger"
	"github.com/jorinvo/slangbrain/messenger/webview"
	"github.com/jorinvo/slangbrain/slack"
	"github.com/jorinvo/slangbrain/translate"
	"golang.org/x/crypto/acme/autocert"
)

const cliUsage = `Slangbrain Messenger bot

Usage: %s [flags]

Slangbrain uses BoltDB as a database.
Data is stored in a single file. No external system is needed.
However, only one application can access the database at a time.

Slangbrain starts a server to serve a webhook handler at /webhook that can be registered as a Messenger bot.
If -http PORT is passed an HTTP-only server is started. Otherwise a production server listening on :80 and :443 is started,
redirecting all traffic to HTTPS. Let's Encrypt is used for automatic certificate loading.

Certain features are better done with a custom web view.
They are rendered at /webview.

/slack can be registered as Slack Outgoing Webhook.
When users send feedback to the bot, the messages are forwarded to Slack
and admin replies in Slack are send back to the users.

/backup provides an endpoint to fetch backups of the database.

Flags:
`

const (
	readTimeout  = 5 * time.Second
	writeTimeout = 10 * time.Second
	idleTimeout  = 120 * time.Second
)

var version = "development"

func main() {
	errorLogger := log.New(os.Stderr, "", 0)
	infoLogger := log.New(os.Stdout, "", 0)

	versionFlag := flag.Bool("version", false, "Print the version of the binary.")
	db := flag.String("db", "", "Required. Path to BoltDB file. Will be created if non-existent.")
	httpPort := flag.Int("http", -1, "Address http server listens on. If given, runs http only. If empty runs https on 443 and redirects http on 80.")
	email := flag.String("email", "", "Requrired unless -http. Email address to use as contact for Let's Encrypt.")
	certCache := flag.String("certdir", "", "Requrired unless -http. Directory to cache certificates.")
	verifyToken := flag.String("verify", "", "Required. Messenger bot verify token.")
	token := flag.String("token", "", "Required. Messenger bot token.")
	secret := flag.String("secret", "", "Required. Facebook app secret.")
	slackHook := flag.String("slackhook", "", "Required. URL of Slack Incoming Webhook. Used to send user messages to admin.")
	slackToken := flag.String("slacktoken", "", "Token for Slack Outgoing Webhook. Used to send admin answers to user messages.")
	backupAuth := flag.String("backupauth", "", "/backup basic auth in the form user:pasword. If empty, /backup is deactivated.")
	domain := flag.String("domain", "fbot.slangbrain.com", "Domain used for certs and internal links.")
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

	if *httpPort == 0 {
		if *email == "" {
			errorLogger.Fatalln("Flag -email is required")
		}
		if *certCache == "" {
			errorLogger.Fatalln("Flag -certdir is required")
		}
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

	translator := translate.New("https://" + *domain)

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
			slackHandler.HandleMessage(f.ChatID, f.Username, f.Message, f.Channel)
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
	csvHandler := api.CSV(store, errorLogger)
	webviewHandler := webview.New(store, errorLogger, translator, "/api/")

	mux := http.NewServeMux()
	mux.Handle("/webhook", webhookHandler)
	mux.Handle("/api/phrases.csv", csvHandler)
	mux.Handle("/api/phrases", http.StripPrefix("/api/phrases", apiHandler))
	mux.Handle("/api/phrases/", http.StripPrefix("/api/phrases/", apiHandler))
	mux.Handle("/webview/manage/", http.StripPrefix("/webview/manage/", webviewHandler))
	mux.Handle("/slack", slackHandler)
	mux.Handle("/status", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "OK")
	}))
	if *backupAuth != "" {
		mux.Handle("/backup", backupHandler)
	}
	handler := gziphandler.GzipHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000;")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self' 'unsafe-inline';")
		w.Header().Set("X-Frame-Options", "DENY")
		mux.ServeHTTP(w, r)
	}))

	httpAddr := ":" + strconv.Itoa(*httpPort)
	httpHandler := handler
	var httpsServer *http.Server

	if *httpPort < 0 {
		httpAddr = ":80"
		httpHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
		})

		certManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(*domain),
			Email:      *email,
			Cache:      autocert.DirCache(*certCache),
		}

		httpsServer = &http.Server{
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
			ErrorLog:     infoLogger,
			Addr:         ":443",
			Handler:      handler,
			TLSConfig: &tls.Config{
				GetCertificate: certManager.GetCertificate,
			},
		}

		go func() {
			infoLogger.Printf("HTTPS server running at %s for domain %s", httpsServer.Addr, *domain)
			if err = httpsServer.ListenAndServeTLS("", ""); err != nil {
				errorLogger.Fatalln("failed to start HTTPS server:", err)
			}
		}()
	}

	httpServer := &http.Server{
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
		ErrorLog:     infoLogger,
		Addr:         httpAddr,
		Handler:      httpHandler,
	}

	go func() {
		infoLogger.Printf("HTTP server running at %s", httpServer.Addr)
		if err = httpServer.ListenAndServe(); err != nil {
			errorLogger.Fatalln("failed to start HTTP server:", err)
		}
	}()

	// Wait for shutdown
	<-shutdownSignals
	infoLogger.Println("Waiting for connections before shutting down server.")
	if err = httpServer.Shutdown(context.Background()); err != nil {
		errorLogger.Fatalln("failed to shutdown http server gracefully:", err)
	}
	if httpsServer != nil {
		if err = httpsServer.Shutdown(context.Background()); err != nil {
			errorLogger.Fatalln("failed to shutdown https server gracefully:", err)
		}
	}
	infoLogger.Println("Server gracefully stopped.")
}
