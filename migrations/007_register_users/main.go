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
	bucketModes         = []byte("modes")
	bucketRegisterDates = []byte("registerdates")
)

func main() {
	dbFile := os.Args[1]
	db, err := bolt.Open(dbFile, 0600, &bolt.Options{Timeout: 1 * time.Second})
	fatal(err)
	defer func() {
		fatal(db.Close())
	}()

	now := itob((time.Now().Unix()))

	fatal(db.Update(func(tx *bolt.Tx) error {
		br := tx.Bucket(bucketRegisterDates)

		return tx.Bucket(bucketModes).ForEach(func(k, _ []byte) error {

			if br.Get(k) == nil {
				fmt.Printf("register %d\n", btoi(k))
				return br.Put(k, now)
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
