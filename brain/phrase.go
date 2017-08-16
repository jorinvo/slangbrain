package brain

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
)

// AddPhrase stores a new phrase.
// Pass time the phrase should be created at.
func (store Store) AddPhrase(id int64, phrase, explanation string, createdAt time.Time) error {
	err := store.db.Update(func(tx *bolt.Tx) error {
		bp := tx.Bucket(bucketPhrases)

		// Get phrase id
		sequence, err := bp.NextSequence()
		if err != nil {
			return err
		}
		prefix := itob(id)
		phraseID := append(prefix, itob(int64(sequence))...)

		// Phrase to GOB
		var buf bytes.Buffer
		tmp := Phrase{Phrase: phrase, Explanation: explanation}
		if err := gob.NewEncoder(&buf).Encode(tmp); err != nil {
			return err
		}

		// Save phrase
		if err := bp.Put(phraseID, buf.Bytes()); err != nil {
			return err
		}

		// Limit number of new studies per day
		newPhrases := 0
		c := tx.Bucket(bucketPhrases).Cursor()
		var p Phrase
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			if err := gob.NewDecoder(bytes.NewReader(v)).Decode(&p); err != nil {
				return err
			}
			if p.Score == 0 {
				newPhrases++
			}
		}

		// Save study time
		next := itob(createdAt.Add(time.Duration(newPhrases/newPerDay*24+firstStudytime) * time.Hour).Unix())
		if err := tx.Bucket(bucketStudytimes).Put(phraseID, next); err != nil {
			return err
		}

		// Save time phrase has been added
		return tx.Bucket(bucketPhraseAddTimes).Put(phraseID, itob(createdAt.Unix()))
	})

	if err != nil {
		return fmt.Errorf("failed to add phrase for id %d: %s - %s: %v", id, phrase, explanation, err)
	}
	return nil
}

// FindPhrase returns a phrase belonging to the passed user that matches the passed function.
func (store Store) FindPhrase(id int64, fn func(Phrase) bool) (Phrase, error) {
	var p Phrase
	err := store.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(bucketPhrases).Cursor()
		prefix := itob(id)
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			var tmp Phrase
			if err := gob.NewDecoder(bytes.NewReader(v)).Decode(&tmp); err != nil {
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
		return p, fmt.Errorf("failed to find phrase with id %d: %v", id, err)
	}
	return p, nil
}

// DeleteStudyPhrase deletes the phrase the passed user currently has to study.
func (store Store) DeleteStudyPhrase(id int64) error {
	err := store.db.Update(func(tx *bolt.Tx) error {
		bs := tx.Bucket(bucketStudytimes)
		bp := tx.Bucket(bucketPhrases)
		bd := tx.Bucket(bucketDeletedPhrases)
		c := bs.Cursor()
		now := time.Now().Unix()
		prefix := itob(id)
		var keyTime int64
		var key []byte

		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			timestamp := btoi(v)
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

		// Backup deleted phrases for now
		if err := bd.Put(key, bp.Get(key)); err != nil {
		}
		// Delete phrase
		return bp.Delete(key)
	})

	if err != nil {
		return fmt.Errorf("failed to delete study phrase for id %d: %v", id, err)
	}
	return nil
}

// GetAllPhrases returns all phrases for a given user in a map with phrase sequence numbers as keys.
func (store Store) GetAllPhrases(id int64) (map[int64]Phrase, error) {
	phrases := map[int64]Phrase{}
	err := store.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(bucketPhrases).Cursor()
		prefix := itob(id)
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			var p Phrase
			if err := gob.NewDecoder(bytes.NewBuffer(v)).Decode(&p); err != nil {
				return err
			}
			phrases[btoi(k[8:])] = p
		}
		return nil
	})
	if err != nil {
		return phrases, fmt.Errorf("failed to get all phrases for %d: %v", id, err)
	}
	return phrases, nil
}
