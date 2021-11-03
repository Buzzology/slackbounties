package api

// SlackDeleteMessageRequest is a request to delete a message.
type SlackDeleteMessageRequest struct {
	// Channel the channel the message belongs to.
	Channel string `json:"channel"`
	// MessageTs is the id of the message to delete.
	MessageTs string `json:"ts"`
}
