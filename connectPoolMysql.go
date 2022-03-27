package tiny

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var connectPool *dbConnectPool

type dbConnectPool struct {
	dbMap map[string]*sql.DB
	m     sync.Mutex
}

func GetMysqlPool() *dbConnectPool {
	initConnectPoolMysql()
	return connectPool
}

func initConnectPoolMysql() {
	if connectPool == nil {
		fmt.Println("initConnectPool +++++++++++")
		connectPool = new(dbConnectPool)
		connectPool.dbMap = make(map[string]*sql.DB)
	}
}

func GetDB(conStr string, connectionLimit int, driver string) *sql.DB {
	initConnectPoolMysql()

	db, has := connectPool.dbMap[conStr]
	if has {
		return db
	} else {

		connectPool.m.Lock()
		db, err := sql.Open(driver, conStr)
		if err != nil {
			db.Close()
			connectPool.m.Unlock()
			panic(err)
		}

		db.SetConnMaxLifetime(time.Minute * 3)
		db.SetMaxOpenConns(connectionLimit)
		db.SetMaxIdleConns(50)
		db.SetConnMaxIdleTime(time.Second * 60)

		connectPool.dbMap[conStr] = db
		connectPool.m.Unlock()
		fmt.Println(driver+" db is opened +++++++++++ !!!!!!!", conStr)

		return db
	}
}
