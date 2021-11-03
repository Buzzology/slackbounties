package api

type SlackReactionAddedEvent struct {
	EventID string              `event_id:"event_id"`
	Type    string              `json:"type"`
	Event   *SlackReactionEvent `json:"event"`
	Token   string              `json:"token"`
	TeamID  string              `json:"team_id"`
}
