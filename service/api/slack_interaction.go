package api

type SlackInteraction struct {
	Type        string       `json:"type"`
	Token       string       `json:"token"`
	ActionTs    string       `json:"text"`
	User        SlackUser    `json:"user"`
	Channel     SlackChannel `json:"channel"`
	Ts          string       `json:"ts"`
	CallbackId  string       `json:"callback_id"`
	TriggerId   string       `json:"trigger_id"`
	ResponseUrl string       `json:"response_url"`
	MessageTs   string       `json:"message_ts"`
	View        *SlackView   `json:"view,omitempty"`
}
