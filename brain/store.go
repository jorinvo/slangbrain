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

var (
	bm = []byte("modes")
	bp = []byte("phrases")
	// bmid = []byte("messengerids")
	bst = []byte("studytimes")
	// bsl     = []byte("studylogs")
	buckets = [][]byte{bm, bp, bst}
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
		if m := tx.Bucket(bm).Get(itob(chatID)); m != nil {
			i, err := btoi(m)
			if err != nil {
				return err
			}
			mode = Mode(i)
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
		return tx.Bucket(bm).Put(itob(chatID), itob(int64(mode)))
	})
	if err != nil {
		return fmt.Errorf("failed to set mode for chatID %d: %d: %v", chatID, mode, err)
	}
	return nil
}

// AddPhrase stores a new phrase.
func (store Store) AddPhrase(chatID int64, phrase, explanation string) error {
	err := store.db.Update(func(tx *bolt.Tx) error {
		bPhrases := tx.Bucket(bp)
		// Get phrase id
		id, err := bPhrases.NextSequence()
		if err != nil {
			return err
		}
		bPhraseID := append(itob(chatID), itob(int64(id))...)
		// Phrase to JSON
		buf, err := json.Marshal(newPhrase(phrase, explanation))
		if err != nil {
			return err
		}
		// Save Phrase
		if err = bPhrases.Put(bPhraseID, buf); err != nil {
			return err
		}
		// Save study times for all study modes
		bStudytimes := tx.Bucket(bst)
		next := itob(time.Now().Add(4 * time.Hour).Unix())
		for _, sm := range Studymodes {
			key := append(bPhraseID, itob(int64(sm))...)
			if err = bStudytimes.Put(key, next); err != nil {
				return err
			}
		}

		return nil
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
		c := tx.Bucket(bst).Cursor()
		now := time.Now().Unix()
		prefix := itob(chatID)
		total := 0
		var oldest int64
		var keyOldest []byte

		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			timestamp, err := btoi(v)
			if err != nil {
				return err
			}
			if timestamp > now {
				continue
			}
			total++
			if timestamp < oldest || oldest == 0 {
				oldest = timestamp
				keyOldest = k
			}
		}
		// No studies found
		if keyOldest == nil {
			return nil
		}
		// Get study from phrase
		var p Phrase
		if err := json.Unmarshal(tx.Bucket(bp).Get(keyOldest[:16]), &p); err != nil {
			return err
		}
		m, err := btoi(keyOldest[16:])
		if err != nil {
			return err
		}
		study = Study{
			Phrase:      p.Phrase,
			Explanation: p.Explanation,
			Mode:        Studymode(m),
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
		bStudytimes := tx.Bucket(bst)
		c := bStudytimes.Cursor()
		now := time.Now()
		uNow := now.Unix()
		prefix := itob(chatID)
		var studyTime int64
		var studyKey []byte

		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			timestamp, err := btoi(v)
			if err != nil {
				return err
			}
			if timestamp > uNow {
				continue
			}
			if timestamp < studyTime || studyTime == 0 {
				studyTime = timestamp
				studyKey = k
			}
		}
		// No studies found
		if studyKey == nil {
			return errors.New("no study found")
		}
		// Get phrase
		var p Phrase
		bPhrases := tx.Bucket(bp)
		key := studyKey[:16]
		if err := json.Unmarshal(bPhrases.Get(key), &p); err != nil {
			return err
		}
		// Update score
		var newScore Score
		if bytes.Equal(studyKey[8:16], itob(int64(GuessPhrase))) {
			p.ScorePhrase += score
			newScore = p.ScorePhrase
		} else {
			p.ScoreExplanation += score
			newScore = p.ScoreExplanation
		}
		if newScore < 0 {
			newScore = 0
		}
		// Save phrase
		buf, err := json.Marshal(p)
		if err != nil {
			return err
		}
		if err = bPhrases.Put(key, buf); err != nil {
			return err
		}
		// Update study time
		next := itob(now.Add((2 << uint(newScore)) * time.Hour).Unix())
		if err = bStudytimes.Put(studyKey, next); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to study with chatID %d: %v", chatID, err)
	}
	return nil
}

// // CountStudies ...
// func (store Store) CountStudies(chatID int64) (int, error) {
// 	if err != nil {
// 		return count, fmt.Errorf("failed to count studies for chatID %d: %v", chatID, err)
// 	}
// 	return count, nil
// }

// FindPhrase ...
func (store Store) FindPhrase(chatID int64, fn func(Phrase) bool) (Phrase, error) {
	var p Phrase
	err := store.db.Update(func(tx *bolt.Tx) error {
		c := tx.Bucket(bp).Cursor()
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
		return p, fmt.Errorf("failed to find phrase with chatID %d: %v", chatID, err)
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

func itob(v int64) []byte {
	b := make([]byte, 8)
	binary.PutVarint(b, int64(v))
	return b
}

func btoi(b []byte) (int64, error) {
	return binary.ReadVarint(bytes.NewBuffer(b))
}
