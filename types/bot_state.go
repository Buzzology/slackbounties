package types

import (
	"google.golang.org/protobuf/types/known/timestamppb"
)

type BotState struct {
	Id            int
	Created       timestamppb.Timestamp
	DayTickover   timestamppb.Timestamp
	WeekTickover  timestamppb.Timestamp
	MonthTickover timestamppb.Timestamp
	YearTickover  timestamppb.Timestamp
}
