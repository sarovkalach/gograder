package main

import (
	"github.com/sarovkalach/gograder/pkg/loader"
)

func main() {
	server := loader.NewServer()
	server.Run()
}
