package main

import (
	"github.com/jemgunay/canlendar-graph"
	"io/ioutil"
	"log"
)

func main() {
	b, err := ioutil.ReadFile("../../config/credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	canlendar.Start(b)
}
