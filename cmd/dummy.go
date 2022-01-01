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
	dbURL := "postgresql://carl:F3AUqIFta3UYC4Av@free-tier5.gcp-europe-west1.cockroachlabs.cloud:26257/defaultdb?options=--cluster%3Dmore-coffee-please-2473"
	CAcert := "https://cockroachlabs.cloud/clusters/b3088b3c-daf9-4ad9-badc-353ebacbb48b/cert"

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
