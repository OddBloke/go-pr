package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/coopernurse/gorp"
	"github.com/gorilla/mux"
	"github.com/stephens2424/muxchain"
	"github.com/stephens2424/muxchain/muxchainutil"
	"github.com/jcelliott/lumber"
)

var log lumber.Logger

func handleUnexpectedError(err error, w http.ResponseWriter) bool {
	if err != nil {
		http.Error(w, err.Error(), 500)
		return true
	}
	return false
}

type Application struct {
	electionDatabase ElectionDB
	router           *mux.Router
}

func (app *Application) configureORM(dbmap *gorp.DbMap) {
	log.Info("Configuring ORM...")
	dbmap.AddTableWithName(Election{}, "elections").SetKeys(true, "Id").ColMap("Name").SetUnique(true)

	err := dbmap.CreateTablesIfNotExists()
	checkErr(err, "Create tables failed")
	log.Info("Tables successfully created.")

	app.electionDatabase = GorpElectionDB{dbmap}
	log.Info("ORM configured.")
}

func (app *Application) configureRouting() {
	log.Info("Configuring routing...")
	router := mux.NewRouter()
	router.HandleFunc("/elections", app.ListElections).Methods("GET")
	router.HandleFunc("/elections", app.AddElection).Methods("POST")
	router.HandleFunc("/elections/{id}", app.GetElection).Methods("GET")
	app.router = router
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
	requestBytes := make([]byte, r.ContentLength)
	_, err := r.Body.Read(requestBytes)
	if handleUnexpectedError(err, w) {
		return
	}
	election := Election{}
	json.Unmarshal(requestBytes, &election)
	if len(election.Name) == 0 {
		http.Error(w, "Empty name forbidden.", 400)
		return
	}
	err = app.electionDatabase.Add(election)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			http.Error(w, "Name taken.", 400)
			return
		}
		handleUnexpectedError(err, w)
		return
	}
}

func (app Application) GetElection(w http.ResponseWriter, r *http.Request) {
	idString := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idString)
	if err == strconv.ErrSyntax {
		http.Error(w, "Invalid ID", 400)
		return
	}
	election, err := app.electionDatabase.Get(id)
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
	elections, err := app.electionDatabase.List()
	if handleUnexpectedError(err, w) {
		return
	}
	output, err := json.Marshal(elections)
	if handleUnexpectedError(err, w) {
		return
	}
	w.Write(output)
}

func init() {
	log = lumber.NewConsoleLogger(lumber.INFO)
	log.Info("Starting up...")
	dbmap := initDb()
	application := CreateApplication(dbmap)
	muxchain.Chain("/", muxchainutil.DefaultLog, application.router)
}

func main() {
	log.Info("Listening and serving on port 8123...")
	err := http.ListenAndServe(":8123", muxchain.Default)
	if err != nil {
		log.Fatal(err.Error())
		os.Exit(1)
	}
	log.Warn("Exiting...")
}
