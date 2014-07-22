package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/coopernurse/gorp"
	"github.com/gorilla/mux"
	"github.com/stephens2424/muxchain"
	"github.com/stephens2424/muxchain/muxchainutil"
)

type Application struct {
	electionDatabase ElectionDB
	router           *mux.Router
}

func (app *Application) configureORM(dbmap *gorp.DbMap) {
	dbmap.AddTableWithName(Election{}, "elections").SetKeys(true, "Id").ColMap("Name").SetUnique(true)

	err := dbmap.CreateTablesIfNotExists()
	checkErr(err, "Create tables failed")

	app.electionDatabase = GorpElectionDB{dbmap}
}

func (app *Application) configureRouting() {
	router := mux.NewRouter()
	router.HandleFunc("/elections", app.AddElection).Methods("POST")
	router.HandleFunc("/elections/{id}", app.GetElection).Methods("GET")
	app.router = router
}

func CreateApplication(dbmap *gorp.DbMap) Application {
	app := Application{}
	app.configureORM(dbmap)
	app.configureRouting()
	return app
}

func (app Application) AddElection(w http.ResponseWriter, r *http.Request) {
	requestBytes := make([]byte, r.ContentLength)
	_, err := r.Body.Read(requestBytes)
	if err != nil {
		http.Error(w, err.Error(), 500)
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
		http.Error(w, err.Error(), 500)
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
	} else if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	output, err := json.Marshal(election)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(output)
}

func init() {
	dbmap := initDb()
	application := CreateApplication(dbmap)
	muxchain.Chain("/", muxchainutil.DefaultLog, application.router)
}

func main() {
	err := http.ListenAndServe(":8123", muxchain.Default)
	if err != nil {
		log.Fatal(err)
	}
}
