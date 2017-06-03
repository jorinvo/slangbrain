package fbot

import (
	"encoding/json"
	"fmt"
	"io"
)

// Can be overwritten for testing
const defaultAPI = "https://graph.facebook.com/v2.6"

// Client is the client which manages communication with the Messenger Platform API.
type Client struct {
	token       string
	verifyToken string
	// API URL can be changed for testing.
	// Must not contain trailing slash.
	API string
}

// New ...
func New(token, verifyToken string) Client {
	return Client{
		token:       token,
		verifyToken: verifyToken,
		API:         defaultAPI,
	}
}

// Button describes a text quick reply.
type Button struct {
	// Text is the reply title
	Text string
	// Payload is the reply information
	Payload string
}

// Helper to check for errors in reply
func checkFacebookError(r io.Reader) error {
	var err error

	var qr queryResponse
	err = json.NewDecoder(r).Decode(&qr)
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
