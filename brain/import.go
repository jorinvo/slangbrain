package brain

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"

	bolt "github.com/coreos/bbolt"
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
			return fmt.Errorf("remove duplicates for %d: %v", id, err)
		}

		if len(phrases) == 0 {
			return nil
		}

		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(phrases); err != nil {
			return fmt.Errorf("gob of phrases %#v for %d: %v", phrases, id, err)
		}

		if err := tx.Bucket(bucketPendingImports).Put(prefix, buf.Bytes()); err != nil {
			return fmt.Errorf("put pending import %#v for %d: %v", phrases, id, err)
		}
		return nil
	})
	return len(phrases), err
}

// Import a list of phrases.
// Ingores phrases that already exist in DB.
// Ignores score of phrases.
// Sets the studytime of the phrases to now.
// Returns the number of actually imported phrases.
func (store Store) Import(id int64, phrases []Phrase) (int, error) {
	count := 0

	err := store.db.Update(func(tx *bolt.Tx) error {
		var err error
		count, err = phraseImporter(tx, itob(id), phrases)
		if err != nil {
			return fmt.Errorf("import phrases for %d: %v", id, err)
		}

		return nil
	})

	return count, err
}

func phraseImporter(tx *bolt.Tx, prefix []byte, phrases []Phrase) (int, error) {
	ps, err := removeDuplicates(tx, prefix, phrases)
	if err != nil {
		return 0, err
	}

	now := time.Now()

	for _, p := range ps {
		if err := phraseAdder(prefix, p, now, now)(tx); err != nil {
			return 0, err
		}
	}

	return len(ps), addCountToBucket(tx.Bucket(bucketImports), prefix, 1)
}

// Go through existing phrases, find duplicates and remove them from phrases
func removeDuplicates(tx *bolt.Tx, prefix []byte, phrases []Phrase) ([]Phrase, error) {
	c := tx.Bucket(bucketPhrases).Cursor()

	for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
		var e Phrase
		if err := gob.NewDecoder(bytes.NewBuffer(v)).Decode(&e); err != nil {
			return nil, fmt.Errorf("gob decode phrase at %#v: %v", k, err)
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

		var phrases []Phrase
		if b := bi.Get(prefix); b != nil {
			if err := gob.NewDecoder(bytes.NewReader(b)).Decode(&phrases); err != nil {
				return err
			}
		} else {
			return ErrNotFound
		}

		var err error
		count, err = phraseImporter(tx, prefix, phrases)
		if err != nil {
			return err
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
