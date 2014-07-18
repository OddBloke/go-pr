package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/stephens2424/muxchain"
	"github.com/stephens2424/muxchain/muxchainutil"
)

type Election struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func (e *Election) String() string {
	return e.Name
}

func AddElection(w http.ResponseWriter, r *http.Request) {
	electionDatabase.Add(&Election{Name: "foo"})
}

func init() {
	router := mux.NewRouter()
	router.HandleFunc("/elections", AddElection).Methods("POST")
	muxchain.Chain("/", muxchainutil.DefaultLog, router)
}

func main() {
	if err := http.ListenAndServe(":8123", muxchain.Default); err != nil {
		log.Fatal(err)
	}
}
