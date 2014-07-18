package main

import (
	"io"
	"log"
	"net/http"
	"strconv"

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
	dbmap.AddTableWithName(Election{}, "elections").SetKeys(true, "Id")

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
	n, err := r.Body.Read(requestBytes)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	name := string(requestBytes[:n])
	err = app.electionDatabase.Add(Election{Name: name})
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func (app Application) GetElection(w http.ResponseWriter, r *http.Request) {
	idString := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idString)
	if err == strconv.ErrSyntax {
		http.Error(w, "Invalid ID", 400)
		return
	}
	election := app.electionDatabase.Get(id)
	io.WriteString(w, election.String())
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
