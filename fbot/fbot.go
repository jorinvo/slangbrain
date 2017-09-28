// Package fbot can be used to communicate with a Facebook Messenger bot.
// The supported API is limited to only the required use cases
// and the data format is abstracted accordingly.
package fbot

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
)

const defaultAPI = "https://graph.facebook.com/v2.6"

// Client can be used to communicate with a Messenger bot.
type Client struct {
	token       string
	secretProof string
	api         string
}

// API can be passed to New for sending requests to a different URL.
// Must not contain trailing slash.
func API(url string) func(*Client) {
	return func(c *Client) {
		if url != "" {
			c.api = url
		}
	}
}

// New rerturns a new client with credentials set up.
// Optionally pass API to overwrite the default API URL.
func New(token, secret string, options ...func(*Client)) Client {
	// Generate secret proof. See https://developers.facebook.com/docs/graph-api/securing-requests/#appsecret_proof
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(token))

	c := Client{
		token:       token,
		secretProof: hex.EncodeToString(mac.Sum(nil)),
		api:         defaultAPI,
	}
	for _, option := range options {
		option(&c)
	}
	return c
}

// Button describes a button that can be send with a Button Template.
// Use URLButton or PayloadButton for initialization.
type Button interface{}

// Reply describes a text quick reply.
type Reply struct {
	// Text is the text on the button visible to the user
	Text string
	// Payload is a string to identify the quick reply event internally in your application.
	Payload string
}

// Helper to check for errors in reply
func checkError(r io.Reader) error {
	var qr struct {
		Error *struct {
			Message   string `json:"message"`
			Type      string `json:"type"`
			Code      int    `json:"code"`
			FBTraceID string `json:"fbtrace_id"`
		} `json:"error"`
		Result string `json:"result"`
	}

	err := json.NewDecoder(r).Decode(&qr)
	if qr.Error != nil {
		err = fmt.Errorf("Facebook error : %s", qr.Error.Message)
	}
	return err
}
