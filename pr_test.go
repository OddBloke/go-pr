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

type PRSuite struct {
	dbmap     *gorp.DbMap
	app       Application
	createURL string
	tableName string
}

func (s *PRSuite) SetUpSuite(c *C) {
	s.dbmap = createTestDatabase()
	s.app = CreateApplication(s.dbmap)
}

func (s *PRSuite) SetUpTest(c *C) {
	err := s.dbmap.TruncateTables()
	if err != nil {
		c.Error(err)
	}
}

func (s *PRSuite) TearDownSuite(c *C) {
	s.dbmap.Db.Close()
}

func (s *PRSuite) PerformRequest(method string, relativePath string, body string) *httptest.ResponseRecorder {
	path := fmt.Sprintf("http://test.example.com%s", relativePath)
	w := httptest.NewRecorder()
	r, err := http.NewRequest(method, path, strings.NewReader(body))
	checkErr(err, "Request creation failed")

	s.app.handler.ServeHTTP(w, r)
	return w
}

func (s *PRSuite) TestAddReturns201(c *C) {
	recorder := s.PerformRequest("POST", s.createURL, `{"name": "Test Name"}`)

	c.Check(recorder.Code, Equals, 201)
}

func (s *PRSuite) TestAddCreatesOneEntity(c *C) {
	s.PerformRequest("POST", s.createURL, `{"name": "Test Name"}`)

	query := fmt.Sprintf("SELECT count(*) FROM %s", s.tableName)
	count, err := s.dbmap.SelectInt(query)
	checkErr(err, "Getting count failed")
	c.Assert(count, Equals, int64(1))
}

type ElectionSuite struct {
	PRSuite
}

func (s *ElectionSuite) SetUpSuite(c *C) {
	s.PRSuite.SetUpSuite(c)
	s.createURL = "/elections"
	s.tableName = "elections"
}

var _ = Suite(&ElectionSuite{})

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

func (s *ElectionSuite) TestGetElectionReturns200(c *C) {
	election := Election{Name: "my test name"}
	s.dbmap.Insert(&election)

	recorder := s.PerformRequest("GET", fmt.Sprintf("/elections/%d", election.Id), "")

	c.Check(recorder.Code, Equals, 200)
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

func (s *ElectionSuite) TestListElectionsReturnsListOfCorrectLength(c *C) {
	election := Election{Name: "my test name"}
	other_election := Election{Name: "my other name"}
	third_election := Election{Name: "my third name"}
	s.dbmap.Insert(&election, &other_election, &third_election)

	recorder := s.PerformRequest("GET", "/elections", "")

	var electionList []Election
	json.Unmarshal(recorder.Body.Bytes(), &electionList)
	c.Check(len(electionList), Equals, 3)
}

func (s *ElectionSuite) TestListElectionReturnsExistingElections(c *C) {
	election := Election{Name: "my test name"}
	other_election := Election{Name: "my other name"}
	s.dbmap.Insert(&election, &other_election)

	recorder := s.PerformRequest("GET", "/elections", "")

	expectedElectionNames := map[string]int{
		"my test name":  0,
		"my other name": 0,
	}
	var electionList []Election
	json.Unmarshal(recorder.Body.Bytes(), &electionList)
	actualElectionNames := make(map[string]int)
	for _, election := range electionList {
		actualElectionNames[election.Name] = 0
	}
	c.Check(actualElectionNames, DeepEquals, expectedElectionNames)
}

type CandidatesSuite struct {
	PRSuite
}

func (s *CandidatesSuite) SetUpTest(c *C) {
	s.PRSuite.SetUpTest(c)

	// Set up test election
	election := Election{Name: "my test name"}
	err := s.dbmap.Insert(&election)
	if err != nil {
		c.Error(err)
	}

	s.createURL = fmt.Sprintf("/elections/%d/candidates", election.Id)
	s.tableName = "candidates"
}

var _ = Suite(&CandidatesSuite{})

func (s *CandidatesSuite) TestAddCandidateReturns404ForMissingElection(c *C) {
	recorder := s.PerformRequest("POST", "/elections/1234/candidates", `{"name": "Test Candidate"}`)

	c.Check(recorder.Code, Equals, 404)
}

func (s *CandidatesSuite) TestAddCandidateCreatesCandidateWithCorrectName(c *C) {
	s.PerformRequest("POST", s.createURL, `{"name": "Test Candidate"}`)

	var createdCandidate Candidate
	err := s.dbmap.SelectOne(&createdCandidate, "select * from candidates")
	if err != nil {
		c.Error(err)
	}
	c.Assert(createdCandidate.Name, Matches, "Test Candidate")
}
