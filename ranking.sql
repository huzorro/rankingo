CREATE TABLE ranking_keyword(
    id int NOT NULL AUTO_INCREMENT,  
    msg varchar(500) NOT NULL DEFAULT "",
    logtime timestamp NOT NULL DEFAULT current_timestamp ON update current_timestamp,
    PRIMARY KEY (id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8

CREATE TABLE ranking_detail(
    id int NOT NULL DEFAULT 0,
    uid int NOT NULL DEFAULT 0,
    owner varchar(50) NOT NULL DEFAULT "",
    keyword varchar(50) NOT NULL DEFAULT "",
    destlink varchar(200) NOT NULL DEFAULT "",
    history_order int NOT NULL DEFAULT 0,
    current_order int NOT NULL DEFAULT 0,    
    history_index int NOT NULL DEFAULT 0,
    current_index int NOT NULL DEFAULT 0,
    city_key varchar(50) NOT NULL DEFAULT "",
    province_key varchar(50) NOT NULL DEFAULT "",
    cost int NOT NULL DEFAULT 0,
    status int NOT NULL DEFAULT 1 comment '0:start, 1:cancel',
    logtime timestamp NOT NULL DEFAULT "0000-00-00 00:00:00",
    uptime timestamp NOT NULL DEFAULT current_timestamp ON update current_timestamp,
    PRIMARY KEY (id),
    UNIQUE KEY  `uk_keyword_destlink`(keyword, destlink)
)ENGINE=InnoDB DEFAULT CHARSET=utf8



CREATE TABLE ranking_log(
    logid int NOT NULL AUTO_INCREMENT,  
    id int NOT NULL DEFAULT 0,
    uid int NOT NULL DEFAULT 0,
    owner varchar(50) NOT NULL DEFAULT "",
    keyword varchar(50) NOT NULL DEFAULT "",
    destlink varchar(200) NOT NULL DEFAULT "",
    history_order int NOT NULL DEFAULT 0,
    current_order int NOT NULL DEFAULT 0,    
    history_index int NOT NULL DEFAULT 0,
    current_index int NOT NULL DEFAULT 0,
    city_key varchar(50) NOT NULL DEFAULT "",
    province_key varchar(50) NOT NULL DEFAULT "",
    cost int NOT NULL DEFAULT 0,
    status int NOT NULL DEFAULT 1 comment '0:stop, 1:cancel',
    logtime timestamp NOT NULL DEFAULT "0000-00-00 00:00:00",
    uptime timestamp NOT NULL DEFAULT current_timestamp ON update current_timestamp,
    PRIMARY KEY (logid)
)ENGINE=InnoDB DEFAULT CHARSET=utf8


grant all privileges on ranking.* to ranking@localhost identified by 'woai840511~';
flush privileges;

-- 
-- rbac

CREATE TABLE sp_user (
    id int NOT NULL AUTO_INCREMENT,
    username varchar(100) NOT NULL DEFAULT '',
    password varchar(100) NOT NULL DEFAULT '',
    roleid int NOT NULL DEFAULT 0,
    accessid int NOT NULL DEFAULT 0,
    logtime timestamp NOT NULL DEFAULT current_timestamp ON update current_timestamp,
    PRIMARY KEY (id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8

INSERT INTO sp_user(username, password, roleid, accessid)  VALUES("root", "admin", 2, 1) 


CREATE TABLE sp_role (
    id int NOT NULL DEFAULT 0,  
    name varchar(100) NOT NULL DEFAULT '' comment 'user, services, admin, guess',
    privilege int NOT NULL DEFAULT 0,
    menu int NOT NULL DEFAULT 0,
    logtime timestamp NOT NULL DEFAULT current_timestamp ON update current_timestamp,
    PRIMARY KEY(id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8

INSERT INTO sp_role (id, name, privilege, menu) VALUES (1, "匿名用户", 7, 0), (2, "管理员", 127, 1)

CREATE TABLE sp_node_privilege (
    id int NOT NULL DEFAULT 0,
    name varchar(100) NOT NULL DEFAULT '',
    node varchar(500) NOT NULL DEFAULT '' comment '1:/login, 2:/login/check, 4:/logout, 8:/key/add, 16:/key/update, 32:/key/show, 64:/key/one',
    logtime timestamp NOT NULL DEFAULT current_timestamp ON update current_timestamp,
    PRIMARY KEY(id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8



INSERT INTO sp_node_privilege (id, name, node)  VALUES (1, "登录页", "/login"), (2, "登录验证请求", "/login/check"), (4, "退出登录", "/logout"), (8, "关键字添加", "/key/add"), (16, "关键字更新", "/key/update"), (32, "关键字列表", "/key/show"), (64, "单个关键字", "/key/one")


CREATE TABLE sp_access_privilege (
    id int NOT NULL AUTO_INCREMENT,
    pri_group varchar(500) NOT NULL DEFAULT '' comment '1;2;3;4;5', 
    pri_rule int NOT NULL DEFAULT 0 comment '1:all, 2:allow, 4:ban',
    logtime timestamp NOT NULL DEFAULT current_timestamp ON update current_timestamp,
    PRIMARY KEY(id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8

INSERT INTO sp_access_privilege (pri_group, pri_rule) VALUES ('', 1);

INSERT INTO sp_access_privilege (pri_group, pri_rule) VALUES ('1', 2);

INSERT INTO sp_access_privilege (pri_group, pri_rule) VALUES ('1', 4);

CREATE TABLE sp_menu_template (
    id int NOT NULL DEFAULT 0 comment '1 2 4 8',
    title varchar(100) NOT NULL DEFAULT '' comment '关键词管理', 
    name varchar(100) NOT NULL DEFAULT '' comment 'show', 
    logtime timestamp NOT NULL DEFAULT current_timestamp ON UPDATE current_timestamp,
    PRIMARY KEY(id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8


INSERT INTO sp_menu_template (id, title, name)  VALUES(1, "关键词管理", "show")


SELECT a.id, a.username, a.password, a.roleid, b.name, b.privilege, a.accessid, c.group, c.rule FROM sp_user a 
    INNER JOIN sp_role b ON a.roleid = b.id
    INNER JOIN sp_access_privilege c ON a.accessid = c.id
    WHERE username = ? AND password = ? 

