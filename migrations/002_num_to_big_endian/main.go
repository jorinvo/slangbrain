package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"time"

	bolt "github.com/coreos/bbolt"
)

// Update storage of all integers to be stored as big endian instead of varint.
// This makes the bytes sortable.
//
// - Go through all buckets and read data in memory
// - Delete buckets
// - Update data
// - Create new buckets
// - Write new data
func main() {
	dbFile := os.Args[1]
	db, err := bolt.Open(dbFile, 0600, &bolt.Options{Timeout: 1 * time.Second})
	fatal(err)
	fatal(migrate(db))
}

func itob(v int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func oldbtoi(b []byte) (int64, error) {
	return binary.ReadVarint(bytes.NewBuffer(b))
}

func migrate(db *bolt.DB) error {
	keyValUpdater := func(k, v []byte) ([]byte, []byte, error) {
		ik, err := oldbtoi(k)
		if err != nil {
			return nil, nil, err
		}
		iv, err := oldbtoi(v)
		if err != nil {
			return nil, nil, err
		}
		return itob(ik), itob(iv), nil
	}
	keyUpdater := func(k, v []byte) ([]byte, []byte, error) {
		ik, err := oldbtoi(k)
		if err != nil {
			return nil, nil, err
		}
		return itob(ik), v, nil
	}

	buckets := []struct {
		bucket  []byte
		updater func(k, v []byte) ([]byte, []byte, error)
	}{
		{
			bucket:  []byte("modes"),
			updater: keyValUpdater,
		},
		{
			bucket: []byte("phrases"),
			updater: func(k, v []byte) ([]byte, []byte, error) {
				id, err := oldbtoi(k[:8])
				if err != nil {
					return nil, nil, err
				}
				seq, err := oldbtoi(k[8:])
				if err != nil {
					return nil, nil, err
				}
				return append(itob(id), itob(seq)...), v, nil
			},
		},
		{
			bucket: []byte("studytimes"),
			updater: func(k, v []byte) ([]byte, []byte, error) {
				id, err := oldbtoi(k[:8])
				if err != nil {
					return nil, nil, err
				}
				seq, err := oldbtoi(k[8:])
				if err != nil {
					return nil, nil, err
				}
				iv, err := oldbtoi(v)
				if err != nil {
					return nil, nil, err
				}
				return append(itob(id), itob(seq)...), itob(iv), nil
			},
		},
		{
			bucket:  []byte("reads"),
			updater: keyValUpdater,
		},
		{
			bucket:  []byte("activities"),
			updater: keyValUpdater,
		},
		{
			bucket:  []byte("subscriptions"),
			updater: keyUpdater,
		},
		{
			bucket:  []byte("profiles"),
			updater: keyUpdater,
		},
		{
			bucket:  []byte("registerdates"),
			updater: keyValUpdater,
		},
	}

	return db.Update(func(tx *bolt.Tx) error {
		for _, bucket := range buckets {
			// Collect and update data
			b := tx.Bucket(bucket.bucket)
			if b == nil {
				fmt.Printf("no bucket: %s\n", bucket.bucket)
				continue
			}
			data := map[string]string{}
			err := b.ForEach(func(k []byte, v []byte) error {
				nk, nv, err := bucket.updater(k, v)
				data[string(nk)] = string(nv)
				return err
			})
			if err != nil {
				return err
			}
			// Delete bucket
			if err := tx.DeleteBucket(bucket.bucket); err != nil {
				return err
			}
			// Create new bucket
			bnew, err := tx.CreateBucket(bucket.bucket)
			if err != nil {
				return err
			}
			// Write data to bucket
			for k, v := range data {
				if err := bnew.Put([]byte(k), []byte(v)); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func fatal(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
