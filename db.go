package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"

	"github.com/coopernurse/gorp"
)

func init() {
	dbmap := initDb()

	configureORM(dbmap)
}

func initDb() *gorp.DbMap {
	// connect to db using standard Go database/sql API
	// use whatever database/sql driver you wish
	db, err := sql.Open("sqlite3", "/tmp/pr_db.sqlite3")
	checkErr(err, "sql.Open failed")

	// construct a gorp DbMap
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}

	return dbmap
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}
