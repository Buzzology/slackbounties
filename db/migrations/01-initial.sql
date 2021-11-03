create table bot_state
(
    id             int unsigned auto_increment
        primary key,
    created        datetime default CURRENT_TIMESTAMP not null,
    day_tickover   datetime default CURRENT_TIMESTAMP not null,
    week_tickover  datetime default CURRENT_TIMESTAMP not null,
    month_tickover datetime default CURRENT_TIMESTAMP not null,
    year_tickover  datetime default CURRENT_TIMESTAMP not null
)
    collate = utf8mb4_bin;

create table channel_accounts
(
    id               int unsigned auto_increment
        primary key,
    user_id          varchar(36) default '' not null,
    channel_id       varchar(36) default '' not null,
    balance          int                    not null,
    earned_today     int                    not null,
    spent_today      int                    not null,
    earned_this_week int                    not null,
    spent_this_week  int                    not null,
    earned_this_year int                    not null,
    spent_this_year  int                    not null,
    earned_all_time  int                    not null,
    spent_all_time   int                    not null,
    created          datetime               not null,
    updated          datetime               not null
)
    collate = utf8mb4_bin;

create table message_bounties
(
    message_id     varchar(36) default '' not null
        primary key,
    user_id        varchar(36) default '' not null,
    current_bounty int                    not null,
    status         int                    not null,
    awarded_to     varchar(36) default '' not null,
    created        datetime               not null,
    updated        datetime               not null,
    channel_id     varchar(36) default '' not null
)
    collate = utf8mb4_bin;

