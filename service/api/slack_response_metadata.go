package api

type SlackResponseMetadata struct {
	NextCursor string   `json:"next_cursor"`
	Messages   []string `json:"messages"`
	Warnings   []string `json:"warnings"`
}
