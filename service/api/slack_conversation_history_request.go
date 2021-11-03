package api

import (
	"encoding/json"
	"fmt"
)

// SlackConversationHistoryRequest is a request to search for a message.
type SlackConversationHistoryRequest struct {
	// Channel is the conversion_id.
	Channel string `json:"channel"`
	// Latest is the TS value.
	Latest string `json:"latest"`
	// Limit is the number of results to return.
	Limit int `json:"limit"`
	// Inclusive is whether is should include the provided ts value.
	Inclusive bool `json:"inclusive"`
}

// UnmarshalJSON will unmarshal ReceivableContactResponse.
func (r *SlackConversationHistoryRequest) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		return fmt.Errorf("no bytes to unmarshal")
	}

	var resp SlackConversationHistoryResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		return err
	}

	return nil
}
