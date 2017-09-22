package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"time"

	bolt "github.com/coreos/bbolt"
)

var (
	bucketPhrases     = []byte("phrases")
	bucketScoretotals = []byte("scoretotals")
)

type phrase struct {
	Phrase      string
	Explanation string
	Score       int
}

func main() {
	dbFile := os.Args[1]
	db, err := bolt.Open(dbFile, 0600, &bolt.Options{Timeout: 1 * time.Second})
	fatal(err)
	defer func() {
		fatal(db.Close())
	}()

	fatal(db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucketScoretotals)
		if err != nil {
			return err
		}

		totals := map[int64]int{}

		// Sum scores per user
		err = tx.Bucket(bucketPhrases).ForEach(func(k, v []byte) error {
			var p phrase
			if err := gob.NewDecoder(bytes.NewReader(v)).Decode(&p); err != nil {
				return err
			}
			totals[btoi(k)] += p.Score
			return nil
		})
		if err != nil {
			return err
		}

		// Write scoretotals to bucket
		for k, v := range totals {
			fmt.Printf("id: %d; score: %6d\n", k, v)
			if err := b.Put(itob(k), itob(int64(v))); err != nil {
				return err
			}
		}

		return nil
	}))
}

func fatal(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func itob(v int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func btoi(b []byte) int64 {
	return int64(binary.BigEndian.Uint64(b))
}
