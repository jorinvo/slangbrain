package translate

import "sort"

const defaultLang = "en_US"

// Translator is a service to get content in multiple languages.
// Always use .New() for initialization.
type Translator struct {
	data map[string]Content
}

// Content holds translated versions of messages and buttons.
type Content struct {
	Msg Msg
	Btn Btn
}

// New returns a translator with messages and buttons loaded in all available languages.
func New() Translator {
	langs := map[string]func() (Msg, buttonLabels){
		defaultLang: en,
		"en_GB":     en,
		"de_DE":     de,
		"th_TH":     th,
	}

	t := Translator{map[string]Content{}}
	for lang, fn := range langs {
		m, b := fn()
		t.data[lang] = Content{Msg: m, Btn: newBtn(b)}
	}

	return t
}

// Load returns the Content in the specified language.
// lang should be 5 letters, as in "en_US".
// Returns default language, if lang is unknown.
// Simply pass an empty string to get the default language.
func (t Translator) Load(lang string) Content {
	c, ok := t.data[lang]
	if ok {
		return c
	}
	return t.data[defaultLang]
}

// Langs returns a list of all supported languages.
func (t Translator) Langs() []string {
	langs := []string{}
	for lang, _ := range t.data {
		langs = append(langs, lang)
	}
	sort.Strings(langs)
	return langs
}
