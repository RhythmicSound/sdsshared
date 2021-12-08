package sdsshared

import (
	"log"
	"os"
	"strconv"
)

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
	// some connector implementations. For local dbs this may be pre/suf-fixed
	// to allow for dataset updates with minimised downtime
	DBURI string
	//DBURI is the path (URL or local path) to the archive file for
	//the dataset used to rebuild the database
	DatasetURI string
	//PublicPort is the port from which this API can be accessed for data retrieval
	PublicPort string
	//LocalDownloadDir is the local relative or absolute path to a downloads folder to use
	LocalDownloadDir string

	//Cloud blob storage related settings

	//DatasetBucketName is the cloud bucket from which to find the dataset archive
	DatasetBucketName = "simple-data-service"
	//DatasetObjectName is the cloud object name found in DatasetBucketName that
	// identifies the dataset archive for download
	DatasetObjectName string
)

func init() {
	//debug mode into bool
	db, err := strconv.ParseBool(GetEnv("debug", "false"))
	if err != nil {
		DebugMode = false
	}
	DebugMode = db
	//database location setting
	DBURI = GetEnv("database_uri", "working/databases/simpledataservice-default/")
	//location of the dataset archive
	DatasetURI = GetEnv("dataset_uri", "working/datasets/data.zip")
	//get name of this running resource -
	ResourceServiceName = GetEnv("name", "Default Resource Name")
	//get port to use
	PublicPort = GetEnv("publicport", "8080")
	//get download dir to use
	LocalDownloadDir = GetEnv("downloaddir", "working/downloads")

	//GCP Authentication
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		if err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "key/simple-data-service-key.json"); err != nil {
			log.Panicln(err)
		}
	}
	DatasetBucketName = GetEnv("bucket", DatasetBucketName)
	DatasetObjectName = GetEnv("objectname", DatasetObjectName)
}
