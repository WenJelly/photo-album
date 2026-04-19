CREATE TABLE `user`
(
    `id`           bigint(20)    NOT NULL AUTO_INCREMENT COMMENT 'id',
    `userEmail`    varchar(256)  NOT NULL DEFAULT '' COMMENT '邮箱',
    `userMobile`   varchar(20)   NOT NULL DEFAULT '' COMMENT '手机号',
    `userPassword` varchar(512)  NOT NULL DEFAULT '' COMMENT '密码',
    `userName`     varchar(256)  NOT NULL DEFAULT '' COMMENT '用户昵称',
    `userAvatar`   varchar(1024) NOT NULL DEFAULT '' COMMENT '用户头像',
    `userProfile`  varchar(512)  NOT NULL DEFAULT '' COMMENT '用户简介',
    `userRole`     varchar(256)  NOT NULL DEFAULT 'user' COMMENT '用户角色：user/admin',
    `editTime`     datetime      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '编辑时间',
    `createTime`   datetime      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updateTime`   datetime      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `isDelete`     tinyint(4)    NOT NULL DEFAULT '0' COMMENT '是否删除',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_userAccount` (`userEmail`),
    KEY `idx_userName` (`userName`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci COMMENT ='用户';