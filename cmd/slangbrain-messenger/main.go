package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"

	"github.com/jorinvo/slangbrain/brain"
	"github.com/jorinvo/slangbrain/messenger"
	_ "github.com/mattn/go-sqlite3"
)

var version string

const cliUsage = `Slangebrain Messenger bot

Usage: %s [flags]

Flags:
`

func main() {
	logger := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)

	versionFlag := flag.Bool("version", false, "Print binary version")
	// silent := flag.Bool("silent", false, "Suppress all output")
	sqlite := flag.String("sqlite", "", "")
	host := flag.String("host", "localhost", "")
	port := flag.Int("port", 8080, "")
	verifyToken := flag.String("verify", "", "")
	token := flag.String("token", "", "")
	init := flag.Bool("init", false, "Pass to setup Bot setting on startup.")

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
		logger.Fatalln("failed to create store:", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			logger.Println(err)
		}
	}()
	logger.Printf("Database initialized: %s", *sqlite)

	handler, err := messenger.Run(messenger.Config{
		Log:         logger,
		Token:       *token,
		VerifyToken: *verifyToken,
		Store:       store,
		Init:        *init,
	})
	if err != nil {
		log.Fatalln("failed to start messenger:", err)
	}

	addr := *host + ":" + strconv.Itoa(*port)

	logger.Printf("Server running at %s", addr)
	err = http.ListenAndServe(addr, handler)
	if err != nil {
		logger.Fatalln("failed to start server:", err)
	}
}
