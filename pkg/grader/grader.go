package grader

import (
	"log"
	"net/http"
)

type Grader struct {
}

type Task struct {
	Name    string `json:"name"`
	User    string `json:"user,omitempty"`
	Timeout int    `json:"timeout"`
}

func NewGrader() *Grader {
	return &Grader{}
}

func (g *Grader) Run() {
	log.Println("Grader Start")
	http.HandleFunc("/", g.ReceiveTask)
	log.Fatal(http.ListenAndServe(":8081", nil))
}
