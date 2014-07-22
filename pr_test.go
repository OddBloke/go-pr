package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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

func (s *ElectionSuite) PerformRequest(method string, relativePath string, body string) *httptest.ResponseRecorder {
	path := fmt.Sprintf("http://test.example.com%s", relativePath)
	w := httptest.NewRecorder()
	r, err := http.NewRequest(method, path, strings.NewReader(body))
	checkErr(err, "Request creation failed")

	testApp := CreateApplication(s.dbmap)
	testApp.router.ServeHTTP(w, r)
	return w
}

func (s *ElectionSuite) TestAddElectionCreatesOneElection(c *C) {
	s.PerformRequest("POST", "/elections", `{"name": "Test Election"}`)

	count, err := s.dbmap.SelectInt("select count(*) from elections")
	checkErr(err, "Getting count failed")
	c.Assert(count, Equals, int64(1))
}

func (s *ElectionSuite) TestAddElectionCreatesElectionWithCorrectName(c *C) {
	s.PerformRequest("POST", "/elections", `{"name": "Test Election"}`)

	var createdElection Election
	err := s.dbmap.SelectOne(&createdElection, "select * from elections")
	if err != nil {
		c.Error(err)
	}
	c.Assert(createdElection.Name, Matches, "Test Election")
}

func (s *ElectionSuite) TestAddElectionRejectsZeroLengthName(c *C) {
	recorder := s.PerformRequest("POST", "/elections", `{"name": ""}`)

	c.Check(recorder.Code, Equals, 400)
	c.Check(recorder.Body.String(), Matches, "Empty name forbidden.\n?")

	count, err := s.dbmap.SelectInt("select count(*) from elections")
	checkErr(err, "Getting count failed")
	c.Check(count, Equals, int64(0))
}

func (s *ElectionSuite) TestAddElectionRejectsDuplicateNames(c *C) {
	s.PerformRequest("POST", "/elections", `{"name": "Duplicate"}`)
	recorder := s.PerformRequest("POST", "/elections", `{"name": "Duplicate"}`)

	c.Check(recorder.Code, Equals, 400)
	c.Check(recorder.Body.String(), Matches, "Name taken.\n?")

	count, err := s.dbmap.SelectInt("select count(*) from elections")
	checkErr(err, "Getting count failed")
	c.Check(count, Equals, int64(1))
}

func (s *ElectionSuite) TestGetElectionReturnsElectionName(c *C) {
	election := Election{Name: "my test name"}
	s.dbmap.Insert(&election)

	recorder := s.PerformRequest("GET", fmt.Sprintf("/elections/%d", election.Id), "")
	returnedElection := Election{}
	json.Unmarshal(recorder.Body.Bytes(), &returnedElection)
	c.Assert(returnedElection.Name, Matches, "my test name")
}

func (s *ElectionSuite) TestGetElection404sForUnknownElection(c *C) {
	recorder := s.PerformRequest("GET", "/elections/1", "")
	c.Check(recorder.Code, Equals, 404)
}

func (s *ElectionSuite) TestListElectionsReturnsEmptyList(c *C) {
	recorder := s.PerformRequest("GET", "/elections", "")
	c.Check(recorder.Code, Equals, 200)

	c.Check(recorder.Body.String(), Equals, "[]")
}
