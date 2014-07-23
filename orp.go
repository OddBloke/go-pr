package main

import "github.com/coopernurse/gorp"

type DB interface {
	AddElection(e Election) error
	GetElection(id int) (Election, error)
	ListElections() ([]Election, error)
}

var electionDatabase DB

type GorpDB struct {
	dbmap *gorp.DbMap
}

func (db GorpDB) AddElection(e Election) error {
	err := db.dbmap.Insert(&e)
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
