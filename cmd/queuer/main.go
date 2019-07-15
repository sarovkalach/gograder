package main

import (
	"github.com/sarovkalach/gograder/pkg/queuer"
)

func main() {
	server := queuer.NewQueuer()
	server.Run()
}
