package main

import "github.com/coopernurse/gorp"

type ElectionDB interface {
	Add(e Election) error
	Get(id int) (Election, error)
	List() ([]Election, error)
}

var electionDatabase ElectionDB

type GorpElectionDB struct {
	dbmap *gorp.DbMap
}

func (db GorpElectionDB) Add(e Election) error {
	err := db.dbmap.Insert(&e)
	return err
}

func (db GorpElectionDB) Get(id int) (Election, error) {
	var election Election
	err := db.dbmap.SelectOne(&election, "SELECT * FROM elections WHERE id=?", id)
	return election, err
}

func (db GorpElectionDB) List() ([]Election, error) {
	var elections []Election
	_, err := db.dbmap.Select(&elections, "SELECT * FROM elections")
	return elections, err
}
