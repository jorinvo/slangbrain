package brain

import (
	"bytes"
	"encoding/binary"
	"net/http"
	"strconv"
	"time"

	bolt "github.com/coreos/bbolt"
	"github.com/jorinvo/slangbrain/brain/bucket"
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

// Find the key for the phrase that should be studied next.
// Key is nil if non could be found.
// Also returns the total number of due studies and a duration until the next phrase is due.
// The duration is only useful if total is 0 otherwise the duration is a negative time.
func findCurrentStudy(tx *bolt.Tx, prefix []byte, now time.Time) ([]byte, int, time.Duration) {
	c := tx.Bucket(bucket.Studytimes).Cursor()
	uNow := now.Unix()
	total := 0
	var keyTime int64
	var key []byte

	for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
		timestamp := btoi(v)
		if timestamp < keyTime || keyTime == 0 {
			keyTime = timestamp
			key = k
		}
		if timestamp <= uNow {
			total++
		}
	}

	return key, total, time.Unix(keyTime, 0).Sub(now)
}

// Add a count to a bucket value.
// Limits to >= 0.
func addCountToBucket(b *bolt.Bucket, key []byte, count int) error {
	if v := b.Get(key); v != nil {
		count += int(btoi(v))
	}
	if count < 0 {
		count = 0
	}
	return b.Put(key, itob(int64(count)))
}
