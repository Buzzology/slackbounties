package service

import (
	"github.com/buzzology/slack_bot/service/api"
)

type Config struct {
	// ApiConfig is the config used for the slack api.
	ApiConfig                 *api.ApiConfig
	BoostReactions            []*BoostReactionValue
	ReleaseBountyReaction     string
	TaskCompletedByMeReaction string
	DailyDecay                int
	DailyIncome               int
	DbConnection              string
	DocumentationUrl          string
}

// NewConfig returns a new instance of config.
func NewConfig() *Config {
	return &Config{
		DbConnection: "",
		ApiConfig: &api.ApiConfig{
			Endpoint: "https://slack.com/api",
			Token:    "<enter-bot-token-here-or-use-toml-config>",
		},
		BoostReactions: []*BoostReactionValue{
			{Emote: "dollar", BoostValue: 1},
			{Emote: "money_with_wings", BoostValue: 2},
			{Emote: "moneybag", BoostValue: 3},
			{Emote: "take_my_money", BoostValue: 4},
			{Emote: "moneyparrot", BoostValue: 5},
		},
		ReleaseBountyReaction:     "medal",
		TaskCompletedByMeReaction: "white_check_mark",
		DailyDecay:                2,
		DailyIncome:               1, // NOTE: This is also re-used as starting balance when creating a new account.
	}
}

type BoostReactionValue struct {
	Emote      string
	BoostValue int
}
