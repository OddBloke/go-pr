package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coopernurse/gorp"
)

func createTestDatabase() *gorp.DbMap {
	db, err := sql.Open("sqlite3", ":memory:")
	checkErr(err, "sql.Open failed")

	// construct a gorp DbMap
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}

	return dbmap
}

func TestAddElection(t *testing.T) {
	dbmap := createTestDatabase()

	w := httptest.NewRecorder()
	r, err := http.NewRequest("POST", "http://test.example.com/elections", nil)
	checkErr(err, "Request creation failed")

	testApp := CreateApplication(dbmap)
	testApp.router.ServeHTTP(w, r)

	count, err := dbmap.SelectInt("select count(*) from elections")
	checkErr(err, "Getting count failed")

	if count != 1 {
		t.Error("Expected 1 election, found ", count)
	}
}
