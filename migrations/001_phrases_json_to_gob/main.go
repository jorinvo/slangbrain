package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	bolt "github.com/coreos/bbolt"
)

type phrase struct {
	Phrase      string
	Explanation string
	Score       int
}

var bucketPhrases = []byte("phrases")

// Read phrases as JSON and save them as GOB.
// Is more efficient in space and time.
//
// Pass db file and a command as args.
// 1. Run json
// 2. Run migrate
// 3. Run gob
// 4. Diff results
func main() {
	dbFile := os.Args[1]
	cmd := os.Args[2]
	db, err := bolt.Open(dbFile, 0600, &bolt.Options{Timeout: 1 * time.Second})
	fatal(err)
	switch cmd {
	case "migrate":
		fatal(migrate(db))
	case "json":
		fatal(readJSON(db))
	case "gob":
		fatal(readGOB(db))
	default:
		log.Fatalln("command not found")
	}
}

func migrate(db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketPhrases)
		return b.ForEach(func(k []byte, v []byte) error {
			var p phrase
			if err := json.Unmarshal(v, &p); err != nil {
				return err
			}

			var buf bytes.Buffer
			if err := gob.NewEncoder(&buf).Encode(p); err != nil {
				return err
			}
			return b.Put(k, buf.Bytes())
		})
	})
}

func readJSON(db *bolt.DB) error {
	return db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketPhrases).ForEach(func(k []byte, v []byte) error {
			var p phrase
			if err := json.Unmarshal(v, &p); err != nil {
				return err
			}
			fmt.Println(k, p)
			return nil
		})
	})
}

func readGOB(db *bolt.DB) error {
	return db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketPhrases).ForEach(func(k []byte, v []byte) error {
			var p phrase
			if err := gob.NewDecoder(bytes.NewReader(v)).Decode(&p); err != nil {
				return err
			}
			fmt.Println(k, p)
			return nil
		})
	})
}

func fatal(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
