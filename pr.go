package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/stephens2424/muxchain"
	"github.com/stephens2424/muxchain/muxchainutil"
)

func AddElection(w http.ResponseWriter, r *http.Request) {
	electionDatabase.Add(&Election{Name: "foo"})
}

func init() {
	router := mux.NewRouter()
	router.HandleFunc("/elections", AddElection).Methods("POST")
	muxchain.Chain("/", muxchainutil.DefaultLog, router)
}

func main() {
	err := http.ListenAndServe(":8123", muxchain.Default)
	if err != nil {
		log.Fatal(err)
	}
}
