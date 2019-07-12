SET NAMES utf8;
SET time_zone = '+00:00';
SET foreign_key_checks = 0;
SET sql_mode = 'NO_AUTO_VALUE_ON_ZERO';

DROP TABLE IF EXISTS `tasks`;
CREATE TABLE `tasks` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `course` varchar(255) NOT NULL,
  `task` varchar(255) NOT NULL,
  `user` varchar(255) NOT NULL,
  `filename` varchar(255) NOT NULL,
  `graded` boolean DEFAULT false,
  
  -- `updated` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

INSERT INTO `tasks` (`id`, `course`, `task`, `user`, `filename`) VALUES
(1,	'Golang course',	'Финальный проект',	`test.txt`,'akalachov'),
-- (2,	'memcache',	'Рассказать про мемкеш с примером использования',	NULL);