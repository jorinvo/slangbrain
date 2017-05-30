package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jorinvo/slangbrain/brain"
	"github.com/jorinvo/slangbrain/messenger"
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
	db := flag.String("db", "", "required")
	host := flag.String("host", "localhost", "")
	port := flag.Int("port", 8080, "")
	verifyToken := flag.String("verify", "", "required unless import")
	token := flag.String("token", "", "required unless import")
	backupDir := flag.String("backup", "", "Directory to write backups to. When not set, backups are disabled.")
	toImport := flag.String("import", "", "")
	studynow := flag.Bool("studynow", false, "")
	rmChat := flag.Int64("rmchat", 0, "")

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

	store, err := brain.CreateStore(*db)
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

	if *studynow {
		studyNow(store, logger)
		return
	}

	if *toImport != "" {
		csvImport(store, logger, *toImport)
		return
	}

	if *rmChat != 0 {
		deleteChat(store, logger, *rmChat)
		return
	}

	if *token == "" {
		logger.Println("Flag -token is required.")
		os.Exit(1)
	}
	if *verifyToken == "" {
		logger.Println("Flag -verify is required.")
		os.Exit(1)
	}

	handler, err := messenger.Run(messenger.Config{
		Log:         logger,
		Token:       *token,
		VerifyToken: *verifyToken,
		Store:       store,
	})
	if err != nil {
		log.Fatalln("failed to start messenger:", err)
	}

	// Listen to system events
	addr := *host + ":" + strconv.Itoa(*port)
	shutdownSignals := make(chan os.Signal)
	alrmSignals := make(chan os.Signal)
	// CTRL-C to shutdown
	signal.Notify(shutdownSignals, os.Interrupt)
	signal.Notify(alrmSignals, syscall.SIGALRM)
	h := &http.Server{Addr: addr, Handler: handler}

	// Start server
	go func() {
		logger.Printf("Server running at %s", addr)
		err = h.ListenAndServe()
		if err != nil {
			logger.Fatalln("failed to start server:", err)
		}
	}()

	// Backup DB on SIGALRM, if -backup is set
	if *backupDir != "" {
		go func() {
			for {
				<-alrmSignals
				logger.Println("Backup triggered")
				filename := "slangbrain-" + strconv.Itoa(int(time.Now().Unix())) + ".db"
				file := filepath.Join(*backupDir, filename)
				if err := store.BackupTo(file); err != nil {
					logger.Printf("Error while backing up files: %v", err)
				} else {
					logger.Printf("Backup successfully written to %s", file)
				}
			}
		}()
		logger.Printf("Backups enabled with directory: %s", *backupDir)
	}

	<-shutdownSignals
	logger.Println("Waiting for connections before shutting down server.")
	if err = h.Shutdown(context.Background()); err != nil {
		logger.Fatalln("failed to shutdown gracefully:", err)
	}
	logger.Println("Server gracefully stopped.")
}

func csvImport(store brain.Store, logger *log.Logger, toImport string) {
	// CSV import
	parts := strings.Split(toImport, ":")
	i, err := strconv.Atoi(parts[0])
	if err != nil {
		logger.Fatal(err)
	}
	chatID := int64(i)
	file := parts[1]
	logger.Printf("Importing to chat ID %d from CSV file %s", chatID, file)
	f, err := os.Open(file)
	if err != nil {
		logger.Fatalln(err)
	}
	count := 0
	reader := csv.NewReader(f)
	for {
		row, err := reader.Read()
		if err == io.EOF {
			logger.Printf("Imported %d phrases", count)
			return
		}
		if err != nil {
			logger.Fatalln(err)
		}
		if len(row) != 2 {
			logger.Printf("Line %d has wrong number of fields. Expected 2, had %d.", count+1, len(row))
		} else {
			count++
			p := strings.TrimSpace(row[0])
			e := strings.TrimSpace(row[1])
			if err = store.AddPhrase(chatID, p, e); err != nil {
				logger.Fatalln(err)
			}
		}
	}
}

func studyNow(store brain.Store, logger *log.Logger) {
	logger.Println("Study now - Setting all study times")
	if err := store.StudyNow(); err != nil {
		logger.Println("Failed to update study times", err)
	}
}

func deleteChat(store brain.Store, logger *log.Logger, id int64) {
	logger.Println("deleting all data for chat with id:", id)
	if err := store.DeleteChat(id); err != nil {
		logger.Println(err)
	}
}
