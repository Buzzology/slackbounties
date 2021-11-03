package api

// SlackDeleteMessageResponse is a response to deleting a message.
type SlackDeleteMessageResponse struct {
	Ok bool   `json:"ok"`
	Ts string `json:"ts"`
	// Channel is the user or channel to which the message should be sent.
	Channel          string                `json:"channel"`
	Message          SlackMessage          `json:"message"`
	ResponseMetadata SlackResponseMetadata `json:"response_metadata"`
	Error            string                `json:"error"`
}
