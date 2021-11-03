package api

type SlackMessage struct {
	Type string `json:"type"`
	User string `json:"user"`
	Text string `json:"text"`
	Ts   string `json:"ts"`
}
