package service

import (
	"context"
	"time"

	"github.com/buzzology/slack_bot/db"
	"github.com/buzzology/slack_bot/service/api"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type IBotState interface {
	Tickover(ctx context.Context, now *time.Time) error
}

type BotStateService struct {
	config                 *Config
	botStateRepo           db.BotStateRepo
	channelAccountsRepo    db.ChannelAccountsRepo
	apiClient              api.SlackApiClient
	channelAccountsService *ChannelAccountsService
	log                    *logrus.Logger
}

func NewBotStateService(
	config *Config,
	botStateRepo db.BotStateRepo,
	channelAccountsRepo db.ChannelAccountsRepo,
	channelAccountsService *ChannelAccountsService,
	apiClient api.SlackApiClient,
	log *logrus.Logger,
) *BotStateService {
	return &BotStateService{
		config:                 config,
		botStateRepo:           botStateRepo,
		channelAccountsRepo:    channelAccountsRepo,
		channelAccountsService: channelAccountsService,
		apiClient:              apiClient,
		log:                    log,
	}
}

// Tickover checks for and then actions any tasks associated with daily, weekly, ..., tickovers.
func (s *BotStateService) Tickover(ctx context.Context, now time.Time) error {
	// TODO: Wrap all of this in a transaction.
	botState, err := s.botStateRepo.Get()
	if err != nil {
		return errors.Wrap(err, "Failed to retrieve bot state while performing tickover.")
	}

	// Retrieve all channel accounts.
	channelIds, err := s.channelAccountsRepo.DistinctChannels()
	if err != nil {
		return errors.Wrap(err, "Failed to retrieve channel accounts for tickover.")
	}

	// Reset daily if required.
	if botState.DayTickover.AsTime().Before(now) {
		// Send a leaderboard to each channel.
		for _, channelId := range channelIds {
			leaderboard, err := s.channelAccountsService.DailyLeaderboard(ctx, channelId)
			if err != nil {
				s.log.Errorf("failed to show the daily leaderboard for: %v", channelId, err)
				continue
			}

			if _, err := s.apiClient.SendMessage(ctx, &api.SlackPostMessageRequest{
				Blocks:  leaderboard.Blocks,
				Channel: channelId,
			}); err != nil {
				s.log.Errorf("failed to send the daily leaderboard for: %v", channelId, err)
				continue
			}
		}

		if err = s.channelAccountsRepo.ResetDaily(); err != nil {
			return errors.Wrap(err, "Failed to reset daily channel accounts.")
		}

		// Set when the next tickover will occur.
		botState.DayTickover = *timestamppb.New(botState.DayTickover.AsTime().Add(time.Hour * 24))

		// We apply decay and income daily as well.
		if err = s.botStateRepo.ApplyIncomeAndDecay(s.config.DailyDecay, s.config.DailyIncome); err != nil {
			return err
		}
	}

	// Reset weekly if required.
	if botState.WeekTickover.AsTime().Before(now) {
		// Send a leaderboard to each channel.
		for _, channelId := range channelIds {
			leaderboard, err := s.channelAccountsService.WeeklyLeaderboard(ctx, channelId)
			if err != nil {
				s.log.Errorf("failed to display weekly leaderboard for: %v", channelId, err)
				continue
			}

			if _, err := s.apiClient.SendMessage(ctx, &api.SlackPostMessageRequest{
				Blocks:  leaderboard.Blocks,
				Channel: channelId,
			}); err != nil {
				s.log.Errorf("failed to send the weekly leaderboard for: %v", channelId, err)
				continue
			}
		}

		if err = s.channelAccountsRepo.ResetWeekly(); err != nil {
			return errors.Wrap(err, "failed to reset weekly")
		}

		// Set when the next tickover should occur.
		botState.WeekTickover = *timestamppb.New(botState.DayTickover.AsTime().Add(time.Hour * 730))
	}

	// Reset yearly if required.
	if botState.YearTickover.AsTime().Before(now) {
		// Send a leaderboard to each channel.
		for _, channelId := range channelIds {
			leaderboard, err := s.channelAccountsService.YearlyLeaderboard(ctx, channelId)
			if err != nil {
				s.log.Errorf("failed to display yearly leaderboard for: %v", channelId, err)
				continue
			}

			if _, err := s.apiClient.SendMessage(ctx, &api.SlackPostMessageRequest{
				Blocks:  leaderboard.Blocks,
				Channel: channelId,
			}); err != nil {
				s.log.Errorf("failed to send the yearly leaderboard for: %v", channelId, err)
				continue
			}
		}

		if err = s.channelAccountsRepo.ResetYearly(); err != nil {
			return errors.Wrap(err, "failed to reset yearly")
		}

		// Set when the next tickover should occur.
		botState.YearTickover = *timestamppb.New(botState.YearTickover.AsTime().Add(time.Hour * 8760))
	}

	// Update the bot's state before returning.
	_, err = s.botStateRepo.Update(botState)
	if err != nil {
		return errors.Wrap(err, "failed to update bot_state when performing tickover.")
	}

	return nil
}
