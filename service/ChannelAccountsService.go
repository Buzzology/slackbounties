package service

import (
	"context"
	"fmt"

	"github.com/buzzology/slack_bot/db"
	"github.com/buzzology/slack_bot/service/api"
	"github.com/buzzology/slack_bot/types"
)

type IChannelAccounts interface {
	DailyLeaderboard(ctx context.Context, channelId string) (*api.SlackBlocks, error)
	WeeklyLeaderboard(ctx context.Context, channelId string) (*api.SlackBlocks, error)
	YearlyLeaderboard(ctx context.Context, channelId string) (*api.SlackBlocks, error)
	AllTimeLeaderboard(ctx context.Context, channelId string) (*api.SlackBlocks, error)
}

type ChannelAccountsService struct {
	config              *Config
	channelAccountsRepo db.ChannelAccountsRepo
	apiClient           api.SlackApiClient
}

func NewChannelAccountsService(
	config *Config,
	channelAccountsRepo db.ChannelAccountsRepo,
	apiClient api.SlackApiClient,
) *ChannelAccountsService {
	return &ChannelAccountsService{
		config:              config,
		channelAccountsRepo: channelAccountsRepo,
		apiClient:           apiClient,
	}
}

// DailyLeaderboard generates a leaderboard for the day's current earnings.
func (s *ChannelAccountsService) DailyLeaderboard(ctx context.Context, channelId string) (*api.SlackBlocks, error) {
	channelAccounts, err := s.channelAccountsRepo.LeadersToday(channelId, 30, 10)
	if err != nil {
		return nil, err
	}

	leaderFields := api.SlackFieldsBlock{
		Type: "section",
	}

	// Add each of the leaders as a block.
	for index, channelAccount := range channelAccounts {
		leaderFields.Fields = append(
			leaderFields.Fields,
			generateLeaderField(
				channelAccount.UserId,
				index+1,
				channelAccount.EarnedToday,
			)...)
	}

	return GenerateLeaderboard("Daily Leaderboard", channelAccounts, leaderFields), nil
}

// WeeklyLeaderboard generates a leaderboard for the day's current earnings.
func (s *ChannelAccountsService) WeeklyLeaderboard(ctx context.Context, channelId string) (*api.SlackBlocks, error) {
	channelAccounts, err := s.channelAccountsRepo.LeadersThisWeek(channelId, 30, 10)
	if err != nil {
		return nil, err
	}

	leaderFields := api.SlackFieldsBlock{
		Type: "section",
	}

	// Add each of the leaders as a block.
	for index, channelAccount := range channelAccounts {
		leaderFields.Fields = append(
			leaderFields.Fields,
			generateLeaderField(
				channelAccount.UserId,
				index+1,
				channelAccount.EarnedThisWeek,
			)...)
	}

	return GenerateLeaderboard("Weekly Leaderboard", channelAccounts, leaderFields), nil
}

// YearlyLeaderboard generates a leaderboard for the day's current earnings.
func (s *ChannelAccountsService) YearlyLeaderboard(ctx context.Context, channelId string) (*api.SlackBlocks, error) {
	channelAccounts, err := s.channelAccountsRepo.LeadersThisYear(channelId, 30, 10)
	if err != nil {
		return nil, err
	}

	leaderFields := api.SlackFieldsBlock{
		Type: "section",
	}

	// Add each of the leaders as a block.
	for index, channelAccount := range channelAccounts {
		leaderFields.Fields = append(
			leaderFields.Fields,
			generateLeaderField(
				channelAccount.UserId,
				index+1,
				channelAccount.EarnedThisYear,
			)...)
	}

	return GenerateLeaderboard("Yearly Leaderboard", channelAccounts, leaderFields), nil
}

// AllTimeLeaderboard generates a leaderboard for the day's current earnings.
func (s *ChannelAccountsService) AllTimeLeaderboard(ctx context.Context, channelId string) (*api.SlackBlocks, error) {
	channelAccounts, err := s.channelAccountsRepo.LeadersAllTime(channelId, 30, 10)
	if err != nil {
		return nil, err
	}

	leaderFields := api.SlackFieldsBlock{
		Type: "section",
	}

	// Add each of the leaders as a block.
	for index, channelAccount := range channelAccounts {
		leaderFields.Fields = append(
			leaderFields.Fields,
			generateLeaderField(
				channelAccount.UserId,
				index+1,
				channelAccount.EarnedAllTime,
			)...)
	}

	return GenerateLeaderboard("All Time Leaderboard", channelAccounts, leaderFields), nil
}

func GenerateLeaderboard(
	title string,
	channelAccounts []*types.ChannelAccount,
	leaderFields api.SlackFieldsBlock,
) *api.SlackBlocks {
	// If there are none, just add a placeholder.
	if len(channelAccounts) == 0 {
		leaderFields.Fields = append(
			leaderFields.Fields,
			api.SlackBlock{
				Type: "mrkdwn",
				Text: "_No bounties earned or awarded yet._",
			},
		)
	}

	return &api.SlackBlocks{
		Blocks: []interface{}{
			&api.SlackBlock{
				Type: "header",
				Text: api.SlackBlock{
					Type: "plain_text",
					Text: title,
				},
			},
			&api.SlackBlockRawType{
				Type: "divider",
			},
			leaderFields,
		},
	}
}

func generateLeaderField(userId string, place int, earned int) []interface{} {
	return []interface{}{
		api.SlackBlock{
			Type: "mrkdwn",
			Text: fmt.Sprintf("%v) *<@%v>*", place, userId),
		},
		api.SlackBlock{
			Type: "plain_text",
			Text: fmt.Sprintf("%v", earned),
		},
	}
}
