package brain

import (
	"fmt"
	"time"

	bolt "github.com/coreos/bbolt"
	"github.com/jorinvo/slangbrain/brain/bucket"
)

// IsDuplicate checks whether a given payload has been sent twice in a row.
// It saves the previous payload and a timestamp in the database and compares them with each call.
func (store Store) IsDuplicate(id int64, payload string) (bool, error) {
	key := itob(id)
	now := time.Now()
	isDuplicate := false

	err := store.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket.PrevPayloads)

		// Check if previous payload was the same and if it was in so recent that it is a duplicate
		if v := b.Get(key); v != nil {
			if string(v[8:]) == payload && now.Sub(time.Unix(btoi(v[:8]), 0)) < payloadDuplicateInterval {
				isDuplicate = true
			}
		}

		// Track payload
		if err := b.Put(key, append(itob(now.Unix()), []byte(payload)...)); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return isDuplicate, fmt.Errorf("[id=%d,p='%s'] failed to check if payload is a duplicate: %v", id, payload, err)
	}
	return isDuplicate, nil
}
