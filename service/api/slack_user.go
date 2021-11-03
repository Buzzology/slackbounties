package api

type SlackUser struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	TeamId   string `json:"team_id"`
}
