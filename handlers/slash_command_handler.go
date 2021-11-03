package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/buzzology/slack_bot/service/api"
	"github.com/buzzology/slack_bot/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// SlashCommandHandler handles and processes events received from slack.
func (h *SlackBotHandler) SlashCommandHandler(w http.ResponseWriter, r *http.Request) {
	var (
		slackBlocks *api.SlackBlocks
		err         error
	)

	ctx := context.Background()

	r.ParseForm()

	// Check the command type.
	switch r.FormValue("command") {
	case "/bountyme":
		{
			slackBlocks, err = h.handleSlashCommandMe(ctx, r.FormValue("user_id"), r.FormValue("channel_id"))
		}
	case "/bountyemotes":
		{
			slackBlocks, err = h.handleSlashCommandEmotes(ctx)
		}
	case "/bountydaily":
		{
			slackBlocks, err = h.handleSlashCommandDailyLeaders(ctx, r.FormValue("channel_id"))
		}
	case "/bountyweekly":
		{
			slackBlocks, err = h.handleSlashCommandWeeklyLeaders(ctx, r.FormValue("channel_id"))
		}
	case "/bountyyearly":
		{
			slackBlocks, err = h.handleSlashCommandYearlyLeaders(ctx, r.FormValue("channel_id"))
		}
	case "/bountyalltime":
		{
			slackBlocks, err = h.handleSlashCommandAllTimeLeaders(ctx, r.FormValue("channel_id"))
		}
	case "/bountyconfig":
		{

		}
	default:
		h.log.Warnf("unrecognised slack command: %v", r.FormValue("command"))
		w.Write([]byte("Unknown command"))
	}

	if err != nil {
		h.log.WithError(err).WithFields(
			logrus.Fields{
				"command":   r.FormValue("command"),
				"user":      r.FormValue("user_id"),
				"user_name": r.FormValue("user_name"),
			}).Error("Unable to process slash command.")

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	message, err := json.Marshal(slackBlocks)
	if err != nil {
		h.log.WithError(err).WithFields(
			logrus.Fields{
				"command":   r.FormValue("command"),
				"user":      r.FormValue("user_id"),
				"user_name": r.FormValue("user_name"),
			}).Error("Failed to serialize slack blocks.")

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Printf("%q", message)

	w.Header().Set("Content-Type", "application/json")
	w.Write(message)
}

// handleSlashCommandDailyLeaders the current leaderboard for today.
func (h *SlackBotHandler) handleSlashCommandDailyLeaders(
	ctx context.Context,
	channelId string,
) (*api.SlackBlocks, error) {
	return h.channelAccountsService.DailyLeaderboard(ctx, channelId)
}

// handleSlashCommandWeeklyLeaders the current leaderboard for today.
func (h *SlackBotHandler) handleSlashCommandWeeklyLeaders(
	ctx context.Context,
	channelId string,
) (*api.SlackBlocks, error) {
	return h.channelAccountsService.WeeklyLeaderboard(ctx, channelId)
}

// handleSlashCommandYearlyLeaders the current leaderboard for today.
func (h *SlackBotHandler) handleSlashCommandYearlyLeaders(
	ctx context.Context,
	channelId string,
) (*api.SlackBlocks, error) {
	return h.channelAccountsService.YearlyLeaderboard(ctx, channelId)
}

// handleSlashCommandAllTimeLeaders the current leaderboard for today.
func (h *SlackBotHandler) handleSlashCommandAllTimeLeaders(
	ctx context.Context,
	channelId string,
) (*api.SlackBlocks, error) {
	return h.channelAccountsService.AllTimeLeaderboard(ctx, channelId)
}

// handleSlashCommandEmotes shows the current user what each emote does.
func (h *SlackBotHandler) handleSlashCommandEmotes(
	ctx context.Context,
) (*api.SlackBlocks, error) {
	bountyReactionFields := api.SlackFieldsBlock{
		Type: "section",
	}
	for _, boostReactionSetting := range h.config.BoostReactions {
		bountyReactionFields.Fields = append(
			bountyReactionFields.Fields,
			api.SlackBlock{
				Type: "mrkdwn",
				Text: fmt.Sprintf("*Boost +%v*", boostReactionSetting.BoostValue),
			},
			api.SlackBlock{
				Type: "plain_text",
				Text: fmt.Sprintf(":%v:", boostReactionSetting.Emote),
			},
		)
	}

	return &api.SlackBlocks{
		Blocks: []interface{}{
			&api.SlackBlock{
				Type: "header",
				Text: api.SlackBlock{
					Type: "plain_text",
					Text: "Emotes Overview",
				},
			},
			&api.SlackBlockRawType{
				Type: "divider",
			},
			&api.SlackFieldsBlock{
				Type: "section",
				Fields: []interface{}{
					api.SlackBlock{
						Type: "mrkdwn",
						Text: "*Award Bounty*",
					},
					api.SlackBlock{
						Type: "plain_text",
						Text: fmt.Sprintf(":%v:", h.config.ReleaseBountyReaction),
					},
					api.SlackBlock{
						Type: "mrkdwn",
						Text: "*Task Completed*",
					},
					api.SlackBlock{
						Type: "plain_text",
						Text: fmt.Sprintf(":%v:", h.config.TaskCompletedByMeReaction),
					},
				},
			},
			bountyReactionFields,
		},
	}, nil
}

// handleSlashCommandMe displays stats for the current user.
func (h *SlackBotHandler) handleSlashCommandMe(
	ctx context.Context,
	userId string,
	channelId string,
) (*api.SlackBlocks, error) {
	// Retrieve the user's account.
	channelAccounts, _, err := h.channelAccountsRepo.List(
		&types.ListChannelAccountsFilter{
			UserId:    userId,
			ChannelId: channelId,
		},
		1,
		"",
		"",
	)

	if err != nil {
		return nil, errors.Wrap(err, "failed to check for channel account")
	}

	var channelAccount *types.ChannelAccount
	if len(channelAccounts) > 0 {
		channelAccount = channelAccounts[0]
	} else {
		channelAccount = &types.ChannelAccount{}
	}

	return &api.SlackBlocks{
		Blocks: []interface{}{
			&api.SlackBlock{
				Type: "header",
				Text: api.SlackBlock{
					Type: "plain_text",
					Text: "Your Bounty :" + h.config.TaskCompletedByMeReaction + ":",
				},
			},
			&api.SlackBlockRawType{
				Type: "divider",
			},
			&api.SlackFieldsBlock{
				Type: "section",
				Fields: []interface{}{
					api.SlackBlock{
						Type: "mrkdwn",
						Text: "*Current Balance*",
					},
					api.SlackBlock{
						Type: "plain_text",
						Text: fmt.Sprint(channelAccount.Balance),
					},
				},
			},
			&api.SlackFieldsBlock{
				Type: "section",
				Fields: []interface{}{
					api.SlackBlock{
						Type: "mrkdwn",
						Text: "*Spent Today*",
					},
					api.SlackBlock{
						Type: "plain_text",
						Text: fmt.Sprint(channelAccount.SpentToday),
					},
					api.SlackBlock{
						Type: "mrkdwn",
						Text: "*Spent this Week*",
					},
					api.SlackBlock{
						Type: "plain_text",
						Text: fmt.Sprint(channelAccount.SpentThisWeek),
					},
					api.SlackBlock{
						Type: "mrkdwn",
						Text: "*Spent this Year*",
					},
					api.SlackBlock{
						Type: "plain_text",
						Text: fmt.Sprint(channelAccount.SpentThisYear),
					},
					api.SlackBlock{
						Type: "mrkdwn",
						Text: "*Spent all Time*",
					},
					api.SlackBlock{
						Type: "plain_text",
						Text: fmt.Sprint(channelAccount.SpentAllTime),
					},
				},
			},
			&api.SlackFieldsBlock{
				Type: "section",
				Fields: []interface{}{
					api.SlackBlock{
						Type: "mrkdwn",
						Text: "*Earned Today*",
					},
					api.SlackBlock{
						Type: "plain_text",
						Text: fmt.Sprint(channelAccount.EarnedToday),
					},
					api.SlackBlock{
						Type: "mrkdwn",
						Text: "*Earned this Week*",
					},
					api.SlackBlock{
						Type: "plain_text",
						Text: fmt.Sprint(channelAccount.EarnedThisWeek),
					},
					api.SlackBlock{
						Type: "mrkdwn",
						Text: "*Earned this Year*",
					},
					api.SlackBlock{
						Type: "plain_text",
						Text: fmt.Sprint(channelAccount.EarnedThisYear),
					},
					api.SlackBlock{
						Type: "mrkdwn",
						Text: "*Earned all Time*",
					},
					api.SlackBlock{
						Type: "plain_text",
						Text: fmt.Sprint(channelAccount.EarnedAllTime),
					},
				},
			},
		},
	}, nil
}
