package fbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const profileURL = "%s/%d?fields=first_name,last_name,profile_pic,locale,timezone,gender&access_token=%s"

// Profile is the public information of a Facebook user
type Profile struct {
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	PictureURL string  `json:"profile_pic"`
	Gender     string  `json:"gender"`
	Locale     string  `json:"locale"`
	Timezone   float64 `json:"timezone"`
}

// GetProfile retrieves the Facebook user associated with that ID
func (c Client) GetProfile(id int64) (Profile, error) {
	var p Profile

	url := fmt.Sprintf(profileURL, c.API, id, c.token)
	resp, err := http.Get(url)
	if err != nil {
		return p, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return p, err
	}

	if err = json.Unmarshal(content, &p); err != nil {
		return p, err
	}

	return p, checkError(bytes.NewReader(content))
}
