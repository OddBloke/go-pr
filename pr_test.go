package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coopernurse/gorp"
	. "gopkg.in/check.v1"
)

func createTestDatabase() *gorp.DbMap {
	db, err := sql.Open("sqlite3", ":memory:")
	checkErr(err, "sql.Open failed")

	// construct a gorp DbMap
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}

	return dbmap
}

func Test(t *testing.T) { TestingT(t) }

type ElectionSuite struct {
	dbmap *gorp.DbMap
}

var _ = Suite(&ElectionSuite{})

func (s *ElectionSuite) SetUpSuite(c *C) {
	s.dbmap = createTestDatabase()
}

func (s *ElectionSuite) SetUpTest(c *C) {
	err := s.dbmap.TruncateTables()
	if err != nil {
		c.Error(err)
	}
}

func (s *ElectionSuite) TearDownSuite(c *C) {
	s.dbmap.Db.Close()
}

func (s *ElectionSuite) TestAddElectionCreatesOneElection(c *C) {
	w := httptest.NewRecorder()
	r, err := http.NewRequest("POST", "http://test.example.com/elections", nil)
	checkErr(err, "Request creation failed")

	testApp := CreateApplication(s.dbmap)
	testApp.router.ServeHTTP(w, r)

	count, err := s.dbmap.SelectInt("select count(*) from elections")
	checkErr(err, "Getting count failed")

	c.Assert(count, Equals, int64(1))
}
