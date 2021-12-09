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

### Using the library
You can use the library to quickly create a service from pre prepared datasets using an existing connector, such as:
```go
package main

import (
	"fmt"
	"log"

	sdsshared "github.com/RhythmicSound/sdsshared"
	badgerconnector "github.com/RhythmicSound/sdsshared/badgerConnector"
)

func main() {
    //See the badgerConnector package in this repo for constrictions this places
    // on the way data is prepared for use by this connector
	connector := badgerconnector.New(sdsshared.ResourceServiceName, sdsshared.DatasetURI, false)

    //A standard server will use the server to provide standardised responses to requests
    // on the given port
	log.Fatalln(sdsshared.StartServer(connector, fmt.Sprintf("Dummy %s Server", sdsshared.ResourceServiceName), 8080))
}

```

Using default type values for the arguments to StartServer allows service name and ports to be set using environment variables at runtime.

An example execute command is: 
```go
debug=false \
name="postcodeUK-Service" \
database_uri="working/databases/postcodeUKdb" \
dataset_uri="https://storage.cloud.google.com/simple-data-service/datasets/postcodesUK.zip" \
objectname="datasets/postcodesUK.zip" \
go run cmd/dummy.go
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
//Will implement DataResource
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

## Standard query response
There is a standardised query response from all data requests to the server
The format is that of the `sdsshared.SimpleData` struct

Example: 
From the query 
```markdown
http:/localhost:8080/fetch?fetch=CR05qp
```
The response from the 'postcodeUK-Service' may be:
```json
{
	"result_count": 1,
	"request_options": {
		"fetch": "CR05qp"
	},
	"meta": {
		"resource": "postcodeUK-Service",
		"dataset_updated": "2021-12-09T15:54:22Z",
		"data_sources": [
			"https://osdatahub.os.uk/downloads/open#OPNAME"
		]
	},
	"data": {
		"values": {
			"1639065240533169347": {
				"Admin_county_code": "",
				"Admin_district_code": "E09000008",
				"Admin_ward_code": "E05011475",
				"Country_code": "E92000001",
				"Eastings": "533803",
				"NHS_HA_code": "E18000007",
				"NHS_regional_HA_code": "E19000003",
				"Northings": "165451",
				"Positional_quality_indicator": "10",
				"Postcode": "CR0 5QP"
			}
		}
	}
}
```