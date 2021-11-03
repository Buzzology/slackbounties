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

type MessageBountiesRepo interface {
	// Init will initialise our message bounties repo.
	Init() error

	// List will return a collection of message bounties.
	List(filter *types.ListMessageBountiesFilter, pageSize int, pageToken string) ([]*types.MessageBounty, string, error)

	// Create will create a new message bounty.
	Create(messageBounty *types.MessageBounty) (*types.MessageBounty, error)

	// Update will update an existing message bounty.
	Update(messageBounty *types.MessageBounty) (*types.MessageBounty, error)

	// BoostBounty will boost an existing bounty.
	BoostBounty(messageId string, boostAmount int) error
}

type messageBountiesRepo struct {
	db  *sql.DB
	log *logrus.Logger
}

func NewMessageBountiesRepo(
	db *sql.DB,
	log *logrus.Logger,
) MessageBountiesRepo {
	return &messageBountiesRepo{
		db:  db,
		log: log,
	}
}

// Init initialises the message bounties repo.
func (r *messageBountiesRepo) Init() error {
	return nil
}

// List will retrieve and list message bounties matching the provided criteria.
func (r *messageBountiesRepo) List(
	filter *types.ListMessageBountiesFilter,
	pageSize int,
	pageToken string,
) ([]*types.MessageBounty, string, error) {
	var args []interface{}
	var query = getMessageBountyQueries()[messageBountiesList]

	// Prepare query
	query, args = r.applyFilter(query, filter, pageSize, pageToken)

	// Execute the query
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, "", err
	}

	// Parse rows
	var messageBounties []*types.MessageBounty
	messageBounties, err = r.scanMessageBounties(rows)
	if err != nil {
		return nil, "", err
	}

	// No results
	if len(messageBounties) == 0 {
		return messageBounties, "", nil
	}

	// Return the results along with the id of the last account as a next page token
	return messageBounties, fmt.Sprint(messageBounties[len(messageBounties)-1].MessageId), nil
}

// Create will create a new message bounty.
func (r *messageBountiesRepo) Create(
	messageBounty *types.MessageBounty,
) (*types.MessageBounty, error) {
	var _, err = r.db.Exec(
		getMessageBountyQueries()[messageBountyCreate],
		messageBounty.MessageId,
		messageBounty.UserId,
		messageBounty.ChannelId,
		messageBounty.CurrentBounty,
		messageBounty.Status,
		messageBounty.AwardedTo,
	)
	if err != nil {
		return nil, err
	}

	// Retrieve the row
	rows, _, err := r.List(
		&types.ListMessageBountiesFilter{
			MessageId: messageBounty.MessageId,
		},
		1,
		"",
	)

	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("failed to retrieve new message bounty: %v", messageBounty.MessageId)
	}

	// Populate the map to return
	return rows[0], nil
}

// BoostBounty will boost the amount currently being offered for a bounty.
func (r *messageBountiesRepo) BoostBounty(
	messageId string,
	amount int,
) error {
	var _, err = r.db.Exec(
		getMessageBountyQueries()[messageBountyBoost],
		amount,
		messageId,
	)
	if err != nil {
		return err
	}

	return nil
}

// Update will modify an existing message bounty.
func (r *messageBountiesRepo) Update(
	messageBounty *types.MessageBounty,
) (*types.MessageBounty, error) {
	var _, err = r.db.Exec(
		getMessageBountyQueries()[messageBountyUpdate],
		messageBounty.CurrentBounty,
		messageBounty.Status,
		messageBounty.AwardedTo,
		messageBounty.MessageId,
	)
	if err != nil {
		return nil, err
	}

	// TODO: Check rows affected, if 0 error.

	// Retrieve the row
	rows, _, err := r.List(
		&types.ListMessageBountiesFilter{
			MessageId: messageBounty.MessageId,
		},
		1,
		"",
	)

	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("failed to retrieve the updated message bounty: %v", messageBounty.MessageId)
	}

	// Populate the map to return
	return rows[0], nil
}

func (r *messageBountiesRepo) applyFilter(
	query string,
	filter *types.ListMessageBountiesFilter,
	pageSize int,
	pageToken string,
) (string, []interface{}) {
	var (
		clauses []string
		args    []interface{}
	)

	if filter == nil {
		return query, args
	}

	// Filter by id if provided
	if filter.MessageId != "" {
		clauses = append(clauses, "message_id = ?")
		args = append(args, filter.MessageId)
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

// scanMessageBounties populates a slice of structs from db rows
func (r *messageBountiesRepo) scanMessageBounties(rows *sql.Rows) ([]*types.MessageBounty, error) {

	var res []*types.MessageBounty

	for rows.Next() {

		var (
			messageBounty types.MessageBounty
			created       time.Time
			updated       time.Time
		)

		// Populate the row
		if err := rows.Scan(
			&messageBounty.MessageId,
			&messageBounty.UserId,
			&messageBounty.ChannelId,
			&messageBounty.CurrentBounty,
			&messageBounty.Status,
			&messageBounty.AwardedTo,
			&created,
			&updated,
		); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, err
		}

		// Assign timestamps
		messageBounty.Created = *timestamppb.New(created)
		messageBounty.Updated = *timestamppb.New(updated)

		res = append(res, &messageBounty)
	}

	return res, nil
}
