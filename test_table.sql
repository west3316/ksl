CREATE DATABASE IF NOT EXISTS `test`;

ALTER DATABASE `test` CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci; 

CREATE TABLE `test`.`user_charge` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '唯一ID',
  `user_id` int(10) unsigned NOT NULL COMMENT '用户ID',
  `create_at` datetime NOT NULL COMMENT '充值时间',
  `value` decimal(20,2) NOT NULL COMMENT '充值金额',
  `result` enum('NoResult','Success','Fail','Locked') NOT NULL COMMENT '充值结果',
  `desc` text DEFAULT NULL COMMENT '备注',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=640 DEFAULT CHARSET=utf8mb4

CREATE TABLE `test`.`user_operate_records` (
  `id` INT NOT NULL,
  `type` VARCHAR(45) NOT NULL,
  `at` DATETIME NOT NULL,
  PRIMARY KEY (`id`));
