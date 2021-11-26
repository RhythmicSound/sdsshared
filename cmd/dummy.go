package main

import (
	"log"

	sdsshared "github.com/RhythmicSound/sds-shared"
	badgerconnector "github.com/RhythmicSound/sds-shared/badgerConnector"
)

func main() {
	repoURIAddress := "https://repo.com/versionedblobstore/blo.zip"

	connector := badgerconnector.New("Dummy Data Server", repoURIAddress)

	log.Fatalln(sdsshared.StartServer(connector, "Dummy", 8080))
}
