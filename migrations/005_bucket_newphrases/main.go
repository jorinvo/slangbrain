package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/boltdb/bolt"
)

const maxNewStudies = 30

var (
	bucketPhrases    = []byte("phrases")
	bucketStudytimes = []byte("studytimes")
	bucketNewPhrases = []byte("newphrases")
	bucketReads      = []byte("reads")
	bucketZeroscores = []byte("zeroscores")
)

type Phrase struct {
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

	now := time.Now()

	fatal(db.Update(func(tx *bolt.Tx) error {
		bn, err := tx.CreateBucketIfNotExists(bucketNewPhrases)
		if err != nil {
			return err
		}

		bs := tx.Bucket(bucketStudytimes)

		// For each phrase, for each user
		return tx.Bucket(bucketReads).ForEach(func(prefix, _ []byte) error {
			c := tx.Bucket(bucketPhrases).Cursor()
			i := 0
			pc := 0
			var newPhrases []byte
			for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
				var p Phrase
				if err := gob.NewDecoder(bytes.NewReader(v)).Decode(&p); err != nil {
					return err
				}

				// Count phrases for logging
				pc++

				if p.Score > 0 {
					continue
				}

				i++

				// Reschedule scheduled zero studies
				if i < maxNewStudies {
					next := itob(now.Add(2 * time.Hour).Unix())
					if err := bs.Put(k, next); err != nil {
						return err
					}
					continue
				}

				// Collect new phrases
				newPhrases = append(newPhrases, k[8:]...)

				// Delete study time of phrase
				if err := bs.Delete(k); err != nil {
					return err
				}
			}

			// Save new phrases to bucket
			if err := bn.Put(prefix, newPhrases); err != nil {
				return err
			}

			// Update zeroscores bucket
			bz := tx.Bucket(bucketZeroscores)
			var zeroscore int64
			if v := bz.Get(prefix); v != nil {
				zeroscore = btoi(v)
			}

			fmt.Printf("%v: total: %4d, new: %4d, zeroscore: %3d actualzeros: %3d\n", btoi(prefix), pc, len(newPhrases)/8, zeroscore, i)

			if zeroscore > maxNewStudies {
				zeroscore = maxNewStudies
			}
			if zeroscore > int64(i) {
				zeroscore = int64(i)
			}

			return bz.Put(prefix, itob(zeroscore))
		})
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
