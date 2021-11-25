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

Implement `DataResource` interface

Example:
```go
type impl struct{}

repoURIAddress := "https://repo.com/versionedblobstore/blo.zip"

func NewImpl(datasetArchiveLocation string){
  return &impl{}
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
}
```

```go
(im impl) Retrieve(someSearchString)([]byte,error){
   ... //Fetching data from the database based on given key value. Returns SimpleData struct in json binary format
}
```

```go

(im impl) Shutdown()error{
  ... //Shutdown scripts
}
```

Run the server
```go
i := NewImpl()

... //Any other logic that may be required

log.Fatalln(sdsshared.StartServer(i))
```