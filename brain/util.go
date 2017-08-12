package brain

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"net/http"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
)

func itob(v int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func btoi(b []byte) int64 {
	return int64(binary.BigEndian.Uint64(b))
}

// BackupTo streams backup as an HTTP response.
func (store Store) BackupTo(w http.ResponseWriter) {
	err := store.db.View(func(tx *bolt.Tx) error {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", `attachment; filename="my.db"`)
		w.Header().Set("Content-Length", strconv.Itoa(int(tx.Size())))
		_, err := tx.WriteTo(w)
		return err
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// StudyNow resets all study times of all users to now.
func (store Store) StudyNow() error {
	return store.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketStudytimes)
		now := itob(time.Now().Unix())
		return b.ForEach(func(k, v []byte) error {
			return b.Put(k, now)
		})
	})
}

// DeleteChat removes all records of a given chat.
func (store Store) DeleteChat(id int64) error {
	return store.db.Update(func(tx *bolt.Tx) error {
		key := itob(id)
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

// GetChatIDs returns chatIDs of all users.
func (store Store) GetChatIDs() ([]int64, error) {
	var ids []int64
	err := store.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketModes).ForEach(func(k, v []byte) error {
			ids = append(ids, btoi(k))
			return nil
		})
	})
	return ids, err
}

// DeletePhrases removes all phrases fn matches.
func (store Store) DeletePhrases(fn func(int64, Phrase) bool) (int, error) {
	deleted := 0
	err := store.db.Update(func(tx *bolt.Tx) error {
		bp := tx.Bucket(bucketPhrases)
		bs := tx.Bucket(bucketStudytimes)
		return bp.ForEach(func(k, v []byte) error {
			id := btoi(k[:8])
			var p Phrase
			if err := gob.NewDecoder(bytes.NewReader(v)).Decode(&p); err != nil {
				return err
			}
			if !fn(id, p) {
				return nil
			}
			if err := bs.Delete(k); err != nil {
				return err
			}
			if err := bp.Delete(k); err != nil {
				return err
			}
			deleted++
			return nil
		})
	})
	return deleted, err
}
