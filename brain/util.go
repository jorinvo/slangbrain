package brain

import (
	"bytes"
	"time"

	"github.com/boltdb/bolt"
)

// BackupTo writes backups to a file.
func (store *Store) BackupTo(file string) error {
	return store.db.View(func(tx *bolt.Tx) error {
		return tx.CopyFile(file, 0600)
	})
}

// StudyNow is only for debugging.
// Resets all study times to now.
func (store *Store) StudyNow() error {
	return store.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketStudytimes)
		now := itob(time.Now().Unix())
		return b.ForEach(func(k, v []byte) error {
			return b.Put(k, now)
		})
	})
}

// DeleteChat is only for debugging.
// Removes all records of a given chat.
func (store *Store) DeleteChat(chatID int64) error {
	return store.db.Update(func(tx *bolt.Tx) error {
		key := itob(chatID)
		// Remove mode
		if err := tx.Bucket(bucketModes).Delete(key); err != nil {
			return err
		}
		// Remove phrases
		bp := tx.Bucket(bucketPhrases)
		c := bp.Cursor()
		for k, _ := c.Seek(key); k != nil && bytes.HasPrefix(k, key); k, _ = c.Next() {
			if err := bp.Delete(k); err != nil {
				return err
			}
		}
		// Remove study times
		bs := tx.Bucket(bucketStudytimes)
		c = bp.Cursor()
		for k, _ := c.Seek(key); k != nil && bytes.HasPrefix(k, key); k, _ = c.Next() {
			if err := bs.Delete(k); err != nil {
				return err
			}
		}
		return nil
	})
}
