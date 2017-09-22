package brain

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"

	bolt "github.com/coreos/bbolt"
	"github.com/jorinvo/slangbrain/common"
)

// Wrap data as common.Profile.
type profile struct {
	data profileData
}

type profileData struct {
	Name      string
	Locale    string
	Timezone  int
	CacheTime time.Time
}

func (p profile) Name() string   { return p.data.Name }
func (p profile) Locale() string { return p.data.Locale }
func (p profile) Timezone() int  { return p.data.Timezone }

// GetProfile fetches a cached profile.
// Returns ErrNotFound if none found or cache is older than profileMaxCacheTime.
func (store Store) GetProfile(id int64) (common.Profile, error) {
	var p profileData
	err := store.db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket(bucketProfiles).Get(itob(id))
		if v == nil {
			return ErrNotFound
		}
		return gob.NewDecoder(bytes.NewReader(v)).Decode(&p)
	})
	if err != nil {
		if err == ErrNotFound {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get profile for id %d: %v", id, err)
	}
	// Check if expired
	if time.Now().Sub(p.CacheTime) > profileMaxCacheTime {
		return nil, ErrNotFound
	}
	return profile{p}, nil
}

// SetProfile caches a profile.
// Pass the caching time for easier testing.
func (store Store) SetProfile(id int64, p common.Profile, cachedAt time.Time) error {
	data := profileData{
		Name:      p.Name(),
		Locale:    p.Locale(),
		Timezone:  p.Timezone(),
		CacheTime: cachedAt,
	}
	err := store.db.Update(func(tx *bolt.Tx) error {
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(data); err != nil {
			return err
		}
		return tx.Bucket(bucketProfiles).Put(itob(id), buf.Bytes())
	})
	if err != nil {
		return fmt.Errorf("failed to set profile for id %d: %v: %v", id, data, err)
	}
	return nil
}
