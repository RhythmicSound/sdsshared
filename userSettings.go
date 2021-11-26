package sdsshared

import "strconv"

var (
	//ResourceServiceName is the name of this service as visible to other services.
	//
	//Passing value through to StartServers function overrides this by hard coding.
	// To avoid this pass an empty string into StartServers
	// function or repass this value as sdsshared.ResourceServiceName.
	ResourceServiceName string
	//Whether verbose logging and test data should be used.
	// Never set true for production
	DebugMode bool
	//DBURI is the path -URL or local path- to the database resource.
	// If it doesn't exist one will be created at this resource for
	// some connector implementations
	DBURI string
	//DBURI is the path =URL or local path= to the archive file for
	//the dataset used to rebuild the database
	DatasetURI string
	//PublicPort is the port from which this API can be accessed for data retrieval
	PublicPort string
)

func init() {
	//debug mode into bool
	db, err := strconv.ParseBool(GetEnv("debug", "false"))
	if err != nil {
		DebugMode = false
	}
	DebugMode = db
	//database location setting
	DBURI = GetEnv("database-uri", "databases/dummy")
	//location of the dataset archive
	DatasetURI = GetEnv("dataset-uri", "/datasets")
	//get name of this running resource -
	ResourceServiceName = GetEnv("name", "Default Resource Name")
	//get port to use
	PublicPort = GetEnv("publicport", "8080")
}