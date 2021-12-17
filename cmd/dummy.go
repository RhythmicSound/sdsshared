package main

import (
	"fmt"
	"log"
	"os"

	sdsshared "github.com/RhythmicSound/sdsshared"
	badgerconnector "github.com/RhythmicSound/sdsshared/badgerConnector"
)

func main() {

	defer os.RemoveAll(sdsshared.LocalDownloadDir)

	connector := badgerconnector.New(sdsshared.ResourceServiceName, sdsshared.DatasetURI, false)

	log.Fatalln(sdsshared.StartServer(connector, fmt.Sprintf("Dummy %s Server", sdsshared.ResourceServiceName), 8080))
}
