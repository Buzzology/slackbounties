package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type SlackApi interface {
	DeleteSlackMessage(ctx context.Context, request *SlackDeleteMessageRequest)
	GetSlackMessage(ctx context.Context, request *SlackConversationHistoryRequest)
	SendMessage(ctx context.Context, request *SlackConversationHistoryRequest) (*SlackPostMessageResponse, error)
	OpenView(ctx context.Context, request *SlackViewsOpenRequest) (*SlackOpenViewResponse, error)
}

type SlackApiClient struct {
	config *ApiConfig
	log    *logrus.Logger
}

// NewSlackApiClient return a new Slack API client.
func NewSlackApiClient(
	config *ApiConfig,
	log *logrus.Logger,
) *SlackApiClient {
	return &SlackApiClient{
		config: config,
		log:    log,
	}
}

// GetSlackMessage retrieves a specific slack message via the slack API. Docs: https://api.slack.com/messaging/retrieving#individual_messages
func (c *SlackApiClient) GetSlackMessage(
	ctx context.Context,
	request *SlackConversationHistoryRequest,
) ([]*SlackMessage, error) {
	var (
		err error
		req *http.Request
	)

	requestUrl, err := url.Parse(c.config.Endpoint + "/conversations.history")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create message get url")
	}

	q := requestUrl.Query()
	q.Set("channel", request.Channel)
	q.Set("latest", request.Latest)
	q.Set("limit", strconv.Itoa(request.Limit))
	q.Set("inclusive", strconv.FormatBool(request.Inclusive))

	requestUrl.RawQuery = q.Encode()
	if req, err = http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		requestUrl.String(),
		nil,
	); err != nil {
		return nil, errors.Wrap(err, "error building SlackConversationHistory request")
	}

	req.Header.Set("Authorization", "Bearer "+c.config.Token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if err = c.handleHTTPResponse(resp); err != nil {
		return nil, errors.Wrap(err, "unsuccessful response to SlackConversationHistory request")
	}

	defer resp.Body.Close()

	var slackConversationHistoryResponse SlackConversationHistoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&slackConversationHistoryResponse); err != nil {
		return nil, errors.Wrap(err, "error decoding SlackConversationHistory response")
	}

	if !slackConversationHistoryResponse.Ok {
		c.log.WithFields(logrus.Fields{
			"error":          slackConversationHistoryResponse.Error,
			"messages_count": len(slackConversationHistoryResponse.Messages),
		}).Error("slack api message failed")
		return nil, errors.Errorf("failed to retrieve slack message: %v", slackConversationHistoryResponse.Error)
	}

	if len(slackConversationHistoryResponse.Messages) == 0 {
		return []*SlackMessage{}, nil
	}

	return slackConversationHistoryResponse.Messages, nil
}

// handleApiErrorResponse is used to handle slack api http errors.
func (c *SlackApiClient) handleHTTPResponse(resp *http.Response) error {
	if resp == nil {
		return errors.New("no response provided")
	}

	if resp.StatusCode > 299 {
		defer resp.Body.Close()
		bodyBytes, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			c.log.WithFields(logrus.Fields{
				"status_code": resp.StatusCode,
				"error":       err.Error(),
			}).Error("Failed to read slack api error response")
		} else {
			c.log.WithFields(logrus.Fields{
				"status_code": resp.StatusCode,
				"body":        string(bodyBytes),
			}).Error("Slack api request failed")
		}

		return errors.Errorf("an http error has occurred: %v", resp.StatusCode)
	}

	return nil
}

/* SendMessage sends a message from the bot to a user or channel.
- To create a private message pass user id as channel
- To post in a channel pass the channel id as channel in request
- To reply to an existing message pass the message's channel (user or channel) and the message id as thread_ts
*/
func (c *SlackApiClient) SendMessage(
	ctx context.Context,
	request *SlackPostMessageRequest,
) (*SlackPostMessageResponse, error) {
	var (
		err     error
		httpReq *http.Request
	)

	requestUrl, err := url.Parse(c.config.Endpoint + "/chat.postMessage")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create post message url")
	}

	buffer := new(bytes.Buffer)
	if err := json.NewEncoder(buffer).Encode(request); err != nil {
		return nil, err
	}

	if httpReq, err = http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		requestUrl.String(),
		buffer,
	); err != nil {
		return nil, errors.Wrap(err, "error building SlackPostMessageRequest request")
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.config.Token)
	httpReq.Header.Set("Content-Type", "application/json; charset=utf-8")

	dump, err := httputil.DumpRequestOut(httpReq, true)
	fmt.Printf("%q", dump)

	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	if err = c.handleHTTPResponse(resp); err != nil {
		return nil, errors.Wrap(err, "unsuccessful response to SlackPostMessage request")
	}

	defer resp.Body.Close()

	var slackApiResponse SlackPostMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&slackApiResponse); err != nil {
		return nil, errors.Wrap(err, "error decoding SlackPostMessage response")
	}

	if !slackApiResponse.Ok {
		c.log.WithFields(logrus.Fields{
			"error": slackApiResponse.Error,
		}).Error("slack api message failed")
		return nil, errors.Errorf("failed to retrieve slack message: %v", slackApiResponse.Error)
	}

	return &slackApiResponse, nil
}

// DeleteSlackMessage is used to delete an existing message.
func (c *SlackApiClient) DeleteSlackMessage(
	ctx context.Context,
	request *SlackDeleteMessageRequest,
) (*SlackDeleteMessageResponse, error) {
	var (
		err     error
		httpReq *http.Request
	)

	requestUrl, err := url.Parse(c.config.Endpoint + "/chat.delete")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create delete message url")
	}

	buffer := new(bytes.Buffer)
	if err := json.NewEncoder(buffer).Encode(request); err != nil {
		return nil, err
	}

	if httpReq, err = http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		requestUrl.String(),
		buffer,
	); err != nil {
		return nil, errors.Wrap(err, "error building SlackDeleteMessage request")
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.config.Token)
	httpReq.Header.Set("Content-Type", "application/json; charset=utf-8")

	// TODO: Remove this when testing is finished.
	dump, err := httputil.DumpRequestOut(httpReq, true)
	fmt.Printf("%q", dump)

	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	if err = c.handleHTTPResponse(resp); err != nil {
		return nil, errors.Wrap(err, "unsuccessful response to SlackDeleteMessage request")
	}

	defer resp.Body.Close()

	var slackApiResponse SlackDeleteMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&slackApiResponse); err != nil {
		return nil, errors.Wrap(err, "error decoding SlackDeleteMessage response")
	}

	if !slackApiResponse.Ok {
		c.log.WithFields(logrus.Fields{
			"error": slackApiResponse.Error,
		}).Error("slack api message failed")
		return nil, errors.Errorf("failed to delete slack message: %v", slackApiResponse.Error)
	}

	return &slackApiResponse, nil
}

// OpenView instructs the slack client to open a modal.
func (c *SlackApiClient) OpenView(
	ctx context.Context,
	request *SlackViewsOpenRequest,
) (*SlackOpenViewResponse, error) {
	var (
		err     error
		httpReq *http.Request
	)

	requestUrl, err := url.Parse(c.config.Endpoint + "/views.open")
	if err != nil {
		return nil, err
	}

	buffer := new(bytes.Buffer)
	if err := json.NewEncoder(buffer).Encode(request); err != nil {
		return nil, err
	}

	if httpReq, err = http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		requestUrl.String(),
		buffer,
	); err != nil {
		return nil, errors.Wrap(err, "error building request to OpenView request")
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.config.Token)
	httpReq.Header.Set("Content-Type", "application/json; charset=utf-8")

	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	if err = c.handleHTTPResponse(resp); err != nil {
		return nil, errors.Wrap(err, "unsuccessful response to OpenView request")
	}

	defer resp.Body.Close()

	var slackApiResponse SlackOpenViewResponse
	if err := json.NewDecoder(resp.Body).Decode(&slackApiResponse); err != nil {
		return nil, errors.Wrap(err, "error decoding OpenView response")
	}

	if !slackApiResponse.Ok {
		c.log.WithFields(logrus.Fields{
			"error": slackApiResponse.Error,
		}).Error("slack OpenView api message failed")
		return nil, errors.Errorf("failed to OpenView: %v", slackApiResponse.Error)
	}

	return &slackApiResponse, nil
}
