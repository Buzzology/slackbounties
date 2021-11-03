package api

type SlackItem struct {
	Type        string `json:"type"`
	Channel     string `json:"channel"`
	File        string `json:"file"`
	FileComment string `json:"file_comment"`
	Ts          string `json:"ts"`
}
