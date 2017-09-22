// Reset zeroscore for each user to the count of all phrases which are currently being studied and have a score of 0.
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
	bucketPhrases    = []byte("phrases")
	bucketStudytimes = []byte("studytimes")
	bucketZeroscores = []byte("zeroscores")
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

	scores := map[int64]int64{}

	fatal(db.Update(func(tx *bolt.Tx) error {
		bz := tx.Bucket(bucketZeroscores)
		bp := tx.Bucket(bucketPhrases)

		err = tx.Bucket(bucketStudytimes).ForEach(func(k, _ []byte) error {
			prefix := k[:8]

			var p phrase
			if err := gob.NewDecoder(bytes.NewReader(bp.Get(k))).Decode(&p); err != nil {
				return err
			}

			// Update zeroscore
			if p.Score == 0 {
				scores[btoi(prefix)]++
			}

			return nil
		})

		for k, v := range scores {
			fmt.Printf("\nid: %d; score: %d\n", k, v)
			if err := bz.Put(itob(k), itob(v)); err != nil {
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
