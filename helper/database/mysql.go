package database

import (
	"database/sql"
	"fmt"
	"glmemo/config"

	// mysql driver
	_ "github.com/go-sql-driver/mysql"
)

// Mysql db
var Mysql *sql.DB

func init() {
	// syslog.Clog.Infoln(true, "初始化 Mysql Service...")

	MySQLMaxConnection, _ := config.GLMEMO.Section("mysql").Key("MySQLMaxConnection").Int()

	var err error
	/* init mysql */
	Mysql, err = sql.Open("mysql", config.GLMEMO.Section("mysql").Key("MySQLConnectString").String())
	if err != nil {
		panic(err)
	}
	Mysql.SetMaxOpenConns(MySQLMaxConnection)

	/* create user */
	sqlStmt := `
		create table if not exists user (
			uuid char(36) not null PRIMARY KEY,
			name varchar(20) not null,
			pwd varchar(20) not null,
			mailbox varchar(320) not null,
			reg_time char(10) not null,
			last_time char(10) not null,
			UNIQUE KEY name_idx (name))charset utf8 collate utf8_general_ci;`
	_, err = Mysql.Exec(sqlStmt)
	if err != nil {
		panic(err)
	}

	/* create tag */
	sqlStmt = `
		create table if not exists tag (
			id int not null auto_increment,
			user_id char(36) not null,
			tag_name varchar(10) not null,
			sum int default 0,
			PRIMARY KEY (id),
			constraint foreign key(user_id) references user(uuid),
			KEY user_id_and_tag_name_idx (user_id,tag_name))charset utf8 collate utf8_general_ci;`
	_, err = Mysql.Exec(sqlStmt)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	/*
		create record
	*/
	sqlStmt = `
	create table if not exists record (
		id char(36) not null,
		user_id char(36) not null,
		update_time char(10) not null,
		title varchar(20) not null,
		text varchar(1024) not null,
		tag_name varchar(10) not null,
		filename varchar(150),
		filepath varchar(200),
		status bool default 0,
		PRIMARY KEY (id),
		constraint foreign key(user_id) references user(uuid),
		constraint foreign key(user_id,tag_name) references tag(user_id,tag_name))charset utf8 collate utf8_general_ci;`
	_, err = Mysql.Exec(sqlStmt)
	if err != nil {
		panic(err)
	}

	/*
		create temprecord
	*/
	sqlStmt = `
	create table if not exists temp_record (
		record_id char(36) not null,
		user_id char(36) not null,
		update_time char(10) not null,
		title varchar(20) not null,
		text varchar(1024) not null,
		tag_name varchar(10) not null,
		filepath varchar(200),
		is_add_save bool not null,
		PRIMARY KEY (record_id),
		constraint foreign key(user_id) references user(uuid),
		constraint foreign key(user_id,tag_name) references tag(user_id,tag_name))charset utf8 collate utf8_general_ci;`
	_, err = Mysql.Exec(sqlStmt)
	if err != nil {
		panic(err)
	}

	/* create manager */
	sqlStmt = `
		create table if not exists manager (
		id char(36) not null PRIMARY KEY,
		name varchar(20) not null,
		pwd varchar(20) not null,
		reg_time char(10) not null,
		last_time char(10) not null,
		UNIQUE KEY name_idx (name))charset utf8 collate utf8_general_ci;`
	_, err = Mysql.Exec(sqlStmt)
	if err != nil {
		panic(err)
	}

	/* create schedule */
	sqlStmt = `
		create table if not exists schedule (
		id char(36) not null,
		title varchar(20) not null,
		text varchar(1024) not null,
		user_id char(36) not null,
		user_mailbox varchar(320) not null,
		reg_time char(10) not null,
		remind_time char(10) not null,
		PRIMARY KEY (id),
		constraint foreign key(user_id) references user(uuid))charset utf8 collate utf8_general_ci;`
	_, err = Mysql.Exec(sqlStmt)
	if err != nil {
		panic(err)
	}

	/* create vip */
	// sqlStmt = `
	// 		create table if not exists vip (
	// 			user_id char(36) not null,
	// 			starttime char(10) not null,
	// 			endtime char(10) not null,
	// 			used int not null,
	// 			status bool not null,
	// 			PRIMARY KEY (user_id),
	// 			constraint foreign key(user_id) references user(uuid))charset utf8 collate utf8_general_ci;`
	// _, err = Mysql.Exec(sqlStmt)
	// if err != nil {
	// 	panic(err)
	// }

}

