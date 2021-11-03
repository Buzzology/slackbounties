package types

import (
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ChannelAccount struct {
	Id             int
	UserId         string
	ChannelId      string
	Balance        int
	EarnedToday    int
	SpentToday     int
	EarnedThisWeek int
	SpentThisWeek  int
	EarnedThisYear int
	SpentThisYear  int
	EarnedAllTime  int
	SpentAllTime   int
	Created        timestamppb.Timestamp
	Updated        timestamppb.Timestamp
}

type ListChannelAccountsFilter struct {
	Id        int
	UserId    string
	ChannelId string
}
