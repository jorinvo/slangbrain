package brain

import (
	"fmt"

	bolt "github.com/coreos/bbolt"
)

// GetMode fetches the mode for a chat.
func (store Store) GetMode(id int64) (Mode, error) {
	var mode Mode
	err := store.db.View(func(tx *bolt.Tx) error {
		if v := tx.Bucket(bucketModes).Get(itob(id)); v != nil {
			mode = Mode(btoi(v))
		} else {
			mode = ModeGetStarted
		}
		return nil
	})
	if err != nil {
		return mode, fmt.Errorf("failed to get mode for id %d: %v", id, err)
	}
	return mode, nil
}

// SetMode updates the mode for a chat.
func (store Store) SetMode(id int64, mode Mode) error {
	err := store.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketModes).Put(itob(id), itob(int64(mode)))
	})
	if err != nil {
		return fmt.Errorf("[id=%d,mode=%v] failed to set mode: %v", id, mode, err)
	}
	return nil
}
