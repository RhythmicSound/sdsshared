package sdsshared

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"
)

//Redirect included to ensure http requests are forwarded to the Https endpoint - ref https://gist.github.com/d-schmidt/587ceec34ce1334a5e60
func redirect(w http.ResponseWriter, req *http.Request) {
	// remove/add not default ports from req.Host
	target := "https://" + req.Host + req.URL.Path
	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}
	log.Printf("redirect to: %s", target)
	http.Redirect(w, req, target,
		//consider the codes 308, 302, or 301. 307 used as also forwards req body
		http.StatusTemporaryRedirect)
}

//StartServer runs the server to interface with the system using the api methods of DataResource
func StartServer(dr DataResource, serverName string, port int) error {
	//set port
	prt := ""
	if port == 0 {
		prt = ":8080"
	} else {
		prt = fmt.Sprintf(":%d", port)
	}

	//set routes
	router := http.NewServeMux()

	router.HandleFunc("/fetch", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Scheme == "http" {
			redirect(w, r)
		}
		term := r.URL.Query().Get("fetch")
		data, err := dr.Retrieve(term) //should already be in SimpleData api format
		if err != nil {
			log.Printf("Error. Could not retrieve data from data resource: %v", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		fmt.Fprint(w, string(data))
	})

	router.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Scheme == "http" {
			redirect(w, r)
		}

	})

	//build server
	server := &http.Server{
		Addr:              prt,
		Handler:           router,
		ReadTimeout:       2 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
		IdleTimeout:       2 * time.Second,
		TLSConfig: &tls.Config{
			ServerName: serverName,
			MinVersion: tls.TLS_AES_128_GCM_SHA256,
		},
	}

	if server.TLSConfig.ServerName == "" {
		server.TLSConfig.ServerName = "Simple-Data-Service Default Name"
	}

	//run startup scripts in the data resource
	if err := dr.Startup(); err != nil {
		return fmt.Errorf("Could not run data resource startup scripts before server launch: %+v", err)
	}

	//Ensure shutdown scripts are run
	defer dr.Shutdown()

	//run server
	return fmt.Errorf("Could not launch server: %+v", server.ListenAndServe())
}

func updateDataset(datasetLink string) {}
