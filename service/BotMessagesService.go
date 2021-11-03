package service

import (
	"context"
	"fmt"

	"github.com/buzzology/slack_bot/db"
	"github.com/buzzology/slack_bot/service/api"
	"github.com/buzzology/slack_bot/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type IBotMessagesService interface {
	SendRemovableBotMessage(ctx context.Context, postMessageRequest *api.SlackPostMessageRequest) error
	RemoveSentBotMessageIfExists(
		ctx context.Context,
		targetUserId string,
		targetReaction string,
		targetMessageId string,
		targetChannelId string,
	) error
}

type BotMessagesService struct {
	config          *Config
	botMessagesRepo db.BotMessagesRepo
	apiClient       api.SlackApiClient
	log             *logrus.Logger
}

func NewBotMessagesService(
	config *Config,
	botMessagesRepo db.BotMessagesRepo,
	apiClient api.SlackApiClient,
	log *logrus.Logger,
) *BotMessagesService {
	return &BotMessagesService{
		config:          config,
		botMessagesRepo: botMessagesRepo,
		apiClient:       apiClient,
		log:             log,
	}
}

// RemoveSentBotMessageIfExists removes the relevant bot message if it exists.
func (s *BotMessagesService) RemoveSentBotMessageIfExists(
	ctx context.Context,
	targetUserId string,
	targetReaction string,
	targetMessageId string,
	targetChannelId string,
) error {
	// Check if there's a bot message that we need to remove.
	botMessages, _, err := s.botMessagesRepo.List(
		&types.ListBotMessagesFilter{
			MessageId: targetMessageId,
			UserId:    targetUserId,
			ChannelId: targetChannelId,
			Status:    1,
			Reaction:  targetReaction,
		},
		1,
		"",
	)
	if err != nil {
		return errors.Wrap(err, "failed to list bot messages when handling remove reaction event")
	}

	if len(botMessages) == 0 {
		s.log.Infof("does not need to be handled as we haven't recorded any messages: %v, %v, %v", targetReaction, targetMessageId, targetUserId)
		return nil
	}

	// Mark the message as removed
	botMessage := botMessages[0]
	botMessage.Status = 2

	// Remove the message via slack client.
	res, err := s.apiClient.DeleteSlackMessage(
		ctx,
		&api.SlackDeleteMessageRequest{
			MessageTs: botMessage.SentMessageId,
			Channel:   botMessage.ChannelId,
		},
	)

	if err != nil {
		return errors.Wrap(err, "the api request to delete a previously sent slack message failed")
	}

	if !res.Ok {
		return fmt.Errorf("deleting a slack message returned an errors: %v", botMessage.Id)
	}

	// Mark the message as removed in the repo.
	if err = s.botMessagesRepo.Update(botMessage); err != nil {
		return errors.Wrap(err, "failed to update the status of a removed message")
	}

	return nil
}

// SendRemovableBotMessage sends a message to the user and allows for it to be removed it the action is corrected.
func (s *BotMessagesService) SendRemovableBotMessage(
	ctx context.Context,
	postMessageRequest *api.SlackPostMessageRequest,
	targetUserId string,
	targetReaction string,
) error {
	// Post reply to message that the bounty has been awarded tagging the awarder.
	res, err := s.apiClient.SendMessage(
		ctx,
		postMessageRequest,
	)
	if err != nil {
		return errors.Wrap(err, "failed to send slack message")
	}

	if !res.Ok {
		return errors.New("slack send message returned an error response")
	}

	// Save the message to the database if we've been provided with a reaction.
	if targetReaction != "" {
		if err = s.botMessagesRepo.Create(
			&types.BotMessage{
				MessageId:     res.Message.Ts,
				SentMessageId: postMessageRequest.ThreadTs,
				ChannelId:     postMessageRequest.Channel,
				Reaction:      targetReaction,
				TargetUserId:  targetUserId,
				Status:        1,
			},
		); err != nil {
			return errors.Wrap(err, "failed to save removable bot message after sending")
		}
	} else {
		// TODO: Send this message as a PM, there is no public reaction to remove.
	}

	return nil
}
