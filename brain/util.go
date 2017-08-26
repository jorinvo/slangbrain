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

// Limit number of studies per day.
// The more new phrases there are, the later the studies should be scheduled.
// Returns an offset to add to a timestamp.
func limitPerDay(tx *bolt.Tx, prefix []byte) (time.Duration, error) {
	var newPhrases float64
	c := tx.Bucket(bucketPhrases).Cursor()
	var tmp Phrase
	for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
		if err := gob.NewDecoder(bytes.NewReader(v)).Decode(&tmp); err != nil {
			return 0, err
		}
		if tmp.Score == 0 {
			newPhrases++
		}
	}

	return time.Duration(newPhrases/newPerDay*24) * time.Hour, nil
}
