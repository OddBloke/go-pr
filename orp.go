package main

import "github.com/coopernurse/gorp"

type DB interface {
	Add(v interface{}) error

	GetElection(id int) (Election, error)
	ListElections() ([]Election, error)

	GetCandidate(electionId, id int) (Candidate, error)
}

var electionDatabase DB

type GorpDB struct {
	dbmap *gorp.DbMap
}

func (db GorpDB) Add(v interface{}) error {
	err := db.dbmap.Insert(v)
	return err
}

func (db GorpDB) GetElection(id int) (Election, error) {
	var election Election
	err := db.dbmap.SelectOne(&election, "SELECT * FROM elections WHERE id=?", id)
	return election, err
}

func (db GorpDB) ListElections() ([]Election, error) {
	var elections []Election
	_, err := db.dbmap.Select(&elections, "SELECT * FROM elections")
	return elections, err
}

func (db GorpDB) GetCandidate(electionId, id int) (Candidate, error) {
	var candidate Candidate
	err := db.dbmap.SelectOne(&candidate, "SELECT * FROM candidates WHERE Id=? AND ElectionId=?", id, electionId)
	return candidate, err
}
