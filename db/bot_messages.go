package db

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/buzzology/slack_bot/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type BotMessagesRepo interface {
	// Init will initialise our bot message repo.
	Init() error

	// List will return a collection of bot messages.
	List(filter *types.ListBotMessagesFilter,
		pageSize int,
		pageToken string,
	) ([]*types.BotMessage, string, error)

	// Create will create a new bot message.
	Create(botMessage *types.BotMessage) error

	// Update will update an existing bot message.
	Update(botMessage *types.BotMessage) error
}

type botMessagesRepo struct {
	db  *sql.DB
	log *logrus.Logger
}

func NewBotMessagesRepo(db *sql.DB, log *logrus.Logger) BotMessagesRepo {
	return &botMessagesRepo{db: db, log: log}
}

// Init will initialise our bot message repo.
func (r *botMessagesRepo) Init() error {
	return nil
}

// Create will create a new bot message.
func (r *botMessagesRepo) Create(botMessage *types.BotMessage) error {
	var _, err = r.db.Exec(
		getBotMessageQueries()[botMessageCreate],
		botMessage.SentMessageId,
		botMessage.MessageId,
		botMessage.ChannelId,
		botMessage.Reaction,
		botMessage.TargetUserId,
	)
	if err != nil {
		return errors.Wrap(err, "failed to create bot message")
	}

	return nil
}

// Update will update an existing bot message.
func (r *botMessagesRepo) Update(botMessage *types.BotMessage) error {
	var _, err = r.db.Exec(
		getBotMessageQueries()[botMessageUpdate],
		botMessage.Status,
		botMessage.Id,
	)
	if err != nil {
		return errors.Wrap(err, "failed to update bot message")
	}

	return nil
}

// List will return a collection of bot messages.
func (r *botMessagesRepo) List(
	filter *types.ListBotMessagesFilter,
	pageSize int,
	pageToken string,
) ([]*types.BotMessage, string, error) {
	var args []interface{}
	var query = getBotMessageQueries()[botMessagesList]

	// Prepare query.
	query, args = r.applyFilter(query, filter, pageSize, pageToken)

	// Execute the query
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, "", err
	}

	// Parse the rows
	var botMessages []*types.BotMessage
	botMessages, err = scanBotMessages(rows)
	if err != nil {
		return nil, "", err
	}

	// No rows returned
	if len(botMessages) == 0 {
		return botMessages, "", nil
	}

	// Return the bot messages
	return botMessages, fmt.Sprint(botMessages[len(botMessages)-1].Id), nil
}

func (r *botMessagesRepo) applyFilter(
	query string,
	filter *types.ListBotMessagesFilter,
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

	// Filter by id
	if filter.Id != "" {
		clauses = append(clauses, "id = ?")
		args = append(args, filter.Id)
	}

	// Filter by message id
	if filter.MessageId != "" {
		clauses = append(clauses, "message_id = ?")
		args = append(args, filter.MessageId)
	}

	// Filter by sent message id
	if filter.SentMessageId != "" {
		clauses = append(clauses, "sent_message_id = ?")
		args = append(args, filter.SentMessageId)
	}

	// Filter by channel id
	if filter.ChannelId != "" {
		clauses = append(clauses, "channel_id = ?")
		args = append(args, filter.ChannelId)
	}

	// Filter by user id
	if filter.UserId != "" {
		clauses = append(clauses, "target_user_id = ?")
		args = append(args, filter.UserId)
	}

	// Filter by status
	if filter.Status > 0 {
		clauses = append(clauses, "status = ?")
		args = append(args, filter.Status)
	}

	// Filter by reaction
	if filter.Reaction != "" {
		clauses = append(clauses, "reaction = ?")
		args = append(args, filter.Reaction)
	}

	if len(clauses) != 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}

	// Validate page size.
	if pageSize > 100 || pageSize < 1 {
		pageSize = 100
	}

	// Limit page size.
	pageTokenI, err := strconv.Atoi(pageToken)
	if err != nil {
		query += fmt.Sprintf(" LIMIT %v, %v", pageTokenI, pageSize)
	} else {
		r.log.Warningf("invalid page token provided: %v", pageToken)
		query += fmt.Sprintf(" LIMIT 0, %v", pageSize)
	}

	return query, args
}

// scanBotMessage will scan a row into a bot message.
func scanBotMessages(rows *sql.Rows) ([]*types.BotMessage, error) {
	var botMessages []*types.BotMessage

	for rows.Next() {
		var (
			botMessage types.BotMessage
			created    time.Time
			updated    time.Time
		)

		if err := rows.Scan(
			&botMessage.Id,
			&botMessage.MessageId,
			&botMessage.SentMessageId,
			&botMessage.ChannelId,
			&botMessage.Reaction,
			&botMessage.Status,
			&botMessage.TargetUserId,
			&created,
			&updated,
		); err != nil {
			return nil, err
		}

		// Assign timestamps
		botMessage.Created = *timestamppb.New(created)
		botMessage.Updated = *timestamppb.New(updated)

		botMessages = append(botMessages, &botMessage)
	}

	return botMessages, nil
}
