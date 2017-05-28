package brain

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
)

// Time to wait for first study in hours
const startStudytime = 2

var (
	bucketModes   = []byte("modes")
	bucketPhrases = []byte("phrases")
	// bmid = []byte("messengerids")
	bucketStudytime = []byte("studytimes")
	// bsl     = []byte("studylogs")
	buckets = [][]byte{bucketModes, bucketPhrases, bucketStudytime}
)

// Store ...
type Store struct {
	db *bolt.DB
}

// CreateStore returns a new Store with a database already setup.
func CreateStore(dbFile string) (Store, error) {
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
	return store, err
}

// GetMode fetches the mode for a chat.
func (store Store) GetMode(chatID int64) (Mode, error) {
	var mode Mode
	err := store.db.View(func(tx *bolt.Tx) error {
		if bm := tx.Bucket(bucketModes).Get(itob(chatID)); bm != nil {
			iMode, err := btoi(bm)
			if err != nil {
				return err
			}
			mode = Mode(iMode)
		} else {
			mode = ModeGetStarted
		}
		return nil
	})
	if err != nil {
		return mode, fmt.Errorf("failed to get mode for chatID %d: %v", chatID, err)
	}
	return mode, nil
}

// SetMode updates the mode for a chat.
func (store Store) SetMode(chatID int64, mode Mode) error {
	err := store.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketModes).Put(itob(chatID), itob(int64(mode)))
	})
	if err != nil {
		return fmt.Errorf("failed to set mode for chatID %d: %d: %v", chatID, mode, err)
	}
	return nil
}

// AddPhrase stores a new phrase.
func (store Store) AddPhrase(chatID int64, phrase, explanation string) error {
	err := store.db.Update(func(tx *bolt.Tx) error {
		bp := tx.Bucket(bucketPhrases)
		// Get phrase id
		sequence, err := bp.NextSequence()
		if err != nil {
			return err
		}
		phraseID := append(itob(chatID), itob(int64(sequence))...)
		// Phrase to JSON
		buf, err := json.Marshal(newPhrase(phrase, explanation))
		if err != nil {
			return err
		}
		// Save Phrase
		if err = bp.Put(phraseID, buf); err != nil {
			return err
		}
		// Save study time
		bs := tx.Bucket(bucketStudytime)
		next := itob(time.Now().Add(startStudytime * time.Hour).Unix())
		return bs.Put(phraseID, next)
	})

	if err != nil {
		return fmt.Errorf("failed to add phrase for chatID %d: %s - %s: %v", chatID, phrase, explanation, err)
	}
	return nil
}

// GetStudy ...
func (store Store) GetStudy(chatID int64) (Study, error) {
	var study Study
	err := store.db.Update(func(tx *bolt.Tx) error {
		c := tx.Bucket(bucketStudytime).Cursor()
		now := time.Now().Unix()
		prefix := itob(chatID)
		total := 0
		var keyTime int64
		var key []byte

		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			timestamp, err := btoi(v)
			if err != nil {
				return err
			}
			if timestamp < keyTime || keyTime == 0 {
				keyTime = timestamp
				key = k
			}
			if timestamp <= now {
				total++
			}
		}
		// No studies found
		if total == 0 {
			var next time.Duration
			if keyTime > 0 {
				next = time.Second * time.Duration(keyTime-now)
			}
			study = Study{Next: next}
			return nil
		}
		// Get study from phrase
		var p Phrase
		if err := json.Unmarshal(tx.Bucket(bucketPhrases).Get(key), &p); err != nil {
			return err
		}
		study = Study{
			Phrase:      p.Phrase,
			Explanation: p.Explanation,
			Total:       total,
		}
		return nil
	})

	if err != nil {
		return study, fmt.Errorf("failed to study with chatID %d: %v", chatID, err)
	}
	return study, nil
}

// ScoreStudy ...
func (store Store) ScoreStudy(chatID int64, score Score) error {
	err := store.db.Update(func(tx *bolt.Tx) error {
		bs := tx.Bucket(bucketStudytime)
		c := bs.Cursor()
		now := time.Now()
		uNow := now.Unix()
		prefix := itob(chatID)
		var keyTime int64
		var key []byte

		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			timestamp, err := btoi(v)
			if err != nil {
				return err
			}
			if timestamp > uNow {
				continue
			}
			if timestamp < keyTime || keyTime == 0 {
				keyTime = timestamp
				key = k
			}
		}
		// No studies found
		if key == nil {
			return errors.New("no study found")
		}
		// Get phrase
		var p Phrase
		bp := tx.Bucket(bucketPhrases)
		if err := json.Unmarshal(bp.Get(key), &p); err != nil {
			return err
		}
		// Update score
		p.Score += score
		newScore := p.Score
		if newScore < 0 {
			newScore = 0
		}
		// Save phrase
		buf, err := json.Marshal(p)
		if err != nil {
			return err
		}
		if err = bp.Put(key, buf); err != nil {
			return err
		}
		// Update study time
		next := itob(now.Add((2 << uint(newScore)) * time.Hour).Unix())
		if err = bs.Put(key, next); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to study with chatID %d: %v", chatID, err)
	}
	return nil
}

// FindPhrase ...
func (store Store) FindPhrase(chatID int64, fn func(Phrase) bool) (Phrase, error) {
	var p Phrase
	err := store.db.Update(func(tx *bolt.Tx) error {
		c := tx.Bucket(bucketPhrases).Cursor()
		prefix := itob(chatID)
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			var tmp Phrase
			if err := json.Unmarshal(v, &tmp); err != nil {
				return err
			}
			if fn(tmp) {
				p = tmp
				return nil
			}
		}
		return nil
	})

	if err != nil {
		return p, fmt.Errorf("failed to find phrase with chatid %d: %v", chatID, err)
	}
	return p, nil
}

// Close the underlying database connection.
func (store *Store) Close() error {
	if err := store.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %v", err)
	}
	return nil
}

// StudyNow is only for debugging. Resets all study times to now.
func (store *Store) StudyNow() error {
	return store.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketStudytime)
		now := itob(time.Now().Unix())
		return b.ForEach(func(k, v []byte) error {
			return b.Put(k, now)
		})
	})
}

func itob(v int64) []byte {
	b := make([]byte, 8)
	binary.PutVarint(b, int64(v))
	return b
}

func btoi(b []byte) (int64, error) {
	return binary.ReadVarint(bytes.NewBuffer(b))
}
