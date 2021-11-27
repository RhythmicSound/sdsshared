package badgerconnector

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

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

	//download and deploy dataset to database and run as datasource
	if _, err := pal.UpdateDataset(nil); err != nil {
		return err
	}

	//turn on debug if needed. If so add test data
	if sdsshared.DebugMode {
		//?TESTING AND DEBUG-----------------------------------
		if err := pal.AddTestData(20); err != nil {
			return err
		} //? TESTING END --------------------------------------
	}

	//Get versioner meta info from downloaded database
	if err := pal.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("_version"))
		if err != nil {
			return err
		}
		if err := item.Value(func(val []byte) error {
			vs := &sdsshared.VersionManager{}
			if err := json.Unmarshal(val, vs); err != nil {
				return err
			}
			pal.versioner = *vs
			return nil
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
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

	//seperator used in CreateKVStoreKey function
	keySeperator := "/"
	//standardise and optimise for time sorting
	value := make(map[string]string, 0)

	err := pal.Database.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(toFind + keySeperator)

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				// This func with val would only be called if item.Value encounters no error.
				timestamp := strings.Split(string(item.Key()), keySeperator)[1]
				value[timestamp] = string(val)
				return nil
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return sdsshared.SimpleData{}, err
	}
	out.Data.Values = value
	out.ResultCount = len(value)

	return out, nil
}

//UpdateDataset function loads data from source and updates db in use
func (pal *Palawan) UpdateDataset(versionManager *sdsshared.VersionManager) (*sdsshared.VersionManager, error) {

	//todo ...

	return &pal.versioner, nil
}

//AddTestData adds [num] items of randomised test data to the database
func (pal *Palawan) AddTestData(num int) error {
	if err := pal.Database.Update(func(txn *badger.Txn) error {
		version := sdsshared.VersionManager{
			CurrentVersion: 1,
			LastUpdated:    time.Now().Format(time.RFC3339),
			DataSources:    []string{"Dummy data warehouse"},
		}
		versionjson, err := json.Marshal(version)
		if err != nil {
			return err
		}

		etry := badger.NewEntry([]byte("_version"), []byte(versionjson))
		if err := txn.SetEntry(etry); err != nil {
			return err
		}

		for x := 0; x < num; x += 1 {
			e := badger.NewEntry([]byte(sdsshared.CreateKVStoreKey(fmt.Sprintf("TestEntry%d", x), "/")), []byte(fmt.Sprintf("Value%d", rand.Int())))
			if err := txn.SetEntry(e); err != nil {
				if err == badger.ErrTxnTooBig {
					err = txn.Commit()
					if err != nil {
						return err
					}
					txn = pal.Database.NewTransaction(true)
					err = txn.SetEntry(e)
				}
				if err != nil {
					return err
				}
			}
		}

		//Print everything in the database to log
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			err := item.Value(func(v []byte) error {
				fmt.Printf("key=%s, value=%s\n", k, v)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	//GC
	for {
		if err := pal.Database.RunValueLogGC(0.7); err != nil {
			break
		}

	}

	return nil
}
