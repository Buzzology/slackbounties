package types

import (
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MessageBounty struct {
	// MessageId is the ts value of the slack message.
	MessageId string
	// ChannelId is the channel to which the bounty belongs.
	ChannelId string
	// UserId should be the pull request (message's) creator.
	UserId string
	// CurrentBounty is the amount the will/was awarded to the reviewer.
	CurrentBounty int
	// Status is either open/closed.
	Status int
	// AwardedTo is the user that received the bounty.
	AwardedTo string
	// Created is when the message bounty was initially created.
	Created timestamppb.Timestamp
	// Updated is when the message bounty was updated.
	Updated timestamppb.Timestamp
}

type ListMessageBountiesFilter struct {
	MessageId string
	UserId    string
	ChannelId string
}
