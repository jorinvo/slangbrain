package fbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/jorinvo/slangbrain/common"
)

// URL to fetch the profile from;
// is relative to the API URL.
const profileURL = "%s/%d?fields=first_name,locale,timezone&access_token=%s"

// profile has all public user information we need;
// needs to be in sync with the URL above.
type profile struct {
	data profileData
}
type profileData struct {
	Name     string `json:"first_name"`
	Locale   string `json:"locale"`
	Timezone int    `json:"timezone"`
}

func (p profile) Name() string   { return p.data.Name }
func (p profile) Locale() string { return p.data.Locale }
func (p profile) Timezone() int  { return p.data.Timezone }

// GetProfile fetches a user profile for an ID.
func (c Client) GetProfile(id int64) (common.Profile, error) {
	url := fmt.Sprintf(profileURL, c.api, id, c.token)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var p profileData
	if err = json.Unmarshal(content, &p); err != nil {
		return profile{p}, err
	}

	return profile{p}, checkError(bytes.NewReader(content))
}
