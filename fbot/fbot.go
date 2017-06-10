// Package fbot can be used to communicate with a Facebook Messenger bot.
// The supported API is limited the only the required use cases
// and the data format is abstracted accordingly.
package fbot

import (
	"encoding/json"
	"fmt"
	"io"
)

const defaultAPI = "https://graph.facebook.com/v2.6"

// Client can be used to communicate with a Messenger bot.
type Client struct {
	token       string
	verifyToken string
	// API URL can be changed for testing.
	// Must not contain trailing slash.
	API string
}

// New rerturns a new client with credentials set up.
func New(token, verifyToken string) Client {
	return Client{
		token:       token,
		verifyToken: verifyToken,
		API:         defaultAPI,
	}
}

// Button describes a text quick reply.
type Button struct {
	// Text is the text on the button visible to the user
	Text string
	// Payload is a string to identify the quick reply event internally in your application.
	Payload string
}

// Helper to check for errors in reply
func checkError(r io.Reader) error {
	var qr queryResponse
	err := json.NewDecoder(r).Decode(&qr)
	if qr.Error != nil {
		err = fmt.Errorf("Facebook error : %s", qr.Error.Message)
	}
	return err
}

type queryResponse struct {
	Error  *queryError `json:"error"`
	Result string      `json:"result"`
}

type queryError struct {
	Message   string `json:"message"`
	Type      string `json:"type"`
	Code      int    `json:"code"`
	FBTraceID string `json:"fbtrace_id"`
}
