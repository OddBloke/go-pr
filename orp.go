package main

import "github.com/coopernurse/gorp"

type ElectionDB interface {
	Get(id int) Election
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

func (db GorpElectionDB) Get(id int) Election {
	var election Election
	db.dbmap.SelectOne(&election, "SELECT * FROM elections WHERE id=?", id)
	return election
}
