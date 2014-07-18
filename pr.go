package main

import (
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/stephens2424/muxchain"
	"github.com/stephens2424/muxchain/muxchainutil"
)

func AddElection(w http.ResponseWriter, r *http.Request) {
	err := electionDatabase.Add(Election{Name: "foo"})
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func GetElection(w http.ResponseWriter, r *http.Request) {
	id_string := mux.Vars(r)["id"]
	id, err := strconv.Atoi(id_string)
	if err == strconv.ErrSyntax {
		http.Error(w, "Invalid ID", 400)
		return
	}
	election := electionDatabase.Get(id)
	io.WriteString(w, election.String())
}

func init() {
	router := mux.NewRouter()
	router.HandleFunc("/elections", AddElection).Methods("POST")
	router.HandleFunc("/elections/{id}", GetElection).Methods("GET")
	muxchain.Chain("/", muxchainutil.DefaultLog, router)
}

func main() {
	err := http.ListenAndServe(":8123", muxchain.Default)
	if err != nil {
		log.Fatal(err)
	}
}
