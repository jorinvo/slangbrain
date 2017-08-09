package common

type Profile interface {
	Name() string
	Locale() string
	Timezone() float64
}
