package db

// Database object names.
const (
	botStateGet    = "get"
	botStateCreate = "create"
	botStateUpdate = "update"

	botMessagesList  = "list"
	botMessageCreate = "create"
	botMessageUpdate = "update"

	channelAccountsList               = "list"
	channelAccountCreate              = "create"
	channelAccountUpdate              = "update"
	channelAccountSpend               = "spend"
	channelAccountAward               = "award"
	channelAccountActiveTodayCount    = "today_count"
	channelAccountActiveThisWeekCount = "this_week_count"
	channelAccountActiveThisYearCount = "this_year_count"
	channelAccountActiveAllTimeCount  = "all_time_count"
	channelAccountResetDaily          = "reset_daily"
	channelAccountResetWeekly         = "reset_weekly"
	channelAccountResetYearly         = "reset_yearly"
	channelAccountApplyIncomeAndDecay = "income_and_decay"
	channelAccountsDistinctChannels   = "channel_accounts_distinct_channels"

	messageBountiesList = "list"
	messageBountyCreate = "create"
	messageBountyUpdate = "update"
	messageBountyBoost  = "boost"
)

func getChannelAccountQueries() map[string]string {
	return map[string]string{
		channelAccountsDistinctChannels: `
			SELECT DISTINCT(channel_id) FROM channel_accounts
		`,
		channelAccountsList: `
			SELECT 
				id,
				user_id,
				channel_id,
				balance,
				earned_today,
				spent_today,
				earned_this_week,
				spent_this_week,
				earned_this_year,
				spent_this_year,
				earned_all_time,
				spent_all_time,
				created,
				updated
			FROM channel_accounts
		`,
		channelAccountCreate: `
			INSERT INTO channel_accounts( 
				user_id,
				channel_id,
				balance,
				earned_today,
				spent_today,
				earned_this_week,
				spent_this_week,
				earned_this_year,
				spent_this_year,
				earned_all_time,
				spent_all_time,
				created,
				updated
			) VALUES (
				?,
				?,
				?,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				CURRENT_TIMESTAMP,
				CURRENT_TIMESTAMP
			)
		`,
		channelAccountUpdate: `
			UPDATE channel_accounts
			SET balance = ?,
				earned_today = ?,
				spent_today = ?,
				earned_this_week = ?,
				spent_this_week = ?,
				earned_this_year = ?,
				spent_this_year = ?,
				earned_all_time = ?,
				spent_all_time = ?,
				updated = CURRENT_TIMESTAMP
			) 
			WHERE id = ?
		`,
		channelAccountSpend: `
			UPDATE channel_accounts
			SET balance = balance - ?,
				spent_today = spent_today + ?,
				spent_this_week = spent_this_week + ?,
				spent_this_year = spent_this_year + ?,
				spent_all_time = spent_all_time + ?,
				updated = CURRENT_TIMESTAMP
			WHERE id = ?
		`,
		channelAccountAward: `
			UPDATE channel_accounts
			SET balance = balance + ?,
				earned_today = earned_today + ?,
				earned_this_week = earned_this_week + ?,
				earned_this_year = earned_this_year + ?,
				earned_all_time = earned_all_time + ?,
				updated = CURRENT_TIMESTAMP
			WHERE id = ?
		`,
		channelAccountActiveTodayCount: `
			SELECT COUNT(1)
			FROM channel_accounts
			WHERE (spent_today > 0 || earned_today > 0)
				AND channel_id = ?
			`,
		channelAccountActiveThisWeekCount: `
			SELECT COUNT(1)
			FROM channel_accounts
			WHERE (spent_this_week > 0 || earned_this_week > 0)
				AND channel_id = ?
		`,
		channelAccountActiveThisYearCount: `
			SELECT COUNT(1)
			FROM channel_accounts
			WHERE (spent_this_year > 0 || earned_this_year > 0)
				AND channel_id = ?
		`,
		channelAccountActiveAllTimeCount: `
			SELECT COUNT(1)
			FROM channel_accounts
			WHERE (spent_all_time > 0 || earned_all_time > 0)
			       AND channel_id = ?
		`,
		channelAccountResetDaily: `
			UPDATE channel_accounts
			SET spent_today = 0,
				earned_today = 0
		`,
		channelAccountResetWeekly: `
			UPDATE channel_accounts
			SET spent_this_week = 0,
				earned_this_week = 0
		`,
		channelAccountResetYearly: `
			UPDATE channel_accounts
			SET spent_this_year = 0,
				earned_this_year = 0
		`,
		channelAccountApplyIncomeAndDecay: `
			UPDATE channel_accounts
			SET balance = (balance - ? + ?)
			WHERE (balance - ? + ?) > 0
		`,
	}
}

func getBotMessageQueries() map[string]string {
	return map[string]string{
		botMessagesList: `
			SELECT id, message_id, sent_message_id, channel_id, reaction, status, target_user_id, created, updated
			FROM bot_messages
		`,
		botMessageCreate: `
			INSERT INTO bot_messages(message_id, sent_message_id, channel_id, reaction, status, target_user_id, created, updated)
			VALUES(?, ?, ?, ?, 1, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`,
		botMessageUpdate: `
			UPDATE bot_messages
			SET status = ?,
				updated = CURRENT_TIMESTAMP
			WHERE id = ?
		`,
	}
}

func getMessageBountyQueries() map[string]string {
	return map[string]string{
		messageBountiesList: `
			SELECT 
				message_id,
				user_id,
				channel_id,
				current_bounty,
				status,
				awarded_to,
				created,
				updated
			FROM message_bounties
		`,
		messageBountyCreate: `
			INSERT INTO message_bounties( 
				message_id,
				user_id,
				channel_id,
				current_bounty,
				status,
				awarded_to,
				created,
				updated
			) VALUES (
				?,
				?,
				?,
				?,
				?,
				?,
				CURRENT_TIMESTAMP,
				CURRENT_TIMESTAMP
			)
		`,
		messageBountyUpdate: `
			UPDATE message_bounties
			SET current_bounty = ?,
				status = ?,
				awarded_to = ?,
				updated = CURRENT_TIMESTAMP
			WHERE message_id = ?
		`,
		messageBountyBoost: `
			UPDATE message_bounties
			SET current_bounty = current_bounty + ?,
				updated = CURRENT_TIMESTAMP
			WHERE message_id = ?
		`,
	}
}

func getBotStateQueries() map[string]string {
	return map[string]string{
		botStateGet: `
			SELECT 
				id,
				created,
				day_tickover,
				week_tickover,
				month_tickover,
				year_tickover
			FROM bot_state
			WHERE id = (
				SELECT MAX(id)
				FROM bot_state
			)
		`,
		botStateCreate: `
			INSERT INTO bot_state() VALUES()
		`,
		botStateUpdate: `
			UPDATE bot_state
			SET day_tickover = ?,
				week_tickover = ?,
				month_tickover = ?,
				year_tickover = ?
			WHERE id = ?
		`,
	}
}
