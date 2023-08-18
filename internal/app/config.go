package app

import (
	"log"
	"os"
)

type config struct {
	mongoUri   string
	accessKey  string
	refreshKey string
}

func newCfg() config {
	monUri, ok := os.LookupEnv("MONGO_URI")
	if !ok {
		log.Fatal("env variable MONGO_URI not found")
	}

	accKey, ok := os.LookupEnv("ACC_SECRET")
	if !ok {
		log.Fatal("env variable ACC_SECRET not found")
	}

	refKey, ok := os.LookupEnv("REF_SECRET")
	if !ok {
		log.Fatal("env variable REF_SECRET not found")
	}

	return config{
		mongoUri:   monUri,
		accessKey:  accKey,
		refreshKey: refKey,
	}
}
