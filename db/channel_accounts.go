package db

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/buzzology/slack_bot/types"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ChannelAccountsRepo interface {
	// Init will initialise our account transactions repo.
	Init() error

	// List will return a collection of account transactions.
	List(filter *types.ListChannelAccountsFilter, pageSize int, pageToken string, order string) ([]*types.ChannelAccount, string, error)

	// Create will create a new channel account for tracking balances etc.
	Create(channelAccount *types.ChannelAccount) (*types.ChannelAccount, error)

	// Update will update an existing channel account.
	Update(channelAccount *types.ChannelAccount) (*types.ChannelAccount, error)

	// Spend will update a channel account to reflect a new spend amount.
	Spend(id int, amount int) error

	// Award will update a channel account to reflect a new award amount.
	Award(id int, amount int) error

	// ActiveTodayCount will count the number of accounts in the channel that are active today.
	ActiveTodayCount(channelId string) (int, error)

	// ActiveThisWeekCount will count the number of accounts in the channel that have been active this week.
	ActiveThisWeekCount(channelId string) (int, error)

	// ActiveThisYearCount will count the number of accounts in the channel that have been active this year.
	ActiveThisYearCount(channelId string) (int, error)

	// ActiveAllTimeCount will count the number of accounts in the channel that have been active all time.
	ActiveAllTimeCount(channelId string) (int, error)

	// LeadersToday will count the number of leading accounts for today.
	LeadersToday(channelId string, percentageToShow int, maxToShow int) ([]*types.ChannelAccount, error)

	// LeadersThisWeek will count the number of leading accounts for today.
	LeadersThisWeek(channelId string, percentageToShow int, maxToShow int) ([]*types.ChannelAccount, error)

	// LeadersThisYear will count the number of leading accounts for today.
	LeadersThisYear(channelId string, percentageToShow int, maxToShow int) ([]*types.ChannelAccount, error)

	// LeadersAllTime will count the number of leading accounts for today.
	LeadersAllTime(channelId string, percentageToShow int, maxToShow int) ([]*types.ChannelAccount, error)

	// ResetDaily will reset daily tracking for all channel accounts.
	ResetDaily() error

	// ResetWeekly will reset weekly tracking for all channel accounts.
	ResetWeekly() error

	// ResetYearly will reset yearly tracking for all channel accounts.
	ResetYearly() error

	// DistinctChannels will return a list of all distinct channels.
	DistinctChannels() ([]string, error)
}

type channelAccountsRepo struct {
	db  *sql.DB
	log *logrus.Logger
}

func NewChannelAccountsRepo(
	db *sql.DB,
	log *logrus.Logger,
) ChannelAccountsRepo {
	return &channelAccountsRepo{
		db:  db,
		log: log,
	}
}

// Init initialises the channel accounts repo.
func (r *channelAccountsRepo) Init() error {
	return nil
}

// List will retrieve and list channel accounts matching the provided criteria.
func (r *channelAccountsRepo) List(
	filter *types.ListChannelAccountsFilter,
	pageSize int,
	pageToken string,
	order string,
) ([]*types.ChannelAccount, string, error) {
	var args []interface{}
	var query = getChannelAccountQueries()[channelAccountsList]

	// Prepare query
	query, args = r.applyFilter(query, filter, pageSize, pageToken, order)

	// Execute the query
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, "", err
	}

	// Parse rows as orders
	var accounts []*types.ChannelAccount
	accounts, err = r.scanChannelAccounts(rows)
	if err != nil {
		return nil, "", err
	}

	// No results
	if len(accounts) == 0 {
		return accounts, "", nil
	}

	// Return the results along with the id of the last account as a next page token
	return accounts, fmt.Sprint(accounts[len(accounts)-1].Id), nil
}

// Create will create a new channel account.
func (r *channelAccountsRepo) Create(
	channelAccount *types.ChannelAccount,
) (*types.ChannelAccount, error) {
	var res, err = r.db.Exec(
		getChannelAccountQueries()[channelAccountCreate],
		channelAccount.UserId,
		channelAccount.ChannelId,
		channelAccount.Balance,
	)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Retrieve the row
	rows, _, err := r.List(
		&types.ListChannelAccountsFilter{
			Id: int(id),
		},
		1,
		"",
		"",
	)

	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("failed to retrieve new channel_account: %v", id)
	}

	return rows[0], nil
}

// Update will modify an existing channel account.
func (r *channelAccountsRepo) Update(
	channelAccount *types.ChannelAccount,
) (*types.ChannelAccount, error) {
	var _, err = r.db.Exec(
		getChannelAccountQueries()[channelAccountUpdate],
		channelAccount.Balance,
		channelAccount.EarnedToday,
		channelAccount.SpentToday,
		channelAccount.EarnedThisWeek,
		channelAccount.SpentThisWeek,
		channelAccount.EarnedThisYear,
		channelAccount.SpentThisYear,
		channelAccount.EarnedAllTime,
		channelAccount.SpentAllTime,
		channelAccount.Id,
	)
	if err != nil {
		return nil, err
	}

	// TODO: Check rows affected, if 0 error.

	// Retrieve the row
	rows, _, err := r.List(
		&types.ListChannelAccountsFilter{
			Id: int(channelAccount.Id),
		},
		1,
		"",
		"",
	)

	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("failed to retrieve the updated channel_account: %v", channelAccount.Id)
	}

	return rows[0], nil
}

// Spend will update a channel account to reflect a new spend amount.
func (r *channelAccountsRepo) Spend(
	id int,
	amount int,
) error {
	var _, err = r.db.Exec(
		getChannelAccountQueries()[channelAccountSpend],
		amount,
		amount,
		amount,
		amount,
		amount,
		id,
	)

	return err
}

// Award will update a channel account to reflect a new award amount.
func (r *channelAccountsRepo) Award(
	id int,
	amount int,
) error {
	var _, err = r.db.Exec(
		getChannelAccountQueries()[channelAccountAward],
		amount,
		amount,
		amount,
		amount,
		amount,
		id,
	)

	return err
}

// ResetDaily will reset daily tracking for all channel accounts.
func (r *channelAccountsRepo) ResetDaily() error {
	var _, err = r.db.Exec(getChannelAccountQueries()[channelAccountResetDaily])
	return err
}

// ResetWeekly will reset weekly tracking for all channel accounts.
func (r *channelAccountsRepo) ResetWeekly() error {
	var _, err = r.db.Exec(getChannelAccountQueries()[channelAccountResetWeekly])
	return err
}

// ResetYearly will reset yearly tracking for all channel accounts.
func (r *channelAccountsRepo) ResetYearly() error {
	var _, err = r.db.Exec(getChannelAccountQueries()[channelAccountResetYearly])
	return err
}

// LeadersToday will count the number of leading accounts for today.
func (r *channelAccountsRepo) LeadersToday(channelId string, percentageToShow int, maxToShow int) ([]*types.ChannelAccount, error) {
	// Get the count of active users
	activeCount, err := r.ActiveTodayCount(channelId)
	if err != nil {
		return nil, err
	}

	return r.getLeaders(" earned_today DESC", activeCount, channelId, percentageToShow, maxToShow)
}

// LeadersThisWeek will count the number of leading accounts for this week.
func (r *channelAccountsRepo) LeadersThisWeek(channelId string, percentageToShow int, maxToShow int) ([]*types.ChannelAccount, error) {
	// Get the count of active users
	activeCount, err := r.ActiveThisWeekCount(channelId)
	if err != nil {
		return nil, err
	}

	return r.getLeaders(" earned_this_week DESC", activeCount, channelId, percentageToShow, maxToShow)
}

// LeadersThisYear will count the number of leading accounts for this week.
func (r *channelAccountsRepo) LeadersThisYear(channelId string, percentageToShow int, maxToShow int) ([]*types.ChannelAccount, error) {
	// Get the count of active users
	activeCount, err := r.ActiveThisYearCount(channelId)
	if err != nil {
		return nil, err
	}

	return r.getLeaders(" earned_this_year DESC", activeCount, channelId, percentageToShow, maxToShow)
}

// LeadersAllTime will count the number of leading accounts for this week.
func (r *channelAccountsRepo) LeadersAllTime(channelId string, percentageToShow int, maxToShow int) ([]*types.ChannelAccount, error) {
	// Get the count of active users
	activeCount, err := r.ActiveAllTimeCount(channelId)
	if err != nil {
		return nil, err
	}

	return r.getLeaders(" earned_all_time DESC", activeCount, channelId, percentageToShow, maxToShow)
}

// getLeaders is used as a generic mechanism to faciliate retrieving leaders (the leaderboard service calls).
func (r *channelAccountsRepo) getLeaders(orderStatement string, activeCount int, channelId string, percentageToShow int, maxToShow int) ([]*types.ChannelAccount, error) {
	numberToShow := calculateNumberOfLeadersToShow(activeCount, percentageToShow, maxToShow)

	// Retrieve the channel accounts we want to show.
	channelAccounts, _, err := r.List(
		&types.ListChannelAccountsFilter{
			Id:        0,
			UserId:    "",
			ChannelId: channelId,
		},
		numberToShow,
		"",
		orderStatement,
	)

	return channelAccounts, err
}

// ActiveTodayCount will count the number of accounts in the channel that are active today.
func (r *channelAccountsRepo) ActiveTodayCount(channelId string) (int, error) {
	return r.count(channelAccountActiveTodayCount, channelId)
}

// ActiveThisWeekCount will count the number of accounts in the channel that have been active this week.
func (r *channelAccountsRepo) ActiveThisWeekCount(channelId string) (int, error) {
	return r.count(channelAccountActiveThisWeekCount, channelId)
}

// ActiveThisYearCount will count the number of accounts in the channel that have been active this year.
func (r *channelAccountsRepo) ActiveThisYearCount(channelId string) (int, error) {
	return r.count(channelAccountActiveThisYearCount, channelId)
}

// ActiveAllTimeCount will count the number of accounts in the channel that have been active all time.
func (r *channelAccountsRepo) ActiveAllTimeCount(channelId string) (int, error) {
	return r.count(channelAccountActiveAllTimeCount, channelId)
}

// DistinctChannels will return a list of all distinct channels. Used for leaderboards etc.
func (r *channelAccountsRepo) DistinctChannels() (channelIds []string, err error) {
	rows, err := r.db.Query(getChannelAccountQueries()[channelAccountsDistinctChannels])
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var channelId string
		if err = rows.Scan(&channelId); err != nil {
			return nil, err
		}

		channelIds = append(channelIds, channelId)
	}

	return channelIds, nil
}

// Count will use the provided query name to count rows.
func (r *channelAccountsRepo) count(
	countQueryToUse string,
	channelId string,
) (count int, err error) {
	var query = getChannelAccountQueries()[countQueryToUse]

	// Execute the query
	if err = r.db.QueryRow(query, channelId).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (r *channelAccountsRepo) applyFilter(
	query string,
	filter *types.ListChannelAccountsFilter,
	pageSize int,
	pageToken string,
	order string,
) (string, []interface{}) {
	var (
		clauses []string
		args    []interface{}
	)

	if filter == nil {
		return query, args
	}

	// Filter by id if provided
	if filter.Id != 0 {
		clauses = append(clauses, "id = ?")
		args = append(args, filter.Id)
	}

	// Filter by user id if provided
	if filter.UserId != "" {
		clauses = append(clauses, "user_id = ?")
		args = append(args, filter.UserId)
	}

	// Filter by channel id if provided
	if filter.ChannelId != "" {
		clauses = append(clauses, "channel_id = ?")
		args = append(args, filter.ChannelId)
	}

	if len(clauses) != 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}

	// Validate page size
	if pageSize > 100 || pageSize <= 0 {
		pageSize = 100
	}

	query = r.applyOrder(query, order)

	// Limit page size
	// NOTE: We will likely want to keep the limit and remove the offset. Instead we should dynamically filter using
	//       a where clause based on the sort order. E.g. if sorting by id `where id > page_token ORDER BY id`
	pageTokenI, err := strconv.Atoi(pageToken)
	if err == nil {
		query += fmt.Sprintf(" LIMIT %v, %v", pageTokenI, pageSize)
	} else {
		r.log.Warningf("invalid page token provided: %v", pageToken)
		query += fmt.Sprintf(" LIMIT 0, %v", pageSize)
	}

	return query, args
}

// scanChannelAccounts populates a slice of structs from db rows
func (r *channelAccountsRepo) scanChannelAccounts(rows *sql.Rows) ([]*types.ChannelAccount, error) {
	var res []*types.ChannelAccount

	for rows.Next() {

		var (
			channelAccount types.ChannelAccount
			created        time.Time
			updated        time.Time
		)

		// Populate the row
		if err := rows.Scan(
			&channelAccount.Id,
			&channelAccount.UserId,
			&channelAccount.ChannelId,
			&channelAccount.Balance,
			&channelAccount.EarnedToday,
			&channelAccount.SpentToday,
			&channelAccount.EarnedThisWeek,
			&channelAccount.SpentThisWeek,
			&channelAccount.EarnedThisYear,
			&channelAccount.SpentThisYear,
			&channelAccount.EarnedAllTime,
			&channelAccount.SpentAllTime,
			&created,
			&updated,
		); err != nil {

			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, err
		}

		// Assign timestamps
		channelAccount.Created = *timestamppb.New(created)
		channelAccount.Updated = *timestamppb.New(updated)

		res = append(res, &channelAccount)
	}

	return res, nil
}

func (r *channelAccountsRepo) applyOrder(
	query string,
	order string,
) string {
	if order == "" {
		return query + " ORDER BY created DESC"
	}

	return query + " ORDER BY " + order
}

// calculateNumberOfLeadersToShow is used to ensure we only show a portion of users on the leaderboard.
func calculateNumberOfLeadersToShow(activeUsersCount int, percentageToShow int, maxToShow int) int {
	if activeUsersCount == 0 {
		return 0
	}

	if activeUsersCount == 1 {
		return 1
	}

	calcedPercentage := activeUsersCount / 100 * percentageToShow
	if calcedPercentage > maxToShow {
		return maxToShow
	} else if calcedPercentage == 0 {
		return 1
	}

	return calcedPercentage
}
