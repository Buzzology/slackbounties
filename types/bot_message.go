package types

import "google.golang.org/protobuf/types/known/timestamppb"

type BotMessage struct {
	Id            int64                 `json:"id"`
	MessageId     string                `json:"message_id"`
	SentMessageId string                `json:"sent_message_id"`
	ChannelId     string                `json:"channel_id"`
	Reaction      string                `json:"reaction"`
	Status        int                   `json:"status"`
	TargetUserId  string                `json:"target_user_id"`
	Created       timestamppb.Timestamp `json:"created"`
	Updated       timestamppb.Timestamp `json:"updated"`
}

type ListBotMessagesFilter struct {
	Id            string
	MessageId     string
	SentMessageId string
	UserId        string
	ChannelId     string
	Status        int
	Reaction      string
}
