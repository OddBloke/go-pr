package main

type NamedEntity interface {
	GetName() string
}

type Election struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func (e *Election) String() string {
	return e.Name
}

func (e *Election) GetName() string {
	return e.Name
}

type Candidate struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	ElectionId int    `json:"election_id"`
}

func (c *Candidate) GetName() string {
	return c.Name
}
