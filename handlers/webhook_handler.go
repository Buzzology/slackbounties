package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/buzzology/slack_bot/service/api"
	"github.com/buzzology/slack_bot/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// WebhookHandler handles and processes events received from slack.
func (h *SlackBotHandler) WebhookHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	logrus.Println("Slack webhook received: ", r.RequestURI)

	defer r.Body.Close()

	// Retrieve the body.
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logrus.Errorf("Failed to read request body: %v, %v", r.RequestURI, err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var event map[string]interface{}
	if err = json.Unmarshal(body, &event); err != nil {
		logrus.Errorf("Failed to convert event to json: %v, %v", string(body), err.Error())
		return
	}

	// TODO: We need to verify the slack signature etc.

	switch event["type"] {
	case "url_verification":
		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-type", "text/plain")
		w.Write([]byte(event["challenge"].(string)))
	case "event_callback":
		w.WriteHeader(http.StatusOK)

		// Retrieve the event body
		h.handleEventCallback(ctx, event, body)
		return
	default:
		logrus.Warningf("Unknown event type: %v", event["type"])
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *SlackBotHandler) handleEventCallback(ctx context.Context, event map[string]interface{}, rawEvent []byte) {
	// Check the event type.
	switch event["event"].(map[string]interface{})["type"] {
	case "reaction_added":

		reactionAddedEvent := &api.SlackReactionAddedEvent{}
		if err := json.Unmarshal(rawEvent, &reactionAddedEvent); err != nil {
			logrus.Errorf("Failed to ummarshall reaction_added event: %v, %v", string(rawEvent), err.Error())
			return
		}

		if err := h.handleReactionAddedEvent(ctx, reactionAddedEvent); err != nil {
			h.log.WithError(err).WithFields(
				logrus.Fields{
					"event":    reactionAddedEvent.EventID,
					"reaction": reactionAddedEvent.Event.Reaction,
					"user":     reactionAddedEvent.Event.User,
				}).Error("Unable to process reaction added event.")

			return
		}
	case "reaction_removed":
		reactionRemovedEvent := &api.SlackReactionRemovedEvent{}
		if err := json.Unmarshal(rawEvent, &reactionRemovedEvent); err != nil {
			logrus.Errorf("Failed to ummarshall reaction_removed event: %v, %v", string(rawEvent), err.Error())
			return
		}

		if err := h.handleReactionRemovedEvent(ctx, reactionRemovedEvent); err != nil {
			h.log.WithError(err).WithFields(
				logrus.Fields{
					"event":    reactionRemovedEvent.EventID,
					"reaction": reactionRemovedEvent.Event.Reaction,
					"user":     reactionRemovedEvent.Event.User,
				}).Error("Unable to process reaction removed event.")

			return
		}
	default:
		logrus.Errorf("Unknown event -> type: %v", event["event"].(map[string]interface{})["type"])
		return
	}
}

func (h *SlackBotHandler) handleReactionRemovedEvent(ctx context.Context, event *api.SlackReactionRemovedEvent) error {
	// Ensure that the target is a message type
	if event.Event.Item.Type != "message" {
		h.log.Infof("skipping as only message reactions are handled: %v, %v, %v", event.Event.Reaction, event.EventID, event.Type)
		return nil
	}

	if err := h.botMessagesService.RemoveSentBotMessageIfExists(
		ctx,
		event.Event.ItemUser,
		event.Event.Reaction,
		event.Event.Item.Ts,
		event.Event.Item.Channel,
	); err != nil {
		return errors.Wrap(err, "failed to remove sent message if it exists.")
	}

	return nil
}

func (h *SlackBotHandler) handleReactionAddedEvent(ctx context.Context, event *api.SlackReactionAddedEvent) error {
	// Ensure that the target is a message type
	if event.Event.Item.Type != "message" {
		h.log.Infof("skipping as only message reactions are handled: %v, %v, %v", event.Event.Reaction, event.EventID, event.Type)
		return nil
	}

	// Check if it's a boost reaction.
	for _, boostReaction := range h.config.BoostReactions {
		if boostReaction.Emote == event.Event.Reaction {
			return h.boostBounty(ctx, event, boostReaction.BoostValue)
		}
	}

	// Check if it's a "task completed" reaction.
	if strings.EqualFold(h.config.TaskCompletedByMeReaction, event.Event.Reaction) {
		return h.claimBounty(ctx, event)
	}

	// Check if it's an "award bounty" reaction.
	if strings.EqualFold(h.config.ReleaseBountyReaction, event.Event.Reaction) {
		// NOTE: We don't assign a user when using the emote, it will attempt to use awarded_to.
		return h.awardBounty(ctx, event.Event.Item.Ts, "", event.Event.User, event.Event.Reaction)
	}

	h.log.Infof("not a handled reaction event type: %v, %v", event.Event.Reaction, event.EventID)
	return nil
}

// claimBounty should be performed by the task completer. The bounty is not awarded until the award bounty emote is received.
func (h *SlackBotHandler) claimBounty(ctx context.Context, event *api.SlackReactionAddedEvent) error {
	// To start, we'll retrieve the bounty.
	messageBounties, _, err := h.messageBountiesRepo.List(
		&types.ListMessageBountiesFilter{
			MessageId: event.Event.Item.Ts,
			ChannelId: event.Event.Item.Channel,
		},
		1,
		"",
	)

	if err != nil {
		return errors.Wrap(err, "failed to retrieve the message bounty")
	}

	if len(messageBounties) == 0 {
		return fmt.Errorf("no bounty exists for message: %v, %v", event.Event.Item.Ts, event.Event.Item.Channel)
	}

	// Ensure that the bounty is still open.
	if messageBounties[0].Status != 1 {
		return fmt.Errorf("bounty can only be claimed while in an open state: %v, %v", event.Event.Item.Ts, event.Event.Item.Channel)
	}

	// Claim the bounty.
	messageBounties[0].AwardedTo = event.Event.User

	// Mark the bounty as awarded.
	if _, err = h.messageBountiesRepo.Update(messageBounties[0]); err != nil {
		return errors.Wrapf(err, "failed to claim bounty: %v", messageBounties[0].MessageId)
	}

	h.apiClient.SendMessage(
		ctx,
		&api.SlackPostMessageRequest{
			Text:     "<@" + event.Event.User + "> has completed the task!",
			Channel:  event.Event.Item.Channel,
			ThreadTs: event.Event.Item.Ts,
		})

	return nil
}

// awardBounty gives the current bounty on a message to the target user. If no target user is provided it will check if someone has attempted to claim it instead.
func (h *SlackBotHandler) awardBounty(
	ctx context.Context,
	messageId string,
	targetUserId string,
	currentUserId string,
	reaction string,
) error {

	// To start, we'll retrieve the bounty.
	messageBounties, _, err := h.messageBountiesRepo.List(
		&types.ListMessageBountiesFilter{
			MessageId: messageId,
		},
		1,
		"",
	)

	if err != nil {
		return errors.Wrap(err, "failed to retrieve the message bounty")
	}

	if len(messageBounties) == 0 {
		return fmt.Errorf("no bounty exists for message: %v", messageId)
	}

	// If we don't have a target user check if someone has already claimed it.
	if targetUserId == "" {
		if messageBounties[0].AwardedTo == "" {
			h.botMessagesService.SendRemovableBotMessage(
				ctx,
				&api.SlackPostMessageRequest{
					Text:     "Heads up <@" + currentUserId + ">! Nobody has claimed the bounty yet :" + h.config.TaskCompletedByMeReaction + ":. Please wait until the bounty is claimed or use the message option to award it directly.",
					Channel:  messageBounties[0].ChannelId,
					ThreadTs: messageBounties[0].MessageId,
				},
				currentUserId,
				reaction,
			)
			return fmt.Errorf("nobody has claimed this bounty yet")
		}
	} else {
		messageBounties[0].AwardedTo = targetUserId
	}

	// Check if it's already been awarded.
	if messageBounties[0].Status == 2 {
		// Check if the awarder is the same person who owns the message. If so we don't need to notify them (probably just adding the emote).
		if messageBounties[0].UserId != currentUserId {
			h.botMessagesService.SendRemovableBotMessage(
				ctx,
				&api.SlackPostMessageRequest{
					Text:     "Heads up <@" + currentUserId + ">! This bounty has already been awarded to <@" + messageBounties[0].AwardedTo + ">.",
					Channel:  messageBounties[0].ChannelId,
					ThreadTs: messageId,
				},
				currentUserId,
				reaction,
			)
		}

		return nil
	}

	// Ensure that the owner of the bounty is the one who awards it.
	if messageBounties[0].UserId != currentUserId {
		h.botMessagesService.SendRemovableBotMessage(
			ctx,
			&api.SlackPostMessageRequest{
				Text:     "Heads up <@" + currentUserId + ">! This bounty can only be awarded by <@" + messageBounties[0].UserId + ">.",
				Channel:  messageBounties[0].ChannelId,
				ThreadTs: messageId,
			},
			currentUserId,
			reaction,
		)
		return nil
	}

	// Ensure that they're not awarding it to themself.
	if targetUserId == currentUserId {
		h.botMessagesService.SendRemovableBotMessage(
			ctx,
			&api.SlackPostMessageRequest{
				Text:     "Heads up <@" + currentUserId + ">! You cannot award the bounty to yourself.",
				Channel:  messageBounties[0].ChannelId,
				ThreadTs: messageId,
			},
			currentUserId,
			reaction,
		)
		return nil
	}

	// Mark the bounty as awarded.
	messageBounties[0].Status = 2
	if _, err = h.messageBountiesRepo.Update(messageBounties[0]); err != nil {
		return errors.Wrapf(err, "failed to award bounty: %v", messageBounties[0].MessageId)
	}

	// Retrieve the channel account for the person we're going to award the bounty.
	channelAccount, err := h.getOrCreateChannelAccount(ctx, messageBounties[0].AwardedTo, messageBounties[0].ChannelId)
	if err != nil {
		return errors.Wrap(err, "failed to get or create channel account when awarded bounty")
	}

	// Award the bounty.
	if err = h.channelAccountsRepo.Award(channelAccount.Id, messageBounties[0].CurrentBounty); err != nil {
		return errors.Wrapf(err, "failed to award bounty to user after retrieving their account: %v", messageBounties[0].MessageId)
	}

	// Post reply to message that the bounty has been awarded tagging the awarder.
	h.apiClient.SendMessage(
		ctx,
		&api.SlackPostMessageRequest{
			Text:     "<@" + currentUserId + "> has awarded the bounty of " + fmt.Sprint(messageBounties[0].CurrentBounty) + " to <@" + messageBounties[0].AwardedTo + ">.",
			Channel:  messageBounties[0].ChannelId,
			ThreadTs: messageId,
		})

	return nil
}

func (h *SlackBotHandler) boostBounty(ctx context.Context, event *api.SlackReactionAddedEvent, boostAmount int) error {
	// To start, ensure that the user has an account they can use.
	channelAccount, err := h.getOrCreateChannelAccount(ctx, event.Event.User, event.Event.Item.Channel)
	if err != nil {
		return err
	}

	// Ensure that the user has a balance that is able to give this reward.
	if channelAccount.Balance < boostAmount {
		h.botMessagesService.SendRemovableBotMessage(
			ctx,
			&api.SlackPostMessageRequest{
				Text:     "Heads up <@" + event.Event.User + ">! Your balance isn't high enough to award :" + event.Event.Reaction + ":. Please remove your reaction to delete this message.",
				Channel:  event.Event.Item.Channel,
				ThreadTs: event.Event.Item.Ts,
			},
			event.Event.User,
			event.Event.Reaction,
		)

		// TODO: Watch reaction removed events so that we can remove this reply when a user removes the reaction.
		return fmt.Errorf("account %v has a balance of %v which is not enough to add a bounty of %v", channelAccount.Id, channelAccount.Balance, boostAmount)
	}

	// Check if there's an existing bounty for the message.
	messageBounties, _, err := h.messageBountiesRepo.List(
		&types.ListMessageBountiesFilter{
			MessageId: event.Event.Item.Ts,
			ChannelId: event.Event.Item.Channel,
		},
		1,
		"",
	)

	if err != nil {
		return errors.Wrap(err, "failed to check for an existing message bounty")
	}

	var messageBounty *types.MessageBounty

	// There is no existing bounty for this message.
	if len(messageBounties) == 0 {
		// Create a bounty for us to record awards etc.
		messageBounty, err = h.messageBountiesRepo.Create(
			&types.MessageBounty{
				MessageId:     event.Event.Item.Ts, // This is the id of the message.
				ChannelId:     event.Event.Item.Channel,
				UserId:        event.Event.ItemUser, // This is the id of the user who created the message, not the emote.
				CurrentBounty: 0,
				Status:        1,
				AwardedTo:     "",
			},
		)

		if err != nil {
			return errors.Wrap(err, "Failed to create an initial message bounty")
		}
	} else if len(messageBounties) > 1 {
		h.log.WithFields(logrus.Fields{
			"message_id": event.Event.Item.Ts,
			"channel_id": event.Event.Item.Channel,
		}).Error("Multiple matching message bounties found.")
		return errors.New("Multiple matching message bounties found, this should not occur.")
	} else {
		// Message bounty already exists, use it.
		messageBounty = messageBounties[0]
	}

	// Ensure that bounty hasn't already been awarded.
	if messageBounty.AwardedTo != "" {
		h.botMessagesService.SendRemovableBotMessage(
			ctx,
			&api.SlackPostMessageRequest{
				Text:     "Heads up <@" + event.Event.User + ">! This bounty has already been awarded and can no longer be boosted. Please remove your emote to delete this message.",
				Channel:  event.Event.Item.Channel,
				ThreadTs: event.Event.Item.Ts,
			},
			event.Event.User,
			event.Event.Reaction,
		)
		return nil
	}

	// Decrement the user's balance.
	if err = h.channelAccountsRepo.Spend(channelAccount.Id, boostAmount); err != nil {
		return errors.Wrapf(err, "Failed to spend for account balance: %v, %v", channelAccount.Id, boostAmount)
	}

	// Boost the message bounty.
	if err = h.messageBountiesRepo.BoostBounty(messageBounty.MessageId, boostAmount); err != nil {
		return errors.Wrapf(err, "Failed to boost bounty: %v, %v", messageBounty.MessageId, boostAmount)
	}

	// Acknowledge the bounty in chat.
	h.apiClient.SendMessage(
		ctx,
		&api.SlackPostMessageRequest{
			Text:     "<@" + event.Event.User + "> has boosted the bounty to " + fmt.Sprint(messageBounty.CurrentBounty+boostAmount) + ".",
			Channel:  event.Event.Item.Channel,
			ThreadTs: event.Event.Item.Ts,
		})

	return nil
}

// getOrCreateChannelAccount checks if the user has an account, if not create it.
func (h *SlackBotHandler) getOrCreateChannelAccount(ctx context.Context, user string, channel string) (*types.ChannelAccount, error) {
	// Retrieve the channel account if it exists.
	channelAccounts, _, err := h.channelAccountsRepo.List(
		&types.ListChannelAccountsFilter{
			UserId:    user,
			ChannelId: channel,
		},
		1,
		"",
		"",
	)

	if err != nil {
		return nil, errors.Wrap(err, "failed to check for an existing channel account")
	}

	// Channel account found, return to caller.
	if len(channelAccounts) == 1 {
		return channelAccounts[0], nil
	}

	if len(channelAccounts) > 1 {
		return nil, fmt.Errorf("multiple channel accounts found for %v %v, this should not occur", user, channel)
	}

	// No channel account, we'll create a new one and return that.
	channelAccount, err := h.channelAccountsRepo.Create(
		&types.ChannelAccount{
			UserId:         user,
			ChannelId:      channel,
			Balance:        h.config.DailyIncome,
			EarnedToday:    0,
			SpentToday:     0,
			EarnedThisWeek: 0,
			SpentThisWeek:  0,
			EarnedThisYear: 0,
			SpentThisYear:  0,
			EarnedAllTime:  0,
			SpentAllTime:   0,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create a new channel account for %v, %v", channelAccount.UserId, channelAccount.ChannelId)
	}

	// Check if this is their first account with the bot (none exist for other channels)
	channelAccounts, _, _ = h.channelAccountsRepo.List(
		&types.ListChannelAccountsFilter{
			UserId: user,
		},
		3,
		"",
		"",
	)

	// We check if it's only the new row and send a welcome message to the user if it is.
	if len(channelAccounts) == 1 {
		h.apiClient.SendMessage(
			ctx,
			&api.SlackPostMessageRequest{
				Text:    "Welcome to SlackBounties! You'll start off with 1 point and can earn more by completing bounties. Use */bountyemotes* to see the full list of emotes, */bountyme* to see your details and */bountydaily* to see the current leaderboard. Check out the following page for more info: " + h.config.DocumentationUrl,
				Channel: user,
			})
	}

	return channelAccount, nil
}

// getReactionTargetMessage retrieves the target message
func (h *SlackBotHandler) getReactionTargetMessage(ctx context.Context, event *api.SlackReactionAddedEvent) (*api.SlackMessage, error) {
	// Get the slack message via the API.
	slackMessages, err := h.apiClient.GetSlackMessage(
		ctx,
		&api.SlackConversationHistoryRequest{
			Channel:   event.Event.Item.Channel,
			Latest:    event.Event.Item.Ts,
			Limit:     1,
			Inclusive: true,
		},
	)
	if err != nil {
		return nil, err
	}

	if len(slackMessages) == 0 {
		return nil, fmt.Errorf("no errors found for %v", event.Event.Item.Ts)
	}

	if len(slackMessages) > 1 {
		return nil, fmt.Errorf("more than one message returned for %v", event.Event.Item.Ts)
	}

	return slackMessages[0], nil
}
