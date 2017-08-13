package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"os"
	"time"

	"github.com/boltdb/bolt"
)

type profileData struct {
	Name      string
	Locale    string
	Timezone  float64
	CacheTime time.Time
}

func main() {
	dbFile := os.Args[1]
	db, err := bolt.Open(dbFile, 0600, &bolt.Options{Timeout: 1 * time.Second})
	fatal(err)
	defer func() {
		fatal(db.Close())
	}()

	fatal(db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("profiles"))
		return b.ForEach(func(k, v []byte) error {
			var p profileData
			// Read profile
			if err := gob.NewDecoder(bytes.NewReader(v)).Decode(&p); err != nil {
				return err
			}
			// Change language
			p.Locale = "de_DE"
			// Write back
			var buf bytes.Buffer
			if err := gob.NewEncoder(&buf).Encode(p); err != nil {
				return err
			}
			return b.Put(k, buf.Bytes())
		})
	}))
}

func fatal(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
