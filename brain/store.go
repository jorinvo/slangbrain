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

const (
	// Time to wait for first study in hours
	baseStudytime = 2
	// Maximum number of new studies per day
	newPerDay = 20
	// Minimum number of studies needed to be due before notifying user
	dueMinCount = 3
	// Time user has to be inactive before being notified
	dueMinInactive = 10 * time.Minute
)

var (
	bucketModes         = []byte("modes")
	bucketPhrases       = []byte("phrases")
	bucketStudytimes    = []byte("studytimes")
	bucketReads         = []byte("reads")
	bucketActivities    = []byte("activities")
	bucketSubscriptions = []byte("subscriptions")
)
var buckets = [][]byte{
	bucketModes,
	bucketPhrases,
	bucketStudytimes,
	bucketReads,
	bucketActivities,
	bucketSubscriptions,
}

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
		prefix := itob(chatID)
		phraseID := append(prefix, itob(int64(sequence))...)
		// Phrase to JSON
		buf, err := json.Marshal(newPhrase(phrase, explanation))
		if err != nil {
			return err
		}
		// Save Phrase
		if err = bp.Put(phraseID, buf); err != nil {
			return err
		}
		// Limit number of new studies per day
		newPhrases := 0
		c := tx.Bucket(bucketPhrases).Cursor()
		var p Phrase
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			if err := json.Unmarshal(v, &p); err != nil {
				return err
			}
			if p.Score == 0 {
				newPhrases++
			}
		}
		// Save study time
		bs := tx.Bucket(bucketStudytimes)
		next := itob(time.Now().Add(time.Duration(newPhrases/newPerDay*24+baseStudytime) * time.Hour).Unix())
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
	err := store.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(bucketStudytimes).Cursor()
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
		bs := tx.Bucket(bucketStudytimes)
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
		next := itob(now.Add((baseStudytime << uint(newScore)) * time.Hour).Unix())
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
	err := store.db.View(func(tx *bolt.Tx) error {
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

// DeleteStudyPhrase ...
func (store Store) DeleteStudyPhrase(chatID int64) error {
	err := store.db.Update(func(tx *bolt.Tx) error {
		bs := tx.Bucket(bucketStudytimes)
		c := bs.Cursor()
		now := time.Now().Unix()
		prefix := itob(chatID)
		var keyTime int64
		var key []byte

		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			timestamp, err := btoi(v)
			if err != nil {
				return err
			}
			if timestamp > now {
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
		// Delete study time
		if err := bs.Delete(key); err != nil {
			return err
		}
		// Delete phrase
		return tx.Bucket(bucketPhrases).Delete(key)
	})

	if err != nil {
		return fmt.Errorf("failed to delete study phrase for chatID %d: %v", chatID, err)
	}
	return nil
}

// SetRead ...
func (store Store) SetRead(chatID int64, t time.Time) error {
	err := store.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketReads).Put(itob(chatID), itob(t.Unix()))
	})
	if err != nil {
		return fmt.Errorf("failed to set read for chatID %d: %v: %v", chatID, t, err)
	}
	return nil
}

// SetActivity ...
func (store Store) SetActivity(chatID int64, t time.Time) error {
	err := store.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketActivities).Put(itob(chatID), itob(t.Unix()))
	})
	if err != nil {
		return fmt.Errorf("failed to set activity for chatID %d: %v: %v", chatID, t, err)
	}
	return nil
}

// GetDueStudies ...
func (store Store) GetDueStudies() (map[int64]uint, error) {
	dueStudies := map[int64]uint{}
	now := time.Now().Unix()
	err := store.db.View(func(tx *bolt.Tx) error {
		err := tx.Bucket(bucketStudytimes).ForEach(func(k, v []byte) error {
			t, err := btoi(v)
			if err != nil || t > now {
				return err
			}
			chatID, err := btoi(k[:8])
			if err != nil {
				return err
			}
			isSubscribed, err := store.IsSubscribed(chatID)
			if err != nil || !isSubscribed {
				return err
			}
			dueStudies[chatID]++
			return nil
		})
		if err != nil {
			return err
		}
		// Check if user should be notified
		ba := tx.Bucket(bucketActivities)
		br := tx.Bucket(bucketReads)
		for chatID, count := range dueStudies {
			// Too little studies due
			if count < dueMinCount {
				fmt.Println("too little studies due")
				delete(dueStudies, chatID)
				continue
			}
			key := itob(chatID)
			// User was active just now
			activity, err := btoi(ba.Get(key))
			if err != nil {
				return err
			}
			if time.Duration(now-activity)*time.Second < dueMinInactive {
				fmt.Println("user was just active")
				delete(dueStudies, chatID)
				continue
			}
			// User hasn't read last message
			read, err := btoi(br.Get(key))
			if err != nil {
				return err
			}
			if read < activity {
				fmt.Println("user hasn't read last message")
				delete(dueStudies, chatID)
				continue
			}
		}
		return nil
	})

	if err != nil {
		return dueStudies, fmt.Errorf("failed to get due studies: %v", err)
	}
	return dueStudies, nil
}

// IsSubscribed ...
func (store Store) IsSubscribed(chatID int64) (bool, error) {
	var isSubscribed bool
	err := store.db.View(func(tx *bolt.Tx) error {
		isSubscribed = tx.Bucket(bucketSubscriptions).Get(itob(chatID)) != nil
		return nil
	})
	if err != nil {
		return isSubscribed, fmt.Errorf("failed to check subscription for chat %d: %v", chatID, err)
	}
	return isSubscribed, nil
}

// Subscribe ...
func (store Store) Subscribe(chatID int64) error {
	err := store.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketSubscriptions).Put(itob(chatID), []byte{'1'})
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe chatID %d: %v", chatID, err)
	}
	return nil
}

// Unsubscribe ...
func (store Store) Unsubscribe(chatID int64) error {
	err := store.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketSubscriptions).Delete(itob(chatID))
	})
	if err != nil {
		return fmt.Errorf("failed to unsubscribe chatID %d: %v", chatID, err)
	}
	return nil
}

// Close the underlying database connection.
func (store *Store) Close() error {
	if err := store.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %v", err)
	}
	return nil
}

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

func itob(v int64) []byte {
	b := make([]byte, 8)
	binary.PutVarint(b, int64(v))
	return b
}

func btoi(b []byte) (int64, error) {
	return binary.ReadVarint(bytes.NewBuffer(b))
}
