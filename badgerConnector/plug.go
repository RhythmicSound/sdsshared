package badgerconnector

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	sdsshared "github.com/RhythmicSound/sdsshared"
	badger "github.com/dgraph-io/badger/v3"
)

//Palawan (a stinky Badger specices) is the main api implementer for the Badger KV database
type Palawan struct {
	ResourceName         string
	Database             *badger.DB
	transitionalDatabase *badger.DB // to put a db whilst doing update database switchover
	updateCount          int        //number of times UpdateDataset method
	versioner            sdsshared.VersionManager
	mu                   *sync.Mutex
	predictiveMode       bool //whether or not the retieve term should be considered the full search term (false) or an incomplete typed term (true)
}

//New creates a new BadgerDB Palawan instance that implements DataResource
func New(resourceName, datasetDownloadLoc string, predictiveMode bool) *Palawan {

	return &Palawan{
		ResourceName:   resourceName,
		predictiveMode: predictiveMode,
		mu:             &sync.Mutex{},
		updateCount:    0,
		versioner: sdsshared.VersionManager{
			Repo:           datasetDownloadLoc,
			LastUpdated:    "",
			CurrentVersion: "0",
			DataSources:    make([]string, 0),
		},
	}
}

// Open the Badger database located in the databaseLocation directory.
// It will be created if it doesn't exist.
func (pal Palawan) Open(databaseLocation string) (*badger.DB, error) {
	options := badger.DefaultOptions(databaseLocation)
	options = options.WithInMemory(false)

	db, err := badger.Open(options)
	if err != nil {
		return nil, err
	}
	return db, nil
}

//Close closes the database. Must be done prior to closing the application
func (pal *Palawan) Close() error {
	return pal.Database.Close()
}

//Startup script function prior to receiving data access requests
func (pal *Palawan) Startup() error {
	if db, err := pal.Open(fmt.Sprintf("%s%d", sdsshared.DBURI, pal.updateCount)); err != nil {
		return fmt.Errorf("Error opening database in badgerConnector.Startup(): %v", err)
	} else {
		pal.Database = db
	}

	//download and deploy dataset to database and run as datasource
	if !sdsshared.DebugMode {
		if err := pal.fetchDataset(pal.versioner.Repo, true); err != nil {
			return fmt.Errorf("Error fetching dataset in badgerConnector.Startup(): %v", err)
		}
		if _, err := pal.loadDataset(nil); err != nil {
			return fmt.Errorf("Error loading dataset in badgerConnector.Startup(): %v", err)
		}
	}

	//turn on debug if needed. If so add test data
	if sdsshared.DebugMode {
		//?TESTING AND DEBUG-----------------------------------
		if err := pal.AddTestData(20); err != nil {
			return fmt.Errorf("Error adding test data to database in badgerConnector.Startup(): %v", err)
		} //? TESTING END --------------------------------------
	}

	//Get versioner meta info from downloaded database
	if err := pal.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("_version"))
		if err != nil {
			return fmt.Errorf("Error getting version data from loaded database in badgerConnector.Startup(). Must contain key '_version': %v", err)
		}
		if err := item.Value(func(val []byte) error {
			vs := &sdsshared.VersionManager{}
			if err := json.Unmarshal(val, vs); err != nil {
				return fmt.Errorf("Error unmarshalling version data in badgerConnector.Startup(): %v", err)
			}
			pal.versioner = *vs
			return nil
		}); err != nil {
			return err
		}
		//Print top 5 items to logout for debug and visual accuracy check
		top := 5
		opt := badger.DefaultIteratorOptions
		opt.PrefetchValues = true
		opt.PrefetchSize = top
		it := txn.NewIterator(opt)
		defer it.Close()
		it.Rewind()
		for i := 0; i < top && it.Valid(); i += 1 {
			item := it.Item()
			item.Value(func(val []byte) error {
				vl := make(map[string]string)
				if err := json.Unmarshal(val, &vl); err != nil {
					return err
				}
				fmt.Printf("%s -> %+v\n", string(item.Key()), vl)
				return nil
			})
			it.Next()
		} //Top complete
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
		Meta: sdsshared.Meta{
			LastUpdated: pal.versioner.LastUpdated,
			DataSources: pal.versioner.DataSources,
			Resource:    pal.ResourceName,
		}, RequestOptions: options,
	}
	//normalise to all uppercase keys
	toFind = strings.ToUpper(toFind)
	//seperator used in CreateKVStoreKey function
	keySeperator := "/"
	//standardise and optimise for time sorting
	value := make(map[string]interface{}, 0)

	err := pal.Database.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		if pal.predictiveMode {
			opts.PrefetchValues = false
		}
		it := txn.NewIterator(opts)
		defer it.Close()
		prefix := []byte(toFind + keySeperator)
		if pal.predictiveMode {
			prefix = []byte(toFind)
		}

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			keyComposite := strings.Split(string(item.Key()), keySeperator)
			//If predictiveMode is on, only a list of matching keys are required without timestamp
			if pal.predictiveMode {
				if _, ok := value[keyComposite[0]]; ok {
					value[keyComposite[0]] = true
				} else {
					value[keyComposite[0]] = keyComposite[1]
				}
				continue
			}
			err := item.Value(func(val []byte) error {
				// This func with val would only be called if item.Value encounters no error.
				timestamp := keyComposite[1]
				jvo := make(map[string]interface{})
				if err := json.Unmarshal(val, &jvo); err != nil {
					return err
				}
				value[timestamp] = jvo
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
func (pal *Palawan) UpdateDataset() (sdsshared.VersionManager, error) {
	//Open new blank db
	pal.updateCount += 1
	db, err := pal.Open(fmt.Sprintf("%s%d", sdsshared.DBURI, pal.updateCount))
	if err != nil {
		return sdsshared.VersionManager{}, err
	}
	//Download new data
	if err := pal.fetchDataset(pal.versioner.Repo, true); err != nil {
		return sdsshared.VersionManager{}, err
	}
	//Load in new data
	if _, err := pal.loadDataset(db); err != nil {
		return sdsshared.VersionManager{}, err
	}
	//Make new db the in use pal.Database
	if err = pal.mount(db); err != nil {
		return sdsshared.VersionManager{}, err
	}

	return pal.versioner, nil
}

//AddTestData adds [num] items of randomised test data to the database
func (pal *Palawan) AddTestData(num int) error {
	if err := pal.Database.Update(func(txn *badger.Txn) error {
		version := sdsshared.VersionManager{
			CurrentVersion: "1.0.0",
			LastUpdated:    time.Now().Format(time.RFC3339),
			DataSources:    []string{"Dummy data warehouse"},
		}
		versionjson, err := json.Marshal(version)
		if err != nil {
			return err
		}

		//add meta data ---
		etryVersion := badger.NewEntry([]byte("_version"), []byte(versionjson))
		if err := txn.SetEntry(etryVersion); err != nil {
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

//fetchDataset downloads the dataset archive from given location to the local downloads location
//
//Mark gcp as true if downloading from a private GCP bucket. Requires GCP Authentication
func (pal Palawan) fetchDataset(datasetURL string, gcp bool) error {

	//If downloading from GCP cloud storage that requires authentication
	if gcp {
		return sdsshared.GCPDownload(sdsshared.DatasetBucketName, sdsshared.DatasetObjectName)
	}

	//Else download from URL---
	if datasetURL == "" {
		datasetURL = pal.versioner.Repo
	}
	//Open download location dir
	if err := os.MkdirAll(sdsshared.LocalDownloadDir, 0755); err != nil {
		return err
	}
	//Download
	client := sdsshared.NewHTTPClient()
	resp, err := client.Get(datasetURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	//create file in download folder
	file, err := os.Create(path.Join(sdsshared.LocalDownloadDir, "datasetupdate.zip"))
	defer file.Close()
	if err != nil {
		return err
	}
	io.Copy(file, resp.Body)

	return nil
}

//loadDataset loads a dataset from a zip archive containing .bak files to an open badgerdb instance
//
//If dbToLoad is nil, it loads the data directly into the pal.Database instance
func (pal *Palawan) loadDataset(dbToLoad *badger.DB) (*badger.DB, error) {
	fileLoc := path.Join(sdsshared.LocalDownloadDir, "datasetupdate.zip")
	lockFirst := false
	//get usable target database
	if dbToLoad == nil {
		dbToLoad = pal.Database
		lockFirst = true
	}
	//open zip
	zipR, err := zip.OpenReader(fileLoc)
	if err != nil {
		return nil, fmt.Errorf("Error. Could not get zip reader in badgerConnector.loadDataset(): %v", err)
	}
	//load backup file(s)
	files := zipR.File
	if lockFirst {
		pal.mu.Lock()
	}
	for _, file := range files {
		if path.Ext(file.Name) == ".bak" {
			f, err := file.Open()
			if err != nil {
				return nil, err
			}
			if err = dbToLoad.Load(f, 10); err != nil {
				return nil, fmt.Errorf("Error running dbToLoad.Load() in loadDataset within BadgetConnerctor: %v", err)
			}
			if err = f.Close(); err != nil {
				return nil, err
			}
		}
	}
	if lockFirst {
		pal.versioner, err = deriveVersioner(dbToLoad)
		if err != nil {
			return nil, fmt.Errorf("Error could not deriver versioner in badgerconnect.loadDataset(): %v", err)
		}
		pal.mu.Unlock()
	}
	if err := zipR.Close(); err != nil {
		return nil, err
	}
	//cleanup downloads
	if err := os.Remove(fileLoc); err != nil {
		return nil, err
	}

	return dbToLoad, nil
}

//mount loads the given badger DB instance to the Palwan instance and closes the existing instance.
//
//this uses the transitional db field as a crossover point.
func (pal *Palawan) mount(dbToMount *badger.DB) error {
	var err error
	pal.mu.Lock()
	pal.transitionalDatabase = pal.Database
	pal.Database = dbToMount
	//Update pal.versioner
	pal.versioner, err = deriveVersioner(dbToMount)
	if err != nil {
		return err
	}
	pal.mu.Unlock()
	//close old db
	err = pal.transitionalDatabase.Close()
	if err != nil {
		return err
	}
	//todo ...

	return nil
}

//deriveVersioner creates the Versioner based on the meta fields of the database
func deriveVersioner(db *badger.DB) (sdsshared.VersionManager, error) {
	vs := sdsshared.VersionManager{}
	//Update pal.versioner
	if err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("_version"))
		if err != nil {
			return err
		}
		item.Value(func(val []byte) error {
			if err := json.Unmarshal(val, &vs); err != nil {
				return err
			}
			return nil
		})

		return nil
	}); err != nil {
		return sdsshared.VersionManager{}, err
	}
	return vs, nil
}
