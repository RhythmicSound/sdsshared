package main

import (
	"fmt"
	"log"

	sdsshared "github.com/RhythmicSound/sds-shared"
	badgerconnector "github.com/RhythmicSound/sds-shared/badgerConnector"
)

func main() {

	connector := badgerconnector.New(sdsshared.ResourceServiceName, sdsshared.DatasetURI, true)

	log.Fatalln(sdsshared.StartServer(connector, fmt.Sprintf("Dummy %s Server", sdsshared.ResourceServiceName), 8080))
}
