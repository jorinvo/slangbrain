package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/jorinvo/slangbrain/brain"

	_ "github.com/mattn/go-sqlite3"
	"github.com/paked/messenger"
)

var version string

const usage = `Slangebrain Messenger bot

Usage: %s [flags]

Flags:
`

type config struct {
	logger  *log.Logger
	verbose bool
	store   brain.Store
}

func main() {
	logger := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)

	versionFlag := flag.Bool("version", false, "Print binary version")
	// silent := flag.Bool("silent", false, "Suppress all output")
	sqlite := flag.String("sqlite", "", "")
	host := flag.String("host", "localhost", "")
	port := flag.Int("port", 8080, "")
	verifyToken := flag.String("verify", "", "")
	token := flag.String("token", "", "")

	// Parse args
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if *versionFlag {
		fmt.Printf("ghbackup %s %s %s\n", version, runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}
	if *sqlite == "" {
		logger.Println("Flag -sqlite is required.")
		os.Exit(1)
	}
	if *token == "" {
		logger.Println("Flag -token is required.")
		os.Exit(1)
	}
	if *verifyToken == "" {
		logger.Println("Flag -verify is required.")
		os.Exit(1)
	}

	store, err := brain.CreateStore("sqlite3", *sqlite)
	if err != nil {
		logger.Println("failed to create store: ", err)
		os.Exit(1)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			logger.Println(err)
		}
	}()

	client := messenger.New(messenger.Options{
		Verify:      true,
		VerifyToken: *verifyToken,
		Token:       *token,
	})

	// err = client.GreetingSetting("Welcome to Slangebrain!")
	// if err != nil {
	// 	log.Println("failed to set greeting:", err)
	// 	os.Exit(1)
	// }

	// TODO: show menu with modes: add and study
	// client.CallToActionsSetting

	// For beginning, always add mode. Just collecting phrases.
	client.HandleMessage(func(m messenger.Message, r *messenger.Response) {
		if m.IsEcho {
			return
		}

		parts := strings.SplitN(m.Text, "\n", 2)
		foreign := strings.TrimSpace(parts[0])
		mother := ""
		if len(parts) > 1 {
			mother = strings.TrimSpace(parts[1])
		}

		err := store.AddPhrase(m.Sender.ID, foreign, mother)
		if err != nil {
			log.Println("failed to save phrase:", err)
		}

		err = r.Text("Phrase saved. Add next one.")
		if err != nil {
			log.Println("failed to send message:", err)
		}
	})

	addr := *host + ":" + strconv.Itoa(*port)

	log.Printf("Server running at %s", addr)
	err = http.ListenAndServe(addr, client.Handler())
	if err != nil {
		log.Println("failed to start server:", err)
		os.Exit(1)
	}
}
