package brain

import (
	"bytes"
	"fmt"
	"time"

	bolt "github.com/coreos/bbolt"
	"github.com/jorinvo/slangbrain/brain/bucket"
)

// UserStats returns the Stats object for a user.
// If UserStats  hasn't been called for a time of at least statInterval.
// Otherwise returns ErrNotReady.
func (store Store) UserStats(id int64) (Stats, error) {
	var stats Stats
	err := store.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket.Stattimes)
		prefix := itob(id)
		now := time.Now()

		v := b.Get(prefix)
		if v == nil {
			if err := b.Put(prefix, itob(now.Unix())); err != nil {
				return err
			}
			return ErrNotReady
		}

		stattime := time.Unix(btoi(v), 0)

		if now.Sub(stattime) < statInterval {
			return ErrNotReady
		}

		score, rank, err := scoreAndRank(tx, prefix)
		if err != nil {
			return err
		}

		stats = Stats{
			Added:   countAdds(tx, prefix, now),
			Studied: countStudies(tx, prefix, now),
			Score:   score,
			Rank:    rank,
		}

		return b.Put(prefix, itob(now.Unix()))
	})

	if err == ErrNotReady {
		return stats, err
	}
	if err != nil {
		return stats, fmt.Errorf("failed to get stats for %d: %v", id, err)
	}
	return stats, nil
}

func countAdds(tx *bolt.Tx, prefix []byte, now time.Time) int {
	count := 0
	limit := now.Add(-statInterval).Unix()
	c := tx.Bucket(bucket.PhraseAddTimes).Cursor()

	for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
		if btoi(v) > limit {
			count++
		}
	}

	return count
}

func countStudies(tx *bolt.Tx, prefix []byte, now time.Time) int {
	count := 0
	limit := now.Add(-statInterval).Unix()
	c := tx.Bucket(bucket.Studies).Cursor()

	for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
		if btoi(k[8:]) > limit {
			count++
		}
	}

	return count
}

func scoreAndRank(tx *bolt.Tx, prefix []byte) (int, int, error) {
	b := tx.Bucket(bucket.Scoretotals)

	score := 0
	if v := b.Get(prefix); v != nil {
		score = int(btoi(v))
	}

	rank := 1
	err := b.ForEach(func(k, v []byte) error {
		if !bytes.Equal(k, prefix) && int(btoi(v)) > score {
			rank++
		}
		return nil
	})

	return score, rank, err
}
