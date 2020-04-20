package main

import (
	"context"
	"flag"
	"log"
)

func main() {
	// parse flags
	var uri string
	flag.StringVar(&uri, "uri", "", "mongodb URI connection string")
	flag.Parse()

	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	client := Client{uri}
	log.Fatal(client.RunDemo(context.TODO(), "test", "trainers"))
}
