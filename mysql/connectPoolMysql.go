package tinyMysql

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var connectPool *connectPoolMysql

type connectPoolMysql struct {
	db *sql.DB
}

func GetMysqlPool() *connectPoolMysql {
	initConnectPoolMysql()
	return connectPool
}

func initConnectPoolMysql() {
	if connectPool == nil {
		fmt.Println("initConnectPoolMysql +++++++++++")
		connectPool = new(connectPoolMysql)
	}
}

func GetDB(conStr string, connectionLimit int) *sql.DB {
	initConnectPoolMysql()
	if connectPool.db == nil {
		db, err := sql.Open("mysql", conStr)
		if err != nil {
			db.Close()
			panic(err)
		}
		connectPool.db = db

		connectPool.db.SetConnMaxLifetime(time.Minute * 3)
		connectPool.db.SetMaxOpenConns(connectionLimit)
		connectPool.db.SetMaxIdleConns(50)
		connectPool.db.SetConnMaxIdleTime(time.Second * 60)

		fmt.Println("mysql db is opened +++++++++++ !!!!!!!")
	}

	return connectPool.db
}
