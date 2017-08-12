package common

// Profile abstracts a user profile.
// It is only used for reading information.
// Can be read from remote or from cache.
type Profile interface {
	Name() string
	Locale() string
	Timezone() float64
}
