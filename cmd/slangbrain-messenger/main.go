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
	_ "github.com/lib/pq"
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
	db := flag.String("db", "", "")
	host := flag.String("host", "localhost", "")
	port := flag.Int("port", 8080, "")
	verifyToken := flag.String("verify", "", "")
	token := flag.String("token", "", "")

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
		logger.Println("Flag -db is required.")
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

	store, err := brain.CreateStore("postgres", *db)
	if err != nil {
		logger.Fatalln("failed to create store:", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			logger.Println(err)
		}
	}()
	logger.Printf("Database initialized: %s", *db)

	handler, err := messenger.Run(messenger.Config{
		Log:         logger,
		Token:       *token,
		VerifyToken: *verifyToken,
		Store:       store,
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
