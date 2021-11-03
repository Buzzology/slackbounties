package api

import (
	"encoding/json"
	"fmt"
)

// SlackPostMessageRequest is a request to search for a message.
type SlackPostMessageRequest struct {
	// Text is the text to send.
	Text string `json:"text"`
	// Channel is the user or channel to which the message should be sent.
	Channel string `json:"channel"`
	// ThreadTs is used to reply to a specific message.
	ThreadTs string `json:"thread_ts"`
	// Blocks is used to send a structured message instead of plain text.
	Blocks interface{} `json:"blocks"`
}

// UnmarshalJSON will unmarshal ReceivableContactResponse.
func (r *SlackPostMessageRequest) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		return fmt.Errorf("no bytes to unmarshal")
	}

	var resp SlackPostMessageRequest
	if err := json.Unmarshal(b, &resp); err != nil {
		return err
	}

	return nil
}
