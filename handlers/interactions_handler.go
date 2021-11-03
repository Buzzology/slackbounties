package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/buzzology/slack_bot/service/api"
	"github.com/buzzology/slack_bot/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// InteractionsHandler handles and processes events received from slack.
func (h *SlackBotHandler) InteractionsHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	logrus.Println("Slack interaction received: ", r.RequestURI)

	ctx := context.Background()

	r.ParseForm()

	// Retrieve the payload.
	payload := r.FormValue("payload")
	if payload == "" {
		logrus.Error("No payload provided for slack interaction")
		return
	}

	defer r.Body.Close()

	// Initially use a generic unmarshall so that we can determine type.
	var payloadJson map[string]interface{}
	if err = json.Unmarshal([]byte(payload), &payloadJson); err != nil {
		logrus.Errorf("Failed to convert interaction to json: %v, %v", payload, err.Error())
		return
	}

	// Check the type of interaction.
	switch payloadJson["type"] {
	case "message_action":
		{
			// This is the initial request to provide a modal.
			if err = h.handleMessageAction(ctx, payloadJson, r.FormValue("payload")); err != nil {
				h.log.WithError(err).Error("Failed to process message_action")
				return
			}
		}
	case "view_submission":
		{
			// This is the request we receive when a submission is made from a modal.
			if err = h.handleViewSubmission(ctx, payloadJson, r.FormValue("payload")); err != nil {
				h.log.WithError(err).Error("Failed to process view_submission")
				return
			}
		}
	default:
		{
			logrus.Warningf("Unknown interaction type: %v", payloadJson["type"])
			return
		}
	}
}

// handleViewSubmission is used to handle modal submissions.
func (h *SlackBotHandler) handleViewSubmission(
	ctx context.Context,
	payloadJson map[string]interface{},
	rawPayload string,
) error {
	// Decode to a slack interaction.
	interaction := &api.SlackInteraction{}
	if err := json.Unmarshal([]byte(rawPayload), &interaction); err != nil {
		logrus.Errorf("Failed to ummarshall interaction payload %v, %v", rawPayload, err.Error())
		return nil
	}

	if err := h.awardBounty(ctx, interaction.View.PrivateMetadata, getTargetBountyUserFromInteraction(interaction), interaction.User.Id, ""); err != nil {
		h.log.WithError(err).WithFields(logrus.Fields{
			"message_id":         interaction.View.PrivateMetadata,
			"target_bounty_user": getTargetBountyUserFromInteraction(interaction),
			"user_id":            interaction.User.Id,
		}).Error("Unable to award bounty via interaction.")
		return errors.New("Unable to award bounty via interaction.")
	}

	return nil
}

// TODO: Find a better way to do this...
func getTargetBountyUserFromInteraction(interaction *api.SlackInteraction) string {
	targetBountyUser := interaction.View.State.Values["award-bounty-user-id"].(map[string]interface{})
	test3 := targetBountyUser["award-bounty-user"].(map[string]interface{})
	test4 := test3["selected_user"]

	return test4.(string)
}

// handleMessageAction is used to handle message actions (requests to show a modal etc).
func (h *SlackBotHandler) handleMessageAction(
	ctx context.Context,
	payloadJson map[string]interface{},
	rawPayload string,
) error {
	// Decode to a slack interaction.
	interaction := &api.SlackInteraction{}
	if err := json.Unmarshal([]byte(rawPayload), &interaction); err != nil {
		logrus.Errorf("Failed to ummarshall interaction payload %v, %v", rawPayload, err.Error())
		return nil
	}

	// Define the generic view that we will display for the modal.
	slackViewsOpenRequest := &api.SlackViewsOpenRequest{
		TriggerId: interaction.TriggerId,
		View: &api.SlackView{
			Type:            "modal",
			CallbackId:      interaction.CallbackId,
			PrivateMetadata: interaction.MessageTs,
			Title: &api.SlackBlock{
				Type: "plain_text",
				Text: "Award a Bounty",
			},
		},
	}

	// Retrieve the message bounty.
	messageBounties, _, err := h.messageBountiesRepo.List(
		&types.ListMessageBountiesFilter{
			MessageId: interaction.MessageTs,
			ChannelId: interaction.Channel.Id,
		},
		1,
		"",
	)
	if err != nil {
		h.log.WithFields(logrus.Fields{
			"message_id": interaction.MessageTs,
			"channel_id": interaction.Channel.Id,
		}).WithError(err).Error("Failed to retrieve message bounties")
		return errors.New("Failed to retrieve message bounties")
	}

	// Check if there is a message bounty.
	if len(messageBounties) == 0 {
		slackViewsOpenRequest.View.Blocks = []interface{}{
			&api.SlackBlock{
				Type: "section",
				Text: &api.SlackBlockText{
					Type: "mrkdwn",
					Text: "There is currently no bounty on this message to award.",
				},
			},
		}

		if _, err := h.apiClient.OpenView(ctx, slackViewsOpenRequest); err != nil {
			return errors.Wrap(err, "Failed to open the no bounty modal when awarding a bounty.")
		}

		return nil
	}

	var messageBounty = messageBounties[0]

	// Ensure that this user is the message owner.
	if messageBounty.UserId != interaction.User.Id {
		slackViewsOpenRequest.View.Blocks = []interface{}{
			&api.SlackBlock{
				Type: "section",
				Text: &api.SlackBlockText{
					Type: "mrkdwn",
					Text: "Only the message creator can award the bounty.",
				},
			},
		}

		if _, err := h.apiClient.OpenView(ctx, slackViewsOpenRequest); err != nil {
			return errors.Wrap(err, "Failed to open the not message creator modal when awarding a bounty.")
		}

		return nil
	}

	// Ensure that the bounty hasn't already been awarded.
	if messageBounty.Status > 1 {
		slackViewsOpenRequest.View.Blocks = []interface{}{
			&api.SlackBlock{
				Type: "section",
				Text: &api.SlackBlockText{
					Type: "mrkdwn",
					Text: "This bounty has already been awarded.",
				},
			},
		}

		if _, err := h.apiClient.OpenView(ctx, slackViewsOpenRequest); err != nil {
			return errors.Wrap(err, "Failed to open the already awarded modal when awarding a bounty.")
		}

		return nil
	}

	// It should be fine to award the bounty if all previous validation is complete.
	slackViewsOpenRequest.View.Submit = &api.SlackBlockSubmit{
		Type: "plain_text",
		Text: "Submit",
	}
	slackViewsOpenRequest.View.Blocks = []interface{}{
		&api.SlackBlock{
			Type:    "input",
			BlockId: "award-bounty-user-id",
			Element: &api.SlackBlockAccessory{
				ActionId: "award-bounty-user",
				Type:     "users_select",
				Placeholder: &api.SlackBlock{
					Type: "plain_text",
					Text: "Select a user",
				},
			},
			Label: &api.SlackBlockLabel{
				Type:  "plain_text",
				Text:  "Pick a user to award the bounty to.",
				Emoji: false,
			},
		},
	}

	if _, err := h.apiClient.OpenView(ctx, slackViewsOpenRequest); err != nil {
		return errors.Wrap(err, "Failed to open the user select modal for awarding a bounty.")
	}

	return nil
}
