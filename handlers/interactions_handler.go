package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/buzzology/slack_bot/service/api"
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

	// Define the view that we want to display for a modal.
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
			Submit: &api.SlackBlockSubmit{
				Type: "plain_text",
				Text: "Submit",
			},
			Blocks: []interface{}{
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
			},
		},
	}

	if _, err := h.apiClient.OpenView(ctx, slackViewsOpenRequest); err != nil {
		return errors.Wrap(err, "Failed to open the user select modal for awarding a bounty.")
	}

	return nil
}
