CREATE TABLE `admin` (
  `cms_user_id` int(11) NOT NULL AUTO_INCREMENT,
  `role` tinyint(2) NOT NULL DEFAULT 1 COMMENT '0-超级管理员 1-普通管理员',
  `name` varchar(20) NOT NULL DEFAULT '' COMMENT '姓名',
  `password` varchar(128) NOT NULL DEFAULT '' COMMENT '密文pwd',
  `head_photo` varchar(128) NOT NULL DEFAULT '' COMMENT '头像七牛云url',
  `mobile` varchar(11) NOT NULL DEFAULT '' COMMENT '手机号',
  `abandon` tinyint(1) NOT NULL DEFAULT 0 COMMENT '0-有效 1-弃用',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL ON UPDATE CURRENT_TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`cms_user_id`),
  UNIQUE KEY `idx_mobile` (`mobile`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='管理员表';


CREATE TABLE `customer` (
  `customer_id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(20) NOT NULL DEFAULT '' COMMENT '客户姓名',
  `password` varchar(128) NOT NULL DEFAULT '' COMMENT '密文pwd',
  `head_photo` varchar(128) NOT NULL DEFAULT '' COMMENT '头像七牛云url',
  `mobile` varchar(11) NOT NULL DEFAULT '' COMMENT '客户手机号',
  `abandon` tinyint(1) NOT NULL DEFAULT 0 COMMENT '0-有效 1-弃用',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL ON UPDATE CURRENT_TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`customer_id`),
  UNIQUE KEY `idx_mobile` (`mobile`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='客户表';

CREATE TABLE `product` (
  `product_id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL DEFAULT '' COMMENT '产品名称',
  `type` tinyint(1) NOT NULL DEFAULT 0 COMMENT '产品类型 0-国内 1-国外',
  `destination` varchar(64) NOT NULL DEFAULT '' COMMENT '目的地名称',
  `count` int(10) NOT NULL DEFAULT 0 COMMENT '剩余数量',
  `price` int(10) NOT NULL DEFAULT 0 COMMENT '产品价格',
  `valid_start_date` varchar(16) NOT NULL DEFAULT '' COMMENT '产品有效起始时间,example:2019-05-01',
  `valid_end_date` varchar(16) NOT NULL DEFAULT '' COMMENT '产品有效终止时间,example:2019-06-01',
  `show` tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否展示 0-不展示 1-展示',
  `titleResourceIds` varchar(512) NOT NULL DEFAULT '' COMMENT 'qiniu ressource id list，逗号分隔！',
  `detailResourceIds` varchar(512) NOT NULL DEFAULT '' COMMENT 'qiniu ressource id list，逗号分隔！',
  `remarks` varchar(512) NOT NULL DEFAULT '' COMMENT '备注',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL ON UPDATE CURRENT_TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`product_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='产品表';

CREATE TABLE `normal_order` (
  `product_order_id` int(11) NOT NULL AUTO_INCREMENT,
  `customer_id` int(11) NOT NULL DEFAULT 0 COMMENT '客户id',
  `product_id` int(11) NOT NULL DEFAULT 0 COMMENT '产品id',
  `withdraw` tinyint(1) NOT NULL DEFAULT 0 COMMENT '订单是否撤销0-有效 1-撤销',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL ON UPDATE CURRENT_TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`product_order_id`),
  KEY `idx_product_id` (`product_id`),
  KEY `idx_withdraw` (`withdraw`),
  KEY `idx_customer_id` (`customer_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='既有产品订单表';

CREATE TABLE `private_order` (
  `private_order_id` int(11) NOT NULL AUTO_INCREMENT,
  `customer_id` int(11) NOT NULL DEFAULT 0 COMMENT '客户id',
  `destination` varchar(64) NOT NULL DEFAULT '' COMMENT '目的地名称',
  `withdraw` tinyint(1) NOT NULL DEFAULT 0 COMMENT '订单是否撤销0-有效 1-撤销',
  `handled` tinyint(1) NOT NULL DEFAULT 0 COMMENT '订单是否已经被处理0-未处理 1-已处理',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL ON UPDATE CURRENT_TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`private_order_id`),
  KEY `idx_customer_id` (`customer_id`),
  KEY `idx_withdraw` (`withdraw`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='私人订制订单表';


CREATE TABLE `sys_conf` (
  `sys_conf_id` int(11) NOT NULL AUTO_INCREMENT,
  `main_tags` varchar(512) NOT NULL DEFAULT '' COMMENT '主标签列表，逗号分隔',
  `enable` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否使用0-不使用 1-使用',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL ON UPDATE CURRENT_TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`sys_conf_id`),
  KEY `idx_enable` (`enable`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='私人订制订单表';



CREATE TABLE `resource` (
  `resource_id` int(11) NOT NULL AUTO_INCREMENT,
  `qiniu_url` varchar(128) NOT NULL DEFAULT '' COMMENT '七牛云资源存储url',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL ON UPDATE CURRENT_TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`resource_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='七牛云资源url表';