package main

import (
	"fmt"
	"log"

	sdsshared "github.com/RhythmicSound/sdsshared"
	badgerconnector "github.com/RhythmicSound/sdsshared/badgerConnector"
)

func main() {

	connector := badgerconnector.New(sdsshared.ResourceServiceName, sdsshared.DatasetURI, true)

	log.Fatalln(sdsshared.StartServer(connector, fmt.Sprintf("Dummy %s Server", sdsshared.ResourceServiceName), 8080))
}
