package main

import (
	"fmt"
	"log"

	sdsshared "github.com/RhythmicSound/sdsshared"
	badgerconnector "github.com/RhythmicSound/sdsshared/badgerConnector"
	cockroarchconnector "github.com/RhythmicSound/sdsshared/cockroarchConnector"
)

func main() {
	//dummy cockroachConnector
	dbURL := "***"
	CAcert := "***"

	connectorCockroach, err := cockroarchconnector.New(dbURL, CAcert, "geography", "public", "postcodesuk", "postcode")
	if err != nil {
		log.Panicln(err)
	}
	go func() {
		log.Fatalln(sdsshared.StartServer(&connectorCockroach, fmt.Sprintf("Dummy Cockroach %s Server", sdsshared.ResourceServiceName), 8081))
	}()

	//dummy badgerConnector
	connector := badgerconnector.New(sdsshared.ResourceServiceName, sdsshared.DatasetURI, false)

	log.Fatalln(sdsshared.StartServer(connector, fmt.Sprintf("Dummy Badger %s Server", sdsshared.ResourceServiceName), 8080))
}
