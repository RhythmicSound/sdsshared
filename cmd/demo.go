package main

import (
	"fmt"
	"log"

	sdsshared "github.com/RhythmicSound/sdsshared"
	badgerconnector "github.com/RhythmicSound/sdsshared/badgerConnector"
	cockroarchconnector "github.com/RhythmicSound/sdsshared/cockroachConnector"
)

/**********************************************************
This demo shows how to launch CockroachDB and BadgerDB backed
data services using the standard package Connectors

This is used at MoreCoffeePlease to modularly scaffold and
deliver API servers to provison our services at scale.
It is intended to be used as the front end api
provision service using data stores prepared with sister
data preparation packages.

For external purposes this can be used as an entry point to
discover how to use sds-shared as a library.

The below example uses the UK Postcode dataset as an example
and provisions 2 servers to serve the same data; one from
a cockroach server and the other from an embedded badger database
after fetching the db files from a cloud location.
**********************************************************/

func main() {

	//Demo CockroachConnector
	connectorCockroach, err := cockroarchconnector.New(sdsshared.DBURI, sdsshared.CACertURI, "geography", "public", "postcodesuk", "postcode")
	if err != nil {
		log.Panicln(err)
	}
	go func() {
		log.Fatalln(sdsshared.StartServer(&connectorCockroach, fmt.Sprintf("Demo Cockroach %s Server", sdsshared.ResourceServiceName), 8081))
	}()

	//Demo badgerConnector
	connector := badgerconnector.New(sdsshared.ResourceServiceName, sdsshared.DatasetURI, false)

	log.Fatalln(sdsshared.StartServer(connector, fmt.Sprintf("Demo Badger %s Server", sdsshared.ResourceServiceName), 8080))
}
