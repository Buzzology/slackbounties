package api

type SlackViewsOpenRequest struct {
	TriggerId string     `json:"trigger_id"`
	View      *SlackView `json:"view"`
}

type SlackView struct {
	Type            string            `json:"type"`
	CallbackId      string            `json:"callback_id"`
	Title           interface{}       `json:"title"`
	Blocks          interface{}       `json:"blocks"`
	Submit          *SlackBlockSubmit `json:"submit,omitempty"`
	PrivateMetadata string            `json:"private_metadata"`
	State           *SlackViewState   `json:"state,omitempty"`
}

type SlackViewState struct {
	Values map[string]interface{} `json:"values,omitempty"`
}
