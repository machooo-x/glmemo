package database

import (
	"database/sql"
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
			reg_time int unsigned not null,
			name varchar(20) not null,
			pwd varchar(20) not null);`
	_, err = Mysql.Exec(sqlStmt)
	if err != nil {
		panic(err)
	}

	/*
		create record
		filename 最多支持两个文件，文件名用逗号隔开
	*/
	sqlStmt = `
	create table if not exists record (
		id char(36) not null,
		user_id char(36) not null,
		update_time char(10) not null,
		title varchar(20) not null,
		text varchar(1024) not null,
		filename varchar(200),
		size int UNSIGNED not null,
		PRIMARY KEY (id),
		constraint foreign key(user_id) references user(uuid));`
	_, err = Mysql.Exec(sqlStmt)
	if err != nil {
		panic(err)
	}

	/*
		create temprecord
		filename 最多支持两个文件，文件名用逗号隔开
	*/
	sqlStmt = `
	create table if not exists temp_record (
		record_id char(36) not null,
		user_id char(36) not null,
		update_time char(10) not null,
		title varchar(20) not null,
		text varchar(1024) not null,
		filename varchar(200),
		size int UNSIGNED not null,
		is_add_save bool not null,
		PRIMARY KEY (record_id),
		constraint foreign key(user_id) references user(uuid));`
	_, err = Mysql.Exec(sqlStmt)
	if err != nil {
		panic(err)
	}

	/* create vip */
	sqlStmt = `
			create table if not exists vip (
				user_id char(36) not null,
				starttime char(10) not null,
				endtime char(10) not null,
				used int not null,
				status bool not null,
				PRIMARY KEY (user_id),
				constraint foreign key(user_id) references user(uuid));`
	_, err = Mysql.Exec(sqlStmt)
	if err != nil {
		panic(err)
	}

}
