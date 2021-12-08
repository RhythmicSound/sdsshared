# Simple Data Service - shared library

This package presents code to be used by each data resource in order to have a standard API used by every service.

## Code Organisation

The following approach is used: 

1. Application domain types go in the root - `api.go`, `server.go`, `util.go`, etc
1. Implementation of the application domain go in subpackages - `badger`, `sqlite`, etc
1. Everything is tied together in the `cmd` subpackages - `cmd/shared`

There is **no** `internal` package here as there need be no hiding of code from implementation code.

## Usage

Add to project
```go
go get github.com/rhythmicsound/sdsshared
 ```

Then you can create a new service by picking your connector and running something like: 
```go
package main

import (
	"fmt"
	"log"

	sdsshared "github.com/RhythmicSound/sds-shared"
	badgerconnector "github.com/RhythmicSound/sds-shared/badgerConnector"
)

func main() {

	connector := badgerconnector.New(sdsshared.ResourceServiceName, sdsshared.DatasetURI)

	log.Fatalln(sdsshared.StartServer(connector, "", 0))
}
```
Using default type values for the arguments to StartServer allows service name and ports to be set using environment variables at runtime.

An example execute command is: 
```go
debug=false name="postcodeUK-Service" database_uri="working/databases/postcodeUKdb" dataset_uri="https://storage.cloud.google.com/simple-data-service/datasets/postcodesUK.zip" objectname="datasets/postcodesUK.zip" go run cmd/dummy.go
```

## Settings
Settings for services created from this library can be hardcoded or set using environment variables

|Variable|Explanation|Default|
|-|-|-|
|`debug`|Whether to print verbose output to log and load test data to database. Not for use in production| "false" |
|`database_uri`| The path -URL or local path- to the database resource to connect to.| "working/databases/simpledataservice-default/" (N.B. this points at a directory as BadgerDB is the default db in use. This could be a URL or path to local file. In some instances, if no db exists in the path given, one could be created.) |
|`dataset_uri`|The path -URL or local path- to the dataset resource used to rebuild the database.|"working/datasets/data.zip"|
|`bucket`|The cloud bucket from which to find the dataset archive. (Required only if downloading the dataset from behind an authentication wall)|"simple-data-service"|
|`object`|The cloud object name found in DatasetBucketName that identifies the dataset archive for download. (Required only if downloading the dataset from behind an authentication wall)|-|
|`name`|The name of this service as visible to other services.|"Default Resource Name"|
|`publicport`|PublicPort is the port from which this API can be accessed for data retrieval|"8080"|
|`downloaddir`|The local path where download files will be saved to|"working/downloads"|

## Writing new backend storage connectors
Implement `DataResource` interface

Example:
```go
type impl struct{
	ResourceName string
	Database     *badger.DB
	versioner    sdsshared.VersionManager
}

func NewImpl(resourceName, datasetArchiveLocation string) *impl{
  return &impl{
    	ResourceName: resourceName,
		versioner: sdsshared.VersionManager{
			Repo:           sdsshared.DatasetURI,
			LastUpdated:    "",
			CurrentVersion: 0,
			DataSources:    make([]string, 0),
		},
  }
}
```

```go

(im impl) Startup()error{
  ... //Startup scripts
}
```

```go
vt := &VersionManager{
    Repo = repoURIAddress
}

(im impl) UpdateDataset(vt)(*VersionManager,error){
  ... // Logic to keep the database synced with a master versioned dataset archive somehwere

    //Passing a VersionManager as an arg should overwrite internal VM created in New. 
    //Must accept nil to use default
}
```

```go
(im impl) Retrieve(someSearchString)(sdsshared.SimpleData,error){
   ... //Fetching data from the database based on given key value. 
    //Use the latest VersionManager to complete the Meta elements of SimpleData response object
}
```

```go

(im impl) Shutdown()error{
  ... //Shutdown scripts
}
```

> See the badgerConnector package for best practise