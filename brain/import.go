package brain

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
)

// QueueImport stores phrases to be imported later.
// It filters out existing phrases.
// It returns the count of phrases stored for import.
// If no new phrases are imported, nothing written to DB.
func (store Store) QueueImport(id int64, phrases []Phrase) (int, error) {
	prefix := itob(id)

	err := store.db.Update(func(tx *bolt.Tx) error {
		var err error
		phrases, err = removeDuplicates(tx, prefix, phrases)
		if err != nil {
			return err
		}

		if len(phrases) == 0 {
			return nil
		}

		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(phrases); err != nil {
			return err
		}

		return tx.Bucket(bucketPendingImports).Put(prefix, buf.Bytes())
	})

	if err != nil {
		return len(phrases), fmt.Errorf("failed to queue import for %d: %v", id, err)
	}
	return len(phrases), nil
}

// Add a list of phrases.
// Ingores phrases that already exist in DB.
// Sets the studytime of the phrases to now.
func (store Store) Import(id int64, phrases []Phrase) (int, error) {
	prefix := itob(id)

	err := store.db.Update(func(tx *bolt.Tx) error {
		var err error
		phrases, err = removeDuplicates(tx, prefix, phrases)
		if err != nil {
			return err
		}

		for _, p := range phrases {
			if err := phraseAdder(prefix, p, time.Now().Add(-studyIntervals[0]))(tx); err != nil {
				return nil
			}
		}

		return nil
	})

	if err != nil {
		return len(phrases), fmt.Errorf("[id=%d] failed to import phrases: %v", id, err)
	}
	return len(phrases), nil
}

// Go through existing phrases, find duplicates and remove them from phrases
func removeDuplicates(tx *bolt.Tx, prefix []byte, phrases []Phrase) ([]Phrase, error) {
	c := tx.Bucket(bucketPhrases).Cursor()
	for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
		var e Phrase
		if err := gob.NewDecoder(bytes.NewBuffer(v)).Decode(&e); err != nil {
			return nil, err
		}
		for i, p := range phrases {
			if e.Explanation == p.Explanation {
				phrases = append(phrases[:i], phrases[i+1:]...)
				break
			}
		}
	}
	return phrases, nil
}

// ApplyImport adds phrases that have been previously queued with QueueImport().
// It returns the count of added phrases and clears the pending imports.
func (store Store) ApplyImport(id int64) (int, error) {
	prefix := itob(id)
	var count int

	err := store.db.Update(func(tx *bolt.Tx) error {
		bi := tx.Bucket(bucketPendingImports)

		b := bi.Get(prefix)
		if b == nil {
			return ErrNotFound
		}
		var phrases []Phrase
		if err := gob.NewDecoder(bytes.NewBuffer(b)).Decode(&phrases); err != nil {
			return err
		}

		count = len(phrases)

		for _, p := range phrases {
			if err := phraseAdder(prefix, p, time.Now())(tx); err != nil {
				return nil
			}
		}

		return bi.Delete(prefix)
	})

	if err != nil {
		return count, fmt.Errorf("failed to apply import for %d: %v", id, err)
	}
	return count, nil
}

// ClearImport removes a queued import from the pending imports bucket.
func (store Store) ClearImport(id int64) error {
	err := store.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketPendingImports).Delete(itob(id))
	})
	if err != nil {
		return fmt.Errorf("failed to clear import for %d: %v", id, err)
	}
	return nil
}
