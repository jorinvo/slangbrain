package main

import (
	"log"
	"os"
	"time"

	"github.com/boltdb/bolt"
)

var (
	bucketZeroscores = []byte("zeroscores")
)

func main() {
	dbFile := os.Args[1]
	db, err := bolt.Open(dbFile, 0600, &bolt.Options{Timeout: 1 * time.Second})
	fatal(err)
	defer func() {
		fatal(db.Close())
	}()

	fatal(db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket(bucketZeroscores)
	}))
}

func fatal(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
