package api

type SlackReactionEvent struct {
	Type     string    `json:"type"`
	User     string    `json:"user"`
	Reaction string    `json:"reaction"`
	ItemUser string    `json:"item_user"`
	Item     SlackItem `json:"item"`
}
