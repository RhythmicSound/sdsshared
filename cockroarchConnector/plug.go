package cockroarchconnector

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path"

	"github.com/RhythmicSound/sdsshared"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

//Blaberus is the central integration object that implements DataResource
type Blaberus struct {
	databaseURL string
	//DBCARoot is the URL of the root cert of the db server.
	// Usually set by collecting the envar into sdsshared.CACertURI
	DBCARoot   string
	Connection *pgxpool.Pool
	Ctx        context.Context
	Prepared   map[string]pgconn.StatementDescription
	Meta       sdsshared.Meta
}

//New creates a new Blaberus core object.
//
//Takes the database URL and the URL of the database server's root CA certificate
// for SSL (pass in sdsshared.DBURI and sdsshared.CACertURI) to dynamically set
// these by standard environment variables.
func New(databaseURL string, caRootURL string) (Blaberus, error) {
	return Blaberus{
		Ctx:         context.Background(),
		databaseURL: databaseURL,
		DBCARoot:    caRootURL,
	}, nil
}

//Startup spins up the database connection and other startup scripts.
// Adds a db connection to the Blaberus object.
func (blab *Blaberus) Startup() error { //todo pull metadata from database-
	//Fatal exit if no database URL given - alert user to fix
	if blab.databaseURL == "" {
		log.Fatalln("No database URL given. Must set environment varable `database_uri`")
	}
	//Download the databases CA Root Certificate
	err := blab.GetCARootCert("certs", "caroot.cert")
	if err != nil {
		return fmt.Errorf("Error downloading the root CA certificate for the database server: %v", err)
	}
	//create server conn
	conn, err := pgxpool.Connect(blab.Ctx, blab.databaseURL)
	if err != nil {
		return fmt.Errorf("Error connecting to db in cockroachConnector.Startup: %v", err)
	}
	blab.Connection = conn
	return nil
}

//UpdateDataset does not do anything in this instance and exists only to implement DataResource
func (blab Blaberus) UpdateDataset() (sdsshared.VersionManager, error) {
	return sdsshared.VersionManager{}, nil
}

//Retrieve fetches data from the database
func (blab Blaberus) Retrieve(searchQuery string, queryMap map[string]string) (sdsshared.SimpleData, error) {
	var err error
	var rows pgx.Rows
	//Get search type from queryMap expected entry
	switch queryMap["type"] {
	case "condition":
		rows, err = blab.Connection.Query(blab.Ctx, "SQL string", "args...")
	default:
		rows, err = blab.Connection.Query(blab.Ctx, "SQL string", "args...")
	}
	defer rows.Close()

	if err != nil {
		return sdsshared.SimpleData{}, fmt.Errorf("Error fetching data from database: %v", err)
	}

	//process results into sdsshared.SimpleData struct ---
	out := sdsshared.SimpleData{
		RequestOptions: queryMap,
		Meta:           blab.Meta, //gathered in startup
	}
	outData := make([]map[string]interface{}, 0)
	headers := rows.FieldDescriptions()
	for rows.Next() {
		//Add another result count to the output
		out.ResultCount += 1
		//get field value list from row
		values, err := rows.Values()
		if err != nil {
			return sdsshared.SimpleData{}, fmt.Errorf("Error getting values from db rows in Retrieve within cockroachConnector in Retrieve: %v", err)
		}
		//Zip field values with headers
		collect := make(map[string]interface{})
		for i, header := range headers {
			collect[string(header.Name)] = values[i]
		}
		//add to data list in SimpleData
		outData = append(outData, collect)
	}
	out.Data.Values = outData

	return out, nil
}

//Shutdown runs scripts that gracefully closes the connection and server
func (blab Blaberus) Shutdown() error {
	blab.Connection.Close()
	return nil
}

//GetCARootCert downloads the root CA certificate for the database server to the
// given directory in a file created with the given filename. The database
// URL is updated with the corresponding `sslrootcert` query value
// (see https://www.cockroachlabs.com/docs/cockroachcloud/authentication.html)
func (blab *Blaberus) GetCARootCert(certDirectory, certFileName string) error {
	//Standard var setting
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Error obtaining pwd in cockroachConnector.Startup: %v", err)
	}
	u, err := url.Parse(blab.databaseURL)
	if err != nil {
		return fmt.Errorf("Error parsing database URL in CockroachConnector.Startup: %v", err)
	}
	//easy update vars --------------------------------------------
	localCARootDirPath := path.Join(cwd, certDirectory)
	caRootLocalPath := path.Join(localCARootDirPath, certFileName)
	//easy update vars end -----------------------------------------

	//get cacert for server from url
	if blab.DBCARoot != "" {
		err = os.MkdirAll(localCARootDirPath, 0766)
		res, err := sdsshared.NewHTTPClient().Get(blab.DBCARoot)
		if err != nil {
			return fmt.Errorf("Error downloading DB server CA Root from given URL in CockroachConnector.Startup: %v", err)
		}
		defer res.Body.Close()
		file, err := os.Create(caRootLocalPath)
		if err != nil {
			return fmt.Errorf("Error creating download file in cockroachConnector.Startup: %v", err)
		}
		defer file.Close()
		if _, err := io.Copy(file, res.Body); err != nil {
			return fmt.Errorf("Error copying ca cert from http response body to file in cockroachConnector.Startup: %v", err)
		}
		//Adjust query in db url to match SSL verication modes and CA root cert location
		u.Query().Set("sslmode", "verify-full")
		u.Query().Set("sslrootcert", caRootLocalPath)
		blab.databaseURL = u.String()
	}
	return nil
}
