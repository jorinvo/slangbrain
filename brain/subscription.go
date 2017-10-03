package brain

import (
	"fmt"

	bolt "github.com/coreos/bbolt"
	"github.com/jorinvo/slangbrain/brain/bucket"
)

// IsSubscribed checks if a user has notifications enabled.
func (store Store) IsSubscribed(id int64) (bool, error) {
	var isSubscribed bool
	err := store.db.View(func(tx *bolt.Tx) error {
		isSubscribed = tx.Bucket(bucket.Subscriptions).Get(itob(id)) != nil
		return nil
	})
	if err != nil {
		return isSubscribed, fmt.Errorf("failed to check subscription for chat %d: %v", id, err)
	}
	return isSubscribed, nil
}

// Subscribe enables notifications for a user.
func (store Store) Subscribe(id int64) error {
	err := store.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucket.Subscriptions).Put(itob(id), []byte{'1'})
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe chat %d: %v", id, err)
	}
	return nil
}

// Unsubscribe disables notifications for a user.
func (store Store) Unsubscribe(id int64) error {
	err := store.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucket.Subscriptions).Delete(itob(id))
	})
	if err != nil {
		return fmt.Errorf("failed to unsubscribe chat %d: %v", id, err)
	}
	return nil
}
