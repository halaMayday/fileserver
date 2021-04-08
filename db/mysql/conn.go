package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
)

var db *sql.DB

func init() {
	db, _ = sql.Open("mysql", "root:123456@tcp(192.168.0.112)/fileserver?charset=utf8")
	db.SetMaxOpenConns(1000)
	err := db.Ping()
	if err != nil {
		fmt.Printf("Failed to connect to mysql ,err:" + err.Error())
		//强制让程序退出
		os.Exit(1)
	}
}

func DBConn() *sql.DB {
	return db
}

//func ParseRows(rows *sql.Rows) []map[string]interface{}{
//	columns,err  := rows.Columns()
//
//
//}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}
