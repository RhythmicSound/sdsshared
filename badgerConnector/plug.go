package badgerconnector

import (
	"fmt"
	"math/rand"

	sdsshared "github.com/RhythmicSound/sds-shared"
	badger "github.com/dgraph-io/badger/v3"
)

//Palawan (a stinky Badger specices) is the main api implementer for the Badger KV database
type Palawan struct {
	ResourceName string
	Database     *badger.DB
	versioner    sdsshared.VersionManager
}

//New creates a new BadgerDB Palawan instance that implements DataResource
func New(resourceName, datasetDownloadLoc string) *Palawan {

	return &Palawan{
		ResourceName: resourceName,
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
	if _, err := pal.Open(sdsshared.DBURI); err != nil {
		return err
	}

	if _, err := pal.UpdateDataset(nil); err != nil {
		return err
	}

	if sdsshared.DebugMode {
		//?TESTING AND DEBUG-----------------------------------
		if err := pal.AddTestData(20); err != nil {
			return err
		} //? TESTING END --------------------------------------
	}

	return nil
}

//Shutdown funs any necarssary shutdown scripts prior to application close
func (pal *Palawan) Shutdown() error {
	return pal.Close()
}

//Retrieve is run each time the server receives a search term to query the db for
func (pal *Palawan) Retrieve(toFind string, options map[string]string) (sdsshared.SimpleData, error) {
	out := sdsshared.SimpleData{
		Meta: struct {
			Resource    string   "json:\"resource\""
			LastUpdated string   "json:\"dataset_updated\""
			DataSources []string "json:\"data_sources\""
		}{
			LastUpdated: pal.versioner.LastUpdated,
			DataSources: pal.versioner.DataSources,
			Resource:    pal.ResourceName,
		}, RequestOptions: options,
	}

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
		return sdsshared.SimpleData{}, err
	}

	return out, nil
}

//UpdateDataset function loads data from source and updates db in use
func (pal *Palawan) UpdateDataset(versionManager *sdsshared.VersionManager) (*sdsshared.VersionManager, error) {

	//todo ...

	return &pal.versioner, nil
}

//AddTestData adds [num] items of randomised test data to the database
func (pal *Palawan) AddTestData(num int) error {
	for x := 0; x < num; x += 1 {
		err := pal.Database.Update(func(txn *badger.Txn) error {
			e := badger.NewEntry([]byte(fmt.Sprintf("TestEntry%d", x)), []byte(fmt.Sprintf("Value%d", rand.Int())))
			err := txn.SetEntry(e)
			return err
		})

		if err != nil {
			return err
		}
	}
	return nil
}
