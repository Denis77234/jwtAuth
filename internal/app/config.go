package app

import (
	"log"
	"os"
)

type config struct {
	port       string
	mongoUri   string
	accessKey  string
	refreshKey string
}

func newCfg() config {
	port, ok := os.LookupEnv("SERVER_PORT")
	if !ok {
		log.Fatal("env variable SERVER_PORT not found")
	}

	monUri, ok := os.LookupEnv("MONGO_URI")
	if !ok {
		log.Fatal("env variable MONGO_URI not found")
	}

	accKey, ok := os.LookupEnv("ACCESS_SECRET")
	if !ok {
		log.Fatal("env variable ACCESS_SECRET not found")
	}

	refKey, ok := os.LookupEnv("REFRESH_SECRET")
	if !ok {
		log.Fatal("env variable REFRESH_SECRET not found")
	}

	return config{
		port:       port,
		mongoUri:   monUri,
		accessKey:  accKey,
		refreshKey: refKey,
	}
}
