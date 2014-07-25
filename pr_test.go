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
	dbmap            *gorp.DbMap
	app              Application
	createURL        string
	fetchOneTemplate string
	tableName        string
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

func (s *PRSuite) TestAddCreatesEntityWithCorrectName(c *C) {
	s.PerformRequest("POST", s.createURL, `{"name": "Test Name"}`)

	query := fmt.Sprintf("select Name from %s", s.tableName)
	name, err := s.dbmap.SelectStr(query)
	if err != nil {
		c.Error(err)
	}
	c.Assert(name, Matches, "Test Name")
}

func (s *PRSuite) TestAddRejectsZeroLengthName(c *C) {
	recorder := s.PerformRequest("POST", s.createURL, `{"name": ""}`)

	c.Check(recorder.Code, Equals, 400)
	c.Check(recorder.Body.String(), Matches, "Empty name forbidden.\n?")

	query := fmt.Sprintf("SELECT count(*) FROM %s", s.tableName)
	count, err := s.dbmap.SelectInt(query)
	checkErr(err, "Getting count failed")
	c.Check(count, Equals, int64(0))
}

func (s *PRSuite) TestAddRejectsDuplicateNames(c *C) {
	s.PerformRequest("POST", s.createURL, `{"name": "Duplicate"}`)
	recorder := s.PerformRequest("POST", s.createURL, `{"name": "Duplicate"}`)

	c.Check(recorder.Code, Equals, 400)
	c.Check(recorder.Body.String(), Matches, "Name taken.\n?")

	query := fmt.Sprintf("SELECT count(*) FROM %s", s.tableName)
	count, err := s.dbmap.SelectInt(query)
	checkErr(err, "Getting count failed")
	c.Check(count, Equals, int64(1))
}

type ElectionSuite struct {
	PRSuite
}

func (s *ElectionSuite) SetUpSuite(c *C) {
	s.PRSuite.SetUpSuite(c)
	s.createURL = "/elections"
	s.fetchOneTemplate = "/elections/%d"
	s.tableName = "elections"
}

func (s *ElectionSuite) CreateEntity(c *C, name string) int {
	election := Election{Name: "my test name"}
	err := s.dbmap.Insert(&election)
	if err != nil {
		c.Error(err)
	}
	return election.Id
}

var _ = Suite(&ElectionSuite{})

func (s *ElectionSuite) TestGetElectionReturns200(c *C) {
	entityId := s.CreateEntity(c, "my test name")

	recorder := s.PerformRequest("GET", fmt.Sprintf(s.fetchOneTemplate, entityId), "")

	c.Check(recorder.Code, Equals, 200)
}

func (s *ElectionSuite) TestGetElectionReturnsElectionName(c *C) {
	entityId := s.CreateEntity(c, "my test name")

	recorder := s.PerformRequest("GET", fmt.Sprintf(s.fetchOneTemplate, entityId), "")
	returnedElection := Election{}
	json.Unmarshal(recorder.Body.Bytes(), &returnedElection)
	c.Assert(returnedElection.Name, Matches, "my test name")
}

func (s *ElectionSuite) TestGetElection404sForUnknownElection(c *C) {
	recorder := s.PerformRequest("GET", fmt.Sprintf(s.fetchOneTemplate, 1234), "")

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

func (s *CandidatesSuite) CreateElection(c *C, name string) Election {
	// Set up test election
	election := Election{Name: name}
	err := s.dbmap.Insert(&election)
	if err != nil {
		c.Error(err)
	}
	return election
}

func (s *CandidatesSuite) SetUpTest(c *C) {
	s.PRSuite.SetUpTest(c)

	election := s.CreateElection(c, "my test name")
	s.createURL = fmt.Sprintf("/elections/%d/candidates", election.Id)
	s.tableName = "candidates"
}

var _ = Suite(&CandidatesSuite{})

func (s *CandidatesSuite) TestAddCandidateReturns404ForMissingElection(c *C) {
	recorder := s.PerformRequest("POST", "/elections/1234/candidates", `{"name": "Test Candidate"}`)

	c.Check(recorder.Code, Equals, 404)
}

func (s *CandidatesSuite) TestAddCandidateDoesNotRejectSameNameForDifferentElections(c *C) {
	payload := `{"name": "Test Candidate"}`
	recorder := s.PerformRequest("POST", s.createURL, payload)
	c.Check(recorder.Code, Equals, 201)

	secondElection := s.CreateElection(c, "Second Election")
	recorder = s.PerformRequest("POST", fmt.Sprintf("/elections/%d/candidates", secondElection.Id), payload)
	c.Check(recorder.Code, Equals, 201)
}
