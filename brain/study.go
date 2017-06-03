package brain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/boltdb/bolt"
)

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
		// Randomize order by spreading studies over a period of time
		diffusion := time.Duration(rand.Intn(studyTimeDiffusion)) * time.Minute
		next := itob(now.Add((baseStudytime<<uint(newScore))*time.Hour + diffusion).Unix())
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
