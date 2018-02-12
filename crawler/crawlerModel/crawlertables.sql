CREATE TABLE `requestInfo` (
`req_id` int(10) NOT NULL AUTO_INCREMENT,
`url` varchar(200) NOT NULL,
`domain` varchar(64) DEFAULT NULL,
`legal` boolean DEFAULT 0,
`created` datetime ,
PRIMARY KEY (`req_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;


CREATE TABLE `blogRecord` (
  `blog_id` int(10) NOT NULL AUTO_INCREMENT,
  `url` varchar(200) NOT NULL,
  `title` varchar(200) DEFAULT NULL,
  `author` varchar(100) DEFAULT NULL,
  `viewNum` SMALLINT(6)  DEFAULT 0,
  `commendNum` SMALLINT(6) DEFAULT 0,
  `blogsize` int(10) DEFAULT 0,
  -- `response` varchar(10000) DEFAULT NULL,
  `created` datetime ,
  PRIMARY KEY (`blog_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;
