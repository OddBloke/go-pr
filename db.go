package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"

	"github.com/coopernurse/gorp"
)

func initDb() *gorp.DbMap {
	log.Print("Initialising database...")
	// connect to db using standard Go database/sql API
	// use whatever database/sql driver you wish
	db, err := sql.Open("sqlite3", "/tmp/pr_db.sqlite3")
	checkErr(err, "sql.Open failed")
	log.Print("Database successfully opened.")

	// construct a gorp DbMap
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}
	log.Print("dbmap constructed.")

	return dbmap
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}
