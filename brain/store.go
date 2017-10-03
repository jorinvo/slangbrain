package brain

import (
	"fmt"
	"math/rand"
	"time"

	bolt "github.com/coreos/bbolt"
	"github.com/jorinvo/slangbrain/brain/bucket"
)

// Store provides functions to interact with the underlying database.
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
		// Ensure buckets exist
		for _, b := range bucket.All {
			_, err = tx.CreateBucketIfNotExists(b)
			if err != nil {
				return fmt.Errorf("failed to create bucket '%s': %v", b, err)
			}
		}

		// Clear expired message IDs
		now := time.Now().Add(-messageIDmaxAge).Unix()
		bm := tx.Bucket(bucket.MessageIDs)
		bm.ForEach(func(k []byte, v []byte) error {
			if len(v) < 8 || btoi(v) < now {
				return bm.Delete(k)
			}
			return nil
		})

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

// TrackNotify sets the last time a notifications was sent to a user.
func (store Store) TrackNotify(id int64, t time.Time) error {
	key := itob(id)
	err := store.db.Update(func(tx *bolt.Tx) error {
		if err := tx.Bucket(bucket.Activities).Put(key, itob(t.Unix())); err != nil {
			return err
		}
		return addCountToBucket(tx.Bucket(bucket.Notifies), key, 1)
	})
	if err != nil {
		return fmt.Errorf("failed to set activity for id %d: %v: %v", id, t, err)
	}
	return nil
}

// SetRead sets the last time the user read a message.
func (store Store) SetRead(id int64, t time.Time) error {
	err := store.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucket.Reads).Put(itob(id), itob(t.Unix()))
	})
	if err != nil {
		return fmt.Errorf("failed to set read for id %d: %v: %v", id, t, err)
	}
	return nil
}

// Register saves the date a user first started using the chatbot.
// This is later on used for statistics.
func (store Store) Register(id int64) error {
	return store.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucket.RegisterDates).Put(itob(id), itob(time.Now().Unix()))
	})
}

// QueueMessage marks a messageID as being processed.
// This ensures each message is only handled once,
// even if the messaging platforms delivers them multiple times.
// Should only be called with each messageID once.
// Otherwise returns store.ErrExists.
func (store Store) QueueMessage(messageID string) error {
	err := store.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket.MessageIDs)
		key := []byte(messageID)
		if b.Get(key) != nil {
			return ErrExists
		}

		b.Put(key, itob(time.Now().Unix()))
		return nil
	})

	if err != nil {
		if err == ErrExists {
			return err
		}
		return fmt.Errorf("failed to add message ID: %s: %v", messageID, err)
	}
	return nil
}
