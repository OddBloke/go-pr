package main

import "github.com/coopernurse/gorp"

type ElectionDB interface {
	Get(id int) (Election, error)
	Add(e Election) error
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
