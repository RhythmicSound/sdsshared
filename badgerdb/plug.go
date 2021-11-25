package badgerdb

import (
	sdsshared "github.com/RhythmicSound/sds-shared"
	badger "github.com/dgraph-io/badger/v3"
)

//Palawan (a stinky Badger specices) is the main api implementer for the Badger KV database
type Palawan struct {
	Database  *badger.DB
	versioner sdsshared.VersionManager
}

//New creates a new BadgerDB Palawan instance that implements DataResource
func New(datasetDownloadLoc string) *Palawan {

	return &Palawan{
		versioner: sdsshared.VersionManager{
			Repo:           datasetDownloadLoc,
			LastUpdated:    "",
			CurrentVersion: 0,
			DataSources:    make([]string, 0),
		},
	}
}

// Open the Badger database located in the databaseLocation directory.
// It will be created if it doesn't exist.
func (pal *Palawan) Open(databaseLocation string) (*badger.DB, error) {
	options := badger.DefaultOptions(databaseLocation)
	options = options.WithInMemory(false)

	db, err := badger.Open(options)
	if err != nil {
		return nil, err
	}
	pal.Database = db
	return db, nil
}

//Close closes the database. Must be done prior to closing the application
func (pal *Palawan) Close() error {
	return pal.Database.Close()
}

//Startup script function prior to receiving data access requests
func (pal *Palawan) Startup() error {
	dbDirLocation := "databases/dummy"

	if _, err := pal.Open(dbDirLocation); err != nil {
		return err
	}
	return nil
}

//Shutdown funs any necarssary shutdown scripts prior to application close
func (pal *Palawan) Shutdown() error {
	return pal.Close()
}

//Retrieve is run each time the server receives a search term to query the db for
func (pal *Palawan) Retrieve(toFind string) ([]byte, error) {
	//standardise and optimise for time sorting
	toFind = sdsshared.CreateKVStoreKey(toFind, "/")
	value := make([]byte, 0)

	err := pal.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(toFind))
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			// This func with val would only be called if item.Value encounters no error.
			value = append(value, val...)
			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return value, nil
}

//UpdateDataset function loads data from source and updates db in use
func (pal *Palawan) UpdateDataset(versionManager *sdsshared.VersionManager) (*sdsshared.VersionManager, error) {

	//todo ...

	return &pal.versioner, nil
}
