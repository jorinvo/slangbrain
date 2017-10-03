package bot

import (
	"github.com/jorinvo/slangbrain/brain"
	"github.com/jorinvo/slangbrain/scope"
)

func (b bot) getUser(id int64) scope.User {
	fetcher := func() (brain.Profile, error) {
		fp, err := b.client.GetProfile(id)
		return brain.Profile(fp), err
	}
	return scope.Get(id, b.store, b.content, b.err, fetcher)
}
