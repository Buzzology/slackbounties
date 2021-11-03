CREATE TABLE `bot_messages` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `message_id` varchar(36) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `sent_message_id` varchar(36) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `channel_id` varchar(36) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `reaction` varchar(36) COLLATE utf8mb4_bin DEFAULT NULL,
  `status` int(11) DEFAULT NULL,
  `target_user_id` varchar(36) CHARACTER SET utf8mb4 NOT NULL DEFAULT '',
  `created` datetime DEFAULT NULL,
  `updated` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;