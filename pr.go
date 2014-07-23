package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/coopernurse/gorp"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jcelliott/lumber"
)

var handler http.Handler
var log lumber.Logger

func handleUnexpectedError(err error, w http.ResponseWriter) bool {
	if err != nil {
		debug.PrintStack()
		log.Error(err.Error())
		http.Error(w, "Server error", 500)
		return true
	}
	return false
}

type Application struct {
	database DB
	handler  http.Handler
}

func (app *Application) configureORM(dbmap *gorp.DbMap) {
	log.Info("Configuring ORM...")
	dbmap.AddTableWithName(Election{}, "elections").SetKeys(true, "Id").ColMap("Name").SetUnique(true)
	dbmap.AddTableWithName(Candidate{}, "candidates").SetKeys(true, "Id")

	err := dbmap.CreateTablesIfNotExists()
	checkErr(err, "Create tables failed")
	log.Info("Tables successfully created.")

	app.database = GorpDB{dbmap}
	log.Info("ORM configured.")
}

func (app *Application) configureRouting() {
	log.Info("Configuring routing...")
	router := mux.NewRouter()
	router.HandleFunc("/elections", app.ListElections).Methods("GET")
	router.HandleFunc("/elections", app.AddElection).Methods("POST")
	router.HandleFunc("/elections/{id}", app.GetElection).Methods("GET")
	router.HandleFunc("/elections/{election_id}/candidates", app.AddCandidate).Methods("POST")
	app.handler = handlers.CombinedLoggingHandler(os.Stdout, router)
	log.Info("Routing configured.")
}

func CreateApplication(dbmap *gorp.DbMap) Application {
	log.Info("Creating application...")
	app := Application{}
	app.configureORM(dbmap)
	app.configureRouting()
	log.Info("Application created.")
	return app
}

func (app Application) AddElection(w http.ResponseWriter, r *http.Request) {
	election := Election{}
	err := json.NewDecoder(r.Body).Decode(&election)
	if handleUnexpectedError(err, w) {
		return
	}
	if len(election.Name) == 0 {
		http.Error(w, "Empty name forbidden.", 400)
		return
	}
	err = app.database.AddElection(election)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			http.Error(w, "Name taken.", 400)
			return
		}
		handleUnexpectedError(err, w)
		return
	}
	w.WriteHeader(201)
}

func getIntURLParameter(w http.ResponseWriter, r *http.Request, parameterName string) (int, error) {
	idString := mux.Vars(r)[parameterName]
	id, err := strconv.Atoi(idString)
	if err == nil {
		return id, nil
	} else if err == strconv.ErrSyntax {
		http.Error(w, "Invalid ID", 400)
	} else {
		handleUnexpectedError(err, w)
	}
	return -1, err
}

func (app Application) GetElection(w http.ResponseWriter, r *http.Request) {
	id, err := getIntURLParameter(w, r, "id")
	if err != nil {
		return
	}
	election, err := app.database.GetElection(id)
	if err == sql.ErrNoRows {
		http.Error(w, "Not found", 404)
		return
	} else if handleUnexpectedError(err, w) {
		return
	}
	output, err := json.Marshal(election)
	if handleUnexpectedError(err, w) {
		return
	}
	w.Write(output)
}

func (app Application) ListElections(w http.ResponseWriter, r *http.Request) {
	elections, err := app.database.ListElections()
	if handleUnexpectedError(err, w) {
		return
	}
	output, err := json.Marshal(elections)
	if handleUnexpectedError(err, w) {
		return
	}
	w.Write(output)
}

func (app Application) AddCandidate(w http.ResponseWriter, r *http.Request) {
	id, err := getIntURLParameter(w, r, "election_id")
	if err != nil {
		return
	}
	election, err := app.database.GetElection(id)
	if err == sql.ErrNoRows {
		http.Error(w, "Not found", 404)
		return
	}
	candidate := Candidate{}
	err = json.NewDecoder(r.Body).Decode(&candidate)
	if handleUnexpectedError(err, w) {
		return
	}
	candidate.ElectionId = election.Id
	err = app.database.AddCandidate(candidate)
	w.WriteHeader(201)
}

func init() {
	log = lumber.NewConsoleLogger(lumber.INFO)
	log.Info("Starting up...")
	dbmap := initDb()
	application := CreateApplication(dbmap)
	handler = application.handler
}

func main() {
	log.Info("Listening and serving on port 8123...")
	err := http.ListenAndServe(":8123", handler)
	if err != nil {
		log.Fatal(err.Error())
		os.Exit(1)
	}
	log.Warn("Exiting...")
}
