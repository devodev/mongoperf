package main

import (
	"context"
	"flag"

	"github.com/sirupsen/logrus"
)

func main() {
	var uri string
	flag.StringVar(&uri, "uri", "", "mongodb URI connection string")
	flag.Parse()

	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	logger := logrus.New()
	client := Client{uri, logger}

	if err := client.RunDemo(context.TODO(), "test", "trainers"); err != nil {
		logger.Fatal(err)
	}
}
