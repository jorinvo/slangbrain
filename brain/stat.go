package brain

import (
	"bytes"
	"fmt"
	"io"
	"time"

	bolt "github.com/coreos/bbolt"
)

const statmsg = "```" + `
users:        %4d
subscriptions:%4d
dbsize:       %5.2fmb

format:         total (avg)
---------------------------
phrases:     %8d (%d)
scoretotal:  %8d (%d)
studies:     %8d (%d)
due studies: %8d (%d)
imports:     %8d (%d)
notifies:    %8d (%d)
zeroscore:   %8d (%d)
new phrases: %8d (%d)
%s` + "```"

// WriteStat writes plain text statistics for the whole DB to the given Writer.
// The formatting is intended for markdown usage such as in Slack.
func (store Store) WriteStat(w io.Writer) error {
	return store.db.View(func(tx *bolt.Tx) error {
		users := tx.Bucket(bucketRegisterDates).Stats().KeyN
		subscriptions := tx.Bucket(bucketSubscriptions).Stats().KeyN
		dbSize := float64(tx.Size()) / 1024.0 / 1024.0 // in mb

		phrasesTotal := tx.Bucket(bucketPhrases).Stats().KeyN
		phrasesAvg := phrasesTotal / users

		scoretotal, err := sum(tx.Bucket(bucketScoretotals), simplesum)
		if err != nil {
			return err
		}
		scoretotalAvg := scoretotal / users

		studiesTotal := tx.Bucket(bucketStudies).Stats().KeyN
		studiesAvg := studiesTotal / users

		now := itob(time.Now().Unix())
		dueStudiesTotal, err := sum(tx.Bucket(bucketStudytimes), func(v []byte) int {
			if bytes.Compare(v, now) < 1 {
				return 1
			}
			return 0
		})
		if err != nil {
			return err
		}
		dueStudiesAvg := dueStudiesTotal / users

		importsTotal, err := sum(tx.Bucket(bucketImports), simplesum)
		if err != nil {
			return err
		}
		importsAvg := importsTotal / users

		notifiesTotal, err := sum(tx.Bucket(bucketNotifies), simplesum)
		if err != nil {
			return err
		}
		notifiesAvg := notifiesTotal / users

		zeroscore, err := sum(tx.Bucket(bucketZeroscores), simplesum)
		if err != nil {
			return err
		}
		zeroscoreAvg := zeroscore / users

		newphrasesTotal, err := sum(tx.Bucket(bucketNewPhrases), count64)
		if err != nil {
			return err
		}
		newphrasesAvg := newphrasesTotal / users

		warnings := ""
		notNewPhrases := phrasesTotal - newphrasesTotal
		if n := tx.Bucket(bucketStudytimes).Stats().KeyN; n != notNewPhrases {
			warnings += fmt.Sprintf("\nWARNING: Number of studytimes (%d) does not match phrases - newphrases (%d).\n", n, notNewPhrases)
		}
		if n := tx.Bucket(bucketPhraseAddTimes).Stats().KeyN; n != phrasesTotal {
			warnings += fmt.Sprintf("\nWARNING: Number of phraseaddtimes (%d) does not match number of phrases (%d).\n", n, phrasesTotal)
		}

		fmt.Fprintf(
			w, statmsg, users, subscriptions, dbSize,
			phrasesTotal, phrasesAvg,
			scoretotal, scoretotalAvg,
			studiesTotal, studiesAvg,
			dueStudiesTotal, dueStudiesAvg,
			importsTotal, importsAvg,
			notifiesTotal, notifiesAvg,
			zeroscore, zeroscoreAvg,
			newphrasesTotal, newphrasesAvg,
			warnings,
		)
		return nil
	})
}

// Sum all values in a bucket
func sum(b *bolt.Bucket, fn func([]byte) int) (int, error) {
	sum := 0
	err := b.ForEach(func(_, v []byte) error {
		sum += fn(v)
		return nil
	})
	return sum, err
}
func simplesum(v []byte) int {
	return int(btoi(v))
}
func count64(v []byte) int {
	return len(v) / 8
}
