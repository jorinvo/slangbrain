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
	p := Phrase{Phrase: phrase, Explanation: explanation}
	if err := store.db.Update(phraseAdder(itob(id), p, createdAt)); err != nil {
		return fmt.Errorf("failed to add phrase for id %d: %s - %s: %v", id, phrase, explanation, err)
	}
	return nil
}

// Abstract adding to reuse it for import.
func phraseAdder(prefix []byte, p Phrase, createdAt time.Time) func(*bolt.Tx) error {
	return func(tx *bolt.Tx) error {
		bp := tx.Bucket(bucketPhrases)
		bz := tx.Bucket(bucketZeroscores)

		// Get phrase id
		sequence, err := bp.NextSequence()
		if err != nil {
			return err
		}
		phraseID := itob(int64(sequence))
		key := append(prefix, phraseID...)

		// Phrase to GOB
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(p); err != nil {
			return err
		}

		// Save phrase
		if err := bp.Put(key, buf.Bytes()); err != nil {
			return err
		}

		// Queue as new phrase
		bn := tx.Bucket(bucketNewPhrases)
		if err := bn.Put(prefix, append(bn.Get(prefix), phraseID...)); err != nil {
			return err
		}

		// Try to schedule it
		var zeroscore int64
		if v := bz.Get(prefix); v != nil {
			zeroscore = btoi(v)
		}
		scheduled, err := scheduleNewPhrases(tx, prefix, time.Now(), int(zeroscore))
		if err != nil {
			return err
		}
		if err := bz.Put(prefix, itob(zeroscore+int64(scheduled))); err != nil {
			return err
		}

		// Save time phrase has been added
		return tx.Bucket(bucketPhraseAddTimes).Put(key, itob(createdAt.Unix()))
	}
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

// DeletePhrase removes a phrase.
func (store Store) DeletePhrase(id int64, seq int) error {
	key := append(itob(id), itob(int64(seq))...)
	err := store.db.Update(func(tx *bolt.Tx) error {
		return phraseDeleter(tx, key)
	})
	if err != nil {
		return fmt.Errorf("failed to delete phrase for key %x: %v", key, err)
	}
	return nil
}

// Reuse deleting functionality to only have one place
// to think about that all related buckets have been cleared.
func phraseDeleter(tx *bolt.Tx, key []byte) error {
	// Delete study time
	if err := tx.Bucket(bucketStudytimes).Delete(key); err != nil {
		return err
	}

	// Delete add time
	if err := tx.Bucket(bucketPhraseAddTimes).Delete(key); err != nil {
		return err
	}

	// Update zeroscore
	p, err := getPhrase(tx, key)
	if err != nil {
		return err
	}
	if p.Score == 0 {
		if err := updateZeroscore(tx, key[:8], -1); err != nil {
			return err
		}
	}

	// Delete phrase
	return tx.Bucket(bucketPhrases).Delete(key)
}

func getPhrase(tx *bolt.Tx, key []byte) (Phrase, error) {
	v := tx.Bucket(bucketPhrases).Get(key)
	var p Phrase
	if v == nil {
		return p, ErrNotFound
	}
	return p, gob.NewDecoder(bytes.NewReader(v)).Decode(&p)
}

// Adds a scoreUpdate to the zeroscore of a user.
// zeroscore cannot be less than zero.
// With each update we also check if we can schedule new phrases.
func updateZeroscore(tx *bolt.Tx, prefix []byte, scoreUpdate int) error {
	zeroscore := int64(scoreUpdate)
	bz := tx.Bucket(bucketZeroscores)
	if v := bz.Get(prefix); v != nil {
		zeroscore += btoi(v)
	}
	if zeroscore < 0 {
		zeroscore = 0
	}

	scheduled, err := scheduleNewPhrases(tx, prefix, time.Now(), int(zeroscore))
	if err != nil {
		return err
	}

	return bz.Put(prefix, itob(zeroscore+int64(scheduled)))
}

// Schedule new phrases for studying.
// Pass the number of new phrases already scheduled.
// Returns the number of phrases that have been additionally scheduled.
func scheduleNewPhrases(tx *bolt.Tx, prefix []byte, now time.Time, scheduled int) (int, error) {
	toSchedule := maxNewStudies - scheduled
	if toSchedule < 1 {
		return 0, nil
	}

	bn := tx.Bucket(bucketNewPhrases)
	bs := tx.Bucket(bucketStudytimes)
	v := bn.Get(prefix)
	var i, o int

	// Schedule as many as allowed to schedule and available from new phrases
	for i < toSchedule {
		o = i * 8
		if o >= len(v) {
			break
		}
		i++

		// Save study time
		phraseID := append(prefix, v[o:o+8]...)
		next := itob(now.Add(studyIntervals[0]).Unix())
		if err := bs.Put(phraseID, next); err != nil {
			return 0, err
		}
	}

	// Remove scheduled phrases from new phrases bucket
	if err := bn.Put(prefix, v[o:]); err != nil {
		return 0, err
	}

	return i, nil
}

// DeleteStudyPhrase deletes the phrase the user currently has to study.
func (store Store) DeleteStudyPhrase(id int64) error {
	err := store.db.Update(func(tx *bolt.Tx) error {
		key, _, _ := findCurrentStudy(tx, itob(id), time.Now())

		// No studies found
		if key == nil {
			return errors.New("no study found")
		}

		return phraseDeleter(tx, key)
	})

	if err != nil {
		return fmt.Errorf("failed to delete study phrase for id %d: %v", id, err)
	}
	return nil
}

// IDPhrase is a phrase format that also contains ID.
type IDPhrase struct {
	ID          int64
	Phrase      string
	Explanation string
	Score       int
}

type idPhrases struct {
	p []IDPhrase
	t map[int64]int64
}

func (p idPhrases) Len() int {
	return len(p.p)
}

func (p idPhrases) Less(i, j int) bool {
	return p.t[p.p[i].ID] > p.t[p.p[j].ID]
}

func (p idPhrases) Swap(i, j int) {
	p.p[j], p.p[i] = p.p[i], p.p[j]
}

// GetAllPhrases returns all phrases for a given user sorted by the time they have been added.
// Need to load all of the user's phrases in memory to be able to sort them.
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
			phrases = append(phrases, IDPhrase{seq, p.Phrase, p.Explanation, p.Score})
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

// UpdatePhrase updates an existing phrase.
// Return ErrNotFound if phrase does not exist.
func (store Store) UpdatePhrase(id int64, seq int, phrase, explanation string) error {
	key := append(itob(id), itob(int64(seq))...)
	err := store.db.Update(func(tx *bolt.Tx) error {
		bp := tx.Bucket(bucketPhrases)
		// Get existing phrase
		p, err := getPhrase(tx, key)
		if err != nil {
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
