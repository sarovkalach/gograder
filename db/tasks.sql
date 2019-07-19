SET NAMES utf8;
SET time_zone = "+04:00";
SET foreign_key_checks = 0;
SET sql_mode = 'NO_AUTO_VALUE_ON_ZERO';

DROP TABLE IF EXISTS `tasks`;
CREATE TABLE `tasks` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `status` int(11) DEFAULT false,
  `course` varchar(255) NOT NULL,
  `name` varchar(255) NOT NULL,
  `filename` varchar(255) NOT NULL,
  `s3bucketname` varchar(255) NOT NULL,
  `user_id` int(11) NOT NULL,
  FOREIGN KEY (`user_id`) REFERENCES `users`(`id`)
      ON UPDATE CASCADE
      ON DELETE RESTRICT,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;