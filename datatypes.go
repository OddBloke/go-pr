package main

type Election struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func (e *Election) String() string {
	return e.Name
}

type Candidate struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	ElectionId int    `json:"election_id"`
}
