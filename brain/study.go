package brain

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/boltdb/bolt"
)

// GetStudy returns the current study the user needs to do.
func (store Store) GetStudy(id int64) (Study, error) {
	var study Study
	err := store.db.View(func(tx *bolt.Tx) error {
		key, total, fromNow := findCurrentStudy(tx, itob(id), time.Now())

		// No studies found
		if total == 0 {
			if fromNow > 0 {
				study = Study{Next: fromNow}
			}
			return nil
		}

		// Get study from phrase
		p, err := getPhrase(tx, key)
		if err != nil {
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
		return study, fmt.Errorf("failed to study with id %d: %v", id, err)
	}
	return study, nil
}

// ScoreStudy sets the score of the current study and moves to the next study.
func (store Store) ScoreStudy(id int64, scoreUpdate int) error {
	err := store.db.Update(func(tx *bolt.Tx) error {
		now := time.Now()
		prefix := itob(id)
		key, _, _ := findCurrentStudy(tx, prefix, now)

		// No studies found
		if key == nil {
			return errors.New("no study found")
		}

		// Get phrase
		p, err := getPhrase(tx, key)
		if err != nil {
			return err
		}

		// Update score
		prevScore := p.Score
		p.Score += scoreUpdate
		if p.Score < 0 {
			p.Score = 0
		}

		// Update zeroscore
		if prevScore == 0 && p.Score != 0 {
			if err := updateZeroscore(tx, prefix, -1); err != nil {
				return err
			}
		} else if prevScore != 0 && p.Score == 0 {
			if err := updateZeroscore(tx, prefix, 1); err != nil {
				return err
			}
		}

		// Update scoretotal
		if err := addCountToBucket(tx.Bucket(bucketScoretotals), prefix, p.Score-prevScore); err != nil {
			return err
		}

		// Save phrase
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(p); err != nil {
			return err
		}
		if err := tx.Bucket(bucketPhrases).Put(key, buf.Bytes()); err != nil {
			return err
		}

		// Update study time
		i := p.Score
		if i >= len(studyIntervals) {
			i = len(studyIntervals) - 1
		}
		next := itob(now.Add(studyIntervals[i] + diffusion()).Unix())
		if err := tx.Bucket(bucketStudytimes).Put(key, next); err != nil {
			return err
		}

		fmt.Printf("phrase: %s; new score: %v; update: %v; next study: %v\n", p.Phrase, p.Score, scoreUpdate, time.Unix(btoi(next), 0).Sub(now))

		// Save study for reference and to analyze them later
		idAndTime := append(prefix, itob(now.Unix())...)
		seqAndScores := append(append(key[8:], itob(int64(scoreUpdate))...), itob(int64(p.Score))...)
		if err := tx.Bucket(bucketStudies).Put(idAndTime, seqAndScores); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to score study with id %d: %v", id, err)
	}
	return nil
}

// Randomize order by spreading studies over a period of time
func diffusion() time.Duration {
	return time.Duration(rand.Intn(studyTimeDiffusion)) * time.Minute
}

// GetNotifyTime gets the time until the user should be notified to study.
// Returns the time until the next studies are ready and a count of the ready studies.
// The returned duration is always at least dueMinInactive.
// The count is 0 if the chat has no phrases yet.
// The returned duration gets delayed if it would be in a user's night time.
// Nighttime is calculated form the passed timezone.
func (store Store) GetNotifyTime(id int64, timezone int) (time.Duration, int, error) {
	due := 0
	now := time.Now()

	// Delay if night
	var delay time.Duration
	userHour := now.Add(delay).UTC().Hour() + timezone
	if userHour > nightStart {
		delay = time.Duration(24-userHour+nightEnd)*time.Hour + time.Duration(60-now.Minute())*time.Minute
	} else if userHour < nightEnd {
		delay = time.Duration(nightEnd-userHour)*time.Hour - time.Duration(now.Minute())*time.Minute
	}
	// Ensure minimum delay
	if delay < dueMinInactive {
		delay = dueMinInactive
	}

	minTime := now.Add(delay).Unix()
	var nexts sortableInts

	err := store.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(bucketStudytimes).Cursor()
		prefix := itob(id)

		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			timestamp := btoi(v)
			if timestamp < minTime {
				due++
			}

			if due >= dueMinCount {
				continue
			}

			l := len(nexts)
			if l < dueMinCount {
				nexts = append(nexts, timestamp)
				sort.Sort(nexts)
				continue
			}

			if timestamp < nexts[l-1] {
				nexts = append(nexts[:l-1], timestamp)
				sort.Sort(nexts)
			}
		}

		return nil
	})

	if err != nil {
		return 0, 0, fmt.Errorf("failed to get next studies for chat %d: %v", id, err)
	}

	// If user has too little phrases, minCount is ignored
	minCount := dueMinCount
	l := len(nexts)
	if minCount > l {
		minCount = l
	}

	// Studies are ready already, notify ASAP
	if due >= minCount {
		return delay, due, nil
	}

	return time.Unix(nexts[l-1], 0).Sub(now), l, nil
}

// EachActiveChat runs a function for each chat
// where the user has been active since the last notification has been sent.
func (store Store) EachActiveChat(fn func(int64)) error {
	return store.db.View(func(tx *bolt.Tx) error {
		active := tx.Bucket(bucketActivities)
		return tx.Bucket(bucketReads).ForEach(func(k, v []byte) error {
			a := active.Get(k)
			if a == nil {
				a = itob(0)
			}
			timeActive := btoi(a)
			timeRead := btoi(v)
			if timeRead > timeActive {
				fn(btoi(k))
			}
			return nil
		})
	})
}

type sortableInts []int64

func (b sortableInts) Len() int {
	return len(b)
}

func (b sortableInts) Less(i, j int) bool {
	return b[i] < b[j]
}

func (b sortableInts) Swap(i, j int) {
	b[j], b[i] = b[i], b[j]
}
