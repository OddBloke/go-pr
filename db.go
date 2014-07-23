package main

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"github.com/coopernurse/gorp"
)

func initDb() *gorp.DbMap {
	log.Info("Initialising database...")
	// connect to db using standard Go database/sql API
	// use whatever database/sql driver you wish
	db, err := sql.Open("sqlite3", "/tmp/pr_db.sqlite3")
	checkErr(err, "sql.Open failed")
	log.Info("Database successfully opened.")

	// construct a gorp DbMap
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}
	log.Info("dbmap constructed.")

	return dbmap
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatal("%s: %s", msg, err.Error())
		os.Exit(1)
	}
}
