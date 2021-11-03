package api

type SlackErrorResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}
