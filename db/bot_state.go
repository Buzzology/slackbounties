package db

import (
	"database/sql"
	"time"

	"github.com/buzzology/slack_bot/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type BotStateRepo interface {
	// Init will initialise our bot state repo.
	Init() error

	// Get will retrieve our bot state.
	Get() (*types.BotState, error)

	// Tickover will perform any tickovers if required.
	Tickover(decay int, income int) (*types.BotState, error)

	// Create a new bot state.
	Create() (*types.BotState, error)

	// Update the bot's state.
	Update(*types.BotState) (*types.BotState, error)

	// ApplyIncomeAndDecay will decrement all users incomes by the decay and increment by the income.
	ApplyIncomeAndDecay(decayToApply int, incomeToApply int) error
}

type botStateRepo struct {
	db  *sql.DB
	log *logrus.Logger
}

func NewBotStateRepo(
	db *sql.DB,
	log *logrus.Logger,
) BotStateRepo {
	return &botStateRepo{
		db:  db,
		log: log,
	}
}

// Init initialises the bot state repo.
func (r *botStateRepo) Init() error {
	return nil
}

// Update the bot's state.
func (r *botStateRepo) Update(botState *types.BotState) (*types.BotState, error) {
	var _, err = r.db.Exec(
		getBotStateQueries()[botStateUpdate],
		botState.DayTickover.AsTime(),
		botState.WeekTickover.AsTime(),
		botState.MonthTickover.AsTime(),
		botState.YearTickover.AsTime(),
		botState.Id,
	)

	if err != nil {
		return nil, err
	}

	// TODO: Check rows affected, if 0 error.

	return r.Get()
}

// Create will create a new bot state.
func (r *botStateRepo) Create() (*types.BotState, error) {
	_, err := r.db.Exec(
		getBotStateQueries()[botStateCreate],
	)

	if err != nil {
		return nil, err
	}

	return r.Get()
}

// Tickover will handle day updates etc.
func (r *botStateRepo) Tickover(decay int, income int) (*types.BotState, error) {
	// TODO: Wrap all of this in a transaction.
	botState, err := r.Get()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to retrieve bot state while performing tickover.")
	}

	now := time.Now()

	// Reset daily if required.
	if botState.DayTickover.AsTime().Before(now) {
		_, err := r.db.Exec(
			getChannelAccountQueries()[channelAccountResetDaily],
		)

		if err != nil {
			return nil, err
		}

		botState.DayTickover = *timestamppb.New(botState.DayTickover.AsTime().Add(time.Hour * 24))

		// We apply decay and income daily as well.
		if err = r.ApplyIncomeAndDecay(decay, income); err != nil {
			return nil, err
		}
	}

	// Reset weekly if required.
	if botState.WeekTickover.AsTime().Before(now) {
		_, err := r.db.Exec(
			getChannelAccountQueries()[channelAccountResetWeekly],
		)

		if err != nil {
			return nil, err
		}

		botState.WeekTickover = *timestamppb.New(botState.DayTickover.AsTime().Add(time.Hour * 730))
	}

	// Reset yearly if required.
	if botState.YearTickover.AsTime().Before(now) {
		_, err := r.db.Exec(
			getChannelAccountQueries()[channelAccountResetYearly],
		)

		if err != nil {
			return nil, err
		}

		botState.YearTickover = *timestamppb.New(botState.YearTickover.AsTime().Add(time.Hour * 8760))
	}

	return r.Update(botState)
}

// Get will retrieve the current bot state.
func (r *botStateRepo) Get() (*types.BotState, error) {
	var args []interface{}
	var query = getBotStateQueries()[botStateGet]

	// Execute the query
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	// Parse rows as orders
	var botStates []*types.BotState
	botStates, err = r.scanBotStates(rows)
	if err != nil {
		return nil, err
	}

	// No results
	if len(botStates) == 0 {
		return nil, nil
	}

	// Return the results along with the id of the last account as a next page token
	return botStates[0], nil
}

// scanBotState populates a slice of bot states from the db.
func (r *botStateRepo) scanBotStates(rows *sql.Rows) ([]*types.BotState, error) {

	var res []*types.BotState

	for rows.Next() {

		var (
			botState      types.BotState
			created       time.Time
			dayTickover   time.Time
			weekTickover  time.Time
			monthTickover time.Time
			yearTickover  time.Time
		)

		// Populate the row
		if err := rows.Scan(
			&botState.Id,
			&created,
			&dayTickover,
			&weekTickover,
			&monthTickover,
			&yearTickover,
		); err != nil {

			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, err
		}

		// Assign timestamps
		botState.Created = *timestamppb.New(created)
		botState.DayTickover = *timestamppb.New(dayTickover)
		botState.WeekTickover = *timestamppb.New(weekTickover)
		botState.MonthTickover = *timestamppb.New(monthTickover)
		botState.YearTickover = *timestamppb.New(yearTickover)

		res = append(res, &botState)
	}

	return res, nil
}

// ApplyIncomeAndDecay will apply income and decay to all channel
func (r *botStateRepo) ApplyIncomeAndDecay(
	decayToApply int,
	incomeToApply int,
) error {
	var query = getChannelAccountQueries()[channelAccountApplyIncomeAndDecay]

	// Execute the query
	_, err := r.db.Exec(query, decayToApply, incomeToApply, decayToApply, incomeToApply)
	return err
}
