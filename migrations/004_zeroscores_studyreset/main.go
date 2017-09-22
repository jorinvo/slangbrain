// To simplify the calulation of the offset used when calculating studytimes
// we introduced a new bucket zeroscores.
//
// The bucket is updated at every place where we change the score of a phrase,
// but with the migration we update the bucket for all existing phrases.
//
// Additionally, all studytimes are reset to times calculated with the new algorithm.
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	bolt "github.com/coreos/bbolt"
)

var (
	bucketPhrases      = []byte("phrases")
	bucketStudytimes   = []byte("studytimes")
	bucketZeroscores   = []byte("zeroscores")
	studyTimeDiffusion = 30
	newPerDay          = 30.0
)

var studyIntervals = [14]time.Duration{
	2 * time.Hour,
	8 * time.Hour,
	20 * time.Hour,
	44 * time.Hour,
	(4*24 - 2) * time.Hour,
	(7*24 - 2) * time.Hour,
	(14*24 - 2) * time.Hour,
	(30*24 - 2) * time.Hour,
	(60*24 - 2) * time.Hour,
	(100*24 - 2) * time.Hour,
	(5*30*24 - 2) * time.Hour,
	(8*30*24 - 2) * time.Hour,
	(12*30*24 - 2) * time.Hour,
	(15*30*24 - 2) * time.Hour,
}

type phrase struct {
	Phrase      string
	Explanation string
	Score       int
}

func main() {
	dbFile := os.Args[1]
	db, err := bolt.Open(dbFile, 0600, &bolt.Options{Timeout: 1 * time.Second})
	fatal(err)
	defer func() {
		fatal(db.Close())
	}()

	now := time.Now()

	fatal(db.Update(func(tx *bolt.Tx) error {
		bz, err := tx.CreateBucketIfNotExists(bucketZeroscores)
		if err != nil {
			return err
		}

		err = tx.Bucket(bucketPhrases).ForEach(func(k []byte, v []byte) error {
			prefix := k[:8]

			var p phrase
			if err := gob.NewDecoder(bytes.NewReader(v)).Decode(&p); err != nil {
				return err
			}

			// Update study time
			i := p.Score
			if i < 0 {
				i = 0
			}
			if i >= len(studyIntervals) {
				i = len(studyIntervals) - 1
			}
			o := limitPerDay(tx, prefix)
			d := diffusion()
			next := studyIntervals[i] + o + d
			if err := tx.Bucket(bucketStudytimes).Put(k, itob(now.Add(next).Unix())); err != nil {
				return err
			}

			fmt.Printf("id:%v;	s: %v;	i: %v;	offset: %v;	diffusion: %v;	next: %v\n", btoi(k[8:]), p.Score, i, o, d, next)

			// Update zeroscore
			if p.Score != 0 {
				return nil
			}
			var zs int64
			if v := bz.Get(prefix); v != nil {
				zs = btoi(v)
			}
			return bz.Put(prefix, itob(zs+1))
		})

		if err != nil {
			return err
		}

		return bz.ForEach(func(k []byte, v []byte) error {
			fmt.Printf("\nid: %d; score: %d\n", btoi(k), btoi(v))
			return nil
		})
	}))
}

func fatal(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func itob(v int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func btoi(b []byte) int64 {
	return int64(binary.BigEndian.Uint64(b))
}

func limitPerDay(tx *bolt.Tx, key []byte) time.Duration {
	var zeroScores float64
	if v := tx.Bucket(bucketZeroscores).Get(key); v != nil {
		zeroScores = float64(btoi(v))
	}
	return time.Duration(zeroScores/newPerDay*24) * time.Hour
}

func diffusion() time.Duration {
	return time.Duration(rand.Intn(studyTimeDiffusion)) * time.Minute
}
