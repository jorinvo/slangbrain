package brain

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"sort"
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
		// Delete add time
		if err := tx.Bucket(bucketPhraseAddTimes).Delete(key); err != nil {
			return err
		}
		// Delete phrase
		return bp.Delete(key)
	})

	if err != nil {
		return fmt.Errorf("failed to delete study phrase for id %d: %v", id, err)
	}
	return nil
}

// IDPhrase is a phrase format that contains the ID but not the Score.
type IDPhrase struct {
	ID                  int64
	Phrase, Explanation string
}

type idPhrases struct {
	p []IDPhrase
	t map[int64]int64
}

func (p idPhrases) Len() int {
	return len(p.p)
}

func (p idPhrases) Less(i, j int) bool {
	return p.t[p.p[i].ID] < p.t[p.p[j].ID]
}

func (p idPhrases) Swap(i, j int) {
	p.p[j], p.p[i] = p.p[i], p.p[j]
}

// GetAllPhrases returns all phrases for a given user in a map with phrase sequence numbers as keys.
func (store Store) GetAllPhrases(id int64) ([]IDPhrase, error) {
	var phrases []IDPhrase
	addTimes := map[int64]int64{}
	err := store.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(bucketPhrases).Cursor()
		bt := tx.Bucket(bucketPhraseAddTimes)
		prefix := itob(id)
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			var p Phrase
			if err := gob.NewDecoder(bytes.NewBuffer(v)).Decode(&p); err != nil {
				return err
			}
			seq := btoi(k[8:])
			phrases = append(phrases, IDPhrase{seq, p.Phrase, p.Explanation})
			var t int64
			tb := bt.Get(k)
			if tb != nil {
				t = btoi(tb)
			}
			addTimes[seq] = t
		}
		return nil
	})
	sort.Sort(idPhrases{phrases, addTimes})
	if err != nil {
		return phrases, fmt.Errorf("failed to get all phrases for %d: %v", id, err)
	}
	return phrases, nil
}

// DeletePhrase removes a phrase.
func (store Store) DeletePhrase(id int64, seq int) error {
	key := append(itob(id), itob(int64(seq))...)
	err := store.db.Update(func(tx *bolt.Tx) error {
		// Delete study time
		if err := tx.Bucket(bucketStudytimes).Delete(key); err != nil {
			return err
		}
		// Delete add time
		if err := tx.Bucket(bucketPhraseAddTimes).Delete(key); err != nil {
			return err
		}
		// Delete phrase
		return tx.Bucket(bucketPhrases).Delete(key)
	})
	if err != nil {
		return fmt.Errorf("failed to delete phrase for key %x: %v", key, err)
	}
	return nil
}

// UpdatePhrase updates an existing phrase.
// Return ErrNotFound if phrase does not exist.
func (store Store) UpdatePhrase(id int64, seq int, phrase, explanation string) error {
	key := append(itob(id), itob(int64(seq))...)
	err := store.db.Update(func(tx *bolt.Tx) error {
		bp := tx.Bucket(bucketPhrases)
		// Get existing phrase
		b := bp.Get(key)
		if b == nil {
			return ErrNotFound
		}
		var p Phrase
		if err := gob.NewDecoder(bytes.NewBuffer(b)).Decode(&p); err != nil {
			return err
		}
		// Update
		p.Phrase = phrase
		p.Explanation = explanation
		// Save phrase
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(p); err != nil {
			return err
		}
		return bp.Put(key, buf.Bytes())
	})

	if err != nil {
		return fmt.Errorf("failed to update phrase for key %x: %s - %s: %v", key, phrase, explanation, err)
	}
	return nil
}
