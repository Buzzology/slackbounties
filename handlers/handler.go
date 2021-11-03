package handlers

import (
	"github.com/buzzology/slack_bot/db"
	"github.com/buzzology/slack_bot/service"
	"github.com/buzzology/slack_bot/service/api"
	"github.com/sirupsen/logrus"
)

type SlackBotHandler struct {
	config                 *service.Config
	apiClient              *api.SlackApiClient
	log                    logrus.FieldLogger
	channelAccountsRepo    db.ChannelAccountsRepo
	messageBountiesRepo    db.MessageBountiesRepo
	channelAccountsService *service.ChannelAccountsService
	botMessagesService     *service.BotMessagesService
	botMessagesRepo        db.BotMessagesRepo
}

func NewSlackBotHandler(
	config *service.Config,
	log logrus.FieldLogger,
	apiClient *api.SlackApiClient,
	channelAccountsRepo db.ChannelAccountsRepo,
	messageBountiesRepo db.MessageBountiesRepo,
	channelAccountsService *service.ChannelAccountsService,
	botMessagesService *service.BotMessagesService,
	botMessagesRepo db.BotMessagesRepo,
) *SlackBotHandler {
	return &SlackBotHandler{
		config:                 config,
		apiClient:              apiClient,
		log:                    log,
		channelAccountsRepo:    channelAccountsRepo,
		messageBountiesRepo:    messageBountiesRepo,
		channelAccountsService: channelAccountsService,
		botMessagesService:     botMessagesService,
		botMessagesRepo:        botMessagesRepo,
	}
}
