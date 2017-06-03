package brain

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"time"

	"github.com/boltdb/bolt"
)

// Store ...
type Store struct {
	db *bolt.DB
}

// New returns a new Store with a database already setup.
func New(dbFile string) (Store, error) {
	db, err := bolt.Open(dbFile, 0600, &bolt.Options{Timeout: 1 * time.Second})
	store := Store{db}
	if err != nil {
		return store, fmt.Errorf("failed to open database: %v", err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		for _, bucket := range buckets {
			_, err = tx.CreateBucketIfNotExists(bucket)
			if err != nil {
				return fmt.Errorf("failed to create bucket '%s': %v", bucket, err)
			}
		}
		return nil
	})
	if err != nil {
		return store, fmt.Errorf("failed to initialize buckets: %v", err)
	}

	rand.Seed(time.Now().UnixNano())

	return store, err
}

// Close the underlying database connection.
func (store *Store) Close() error {
	if err := store.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %v", err)
	}
	return nil
}

// SetActivity ...
func (store Store) SetActivity(chatID int64, t time.Time) error {
	err := store.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketActivities).Put(itob(chatID), itob(t.Unix()))
	})
	if err != nil {
		return fmt.Errorf("failed to set activity for chatID %d: %v: %v", chatID, t, err)
	}
	return nil
}

// SetRead ...
func (store Store) SetRead(chatID int64, t time.Time) error {
	err := store.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketReads).Put(itob(chatID), itob(t.Unix()))
	})
	if err != nil {
		return fmt.Errorf("failed to set read for chatID %d: %v: %v", chatID, t, err)
	}
	return nil
}

func itob(v int64) []byte {
	b := make([]byte, 8)
	binary.PutVarint(b, int64(v))
	return b
}

func btoi(b []byte) (int64, error) {
	return binary.ReadVarint(bytes.NewBuffer(b))
}
