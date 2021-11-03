package api

type SlackConversationHistoryResponse struct {
	Ok               bool                  `json:"ok"`
	Latest           string                `json:"latest"`
	Messages         []*SlackMessage       `json:"messages"`
	HasMore          bool                  `json:"has_more"`
	PinCount         int                   `json:"pin_count"`
	ResponseMetadata SlackResponseMetadata `json:"response_metadata"`
	Error            string                `json:"error"`
}
