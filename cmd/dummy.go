package main

import (
	"log"

	sdsshared "github.com/RhythmicSound/sds-shared"
	"github.com/RhythmicSound/sds-shared/badgerdb"
)

func main() {
	repoURIAddress := "https://repo.com/versionedblobstore/blo.zip"

	connector := badgerdb.New(repoURIAddress)

	log.Fatalln(sdsshared.StartServer(connector, "Dummy", 8080))
}
