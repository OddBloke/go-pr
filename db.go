package main

type ElectionDB interface {
	GET(id int) *Election
	Add(e *Election) (int, error)
}

var electionDatabase ElectionDB
