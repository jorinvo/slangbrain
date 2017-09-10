package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/boltdb/bolt"
)

var (
	bucketPhrases        = []byte("phrases")
	bucketStudytimes     = []byte("studytimes")
	bucketPhraseAddTimes = []byte("phraseaddtimes")
	bucketNewPhrases     = []byte("newphrases")
)

func main() {
	dbFile := os.Args[1]
	db, err := bolt.Open(dbFile, 0600, &bolt.Options{Timeout: 1 * time.Second})
	fatal(err)
	defer func() {
		fatal(db.Close())
	}()

	now := itob(time.Now().Unix())

	fatal(db.Update(func(tx *bolt.Tx) error {
		bs := tx.Bucket(bucketStudytimes)
		bp := tx.Bucket(bucketPhrases)
		ba := tx.Bucket(bucketPhraseAddTimes)
		bn := tx.Bucket(bucketNewPhrases)

		err := bs.ForEach(func(k, _ []byte) error {
			// Remove studytime for non-existend phrases
			if bp.Get(k) == nil {
				m := ""
				if ba.Get(k) != nil {
					m = "and add time "
					if err := ba.Delete(k); err != nil {
						return err
					}
				}
				fmt.Printf("remove study %sfor key %x (id %d, seq %d)\n", m, k, btoi(k[:8]), btoi(k[8:]))
				return bs.Delete(k)
			}

			// Remove phrases from new phrases if they are already scheduled for studying
			if bn.Get(k) != nil {
				fmt.Printf("remove scheduled phrase from new phrases: key %x (id %d, seq %d)\n", k, btoi(k[:8]), btoi(k[8:]))
				return bn.Delete(k)
			}

			return nil
		})
		if err != nil {
			return err
		}

		// Add missing add times
		return bp.ForEach(func(k, _ []byte) error {
			if ba.Get(k) == nil {
				fmt.Printf("add missing add time for key %x (id %d, seq %d)\n", k, btoi(k[:8]), btoi(k[8:]))
				return ba.Put(k, now)
			}
			return nil
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
