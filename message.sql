CREATE TABLE `message` (
  `uuid` varchar(255) NOT NULL,
  `message_id` varchar(255) NOT NULL,
  `subscription` varchar(255) NOT NULL,
  `data` text,
  `decoded_data` text,
  `attribute` text,
  `publish_time` datetime NOT NULL,
  `delivery_attempt` int(11) DEFAULT NULL,
  `ordering_key` text
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

ALTER TABLE `message` ADD PRIMARY KEY (`uuid`);