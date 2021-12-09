package sdsshared

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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

//StartServer runs the server to interface with the system using the api methods of DataResource.
//
//If port is set to 0 a default setting or the setting given using environment variable
// `publicport` will be used
//
//Similarly if serverName is not set the default will be used or the value in environment
// variable `name` suffixed with the word 'server'
func StartServer(dr DataResource, serverName string, port int) error {
	//set port
	prt := ""
	if port == 0 {
		prt = fmt.Sprintf(":%s", PublicPort)
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

		args := make(map[string]string)
		for k, v := range r.URL.Query() {
			args[k] = strings.Join(v, ",")
		}
		data, err := dr.Retrieve(term, args)
		if err != nil {
			log.Printf("Error. Could not retrieve data from data resource: %v", err)
			errMsgPayload, err := returnErrorJSON("Dataset fetch error", http.StatusInternalServerError, err.Error())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			fmt.Fprint(w, errMsgPayload)
			return
		}
		dataJSON, err := json.Marshal(data)
		if err != nil {
			log.Printf("Error marshalling returned SimpleData struct: %v", err)
			errMsgPayload, err := returnErrorJSON("Marshaling results error", http.StatusInternalServerError, err.Error())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			fmt.Fprint(w, errMsgPayload)
			return
		}
		w.Header().Add("content-type", "application/json")
		fmt.Fprint(w, string(dataJSON))
	})

	router.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Scheme == "http" {
			redirect(w, r)
		}

		newVersionInfo, err := dr.UpdateDataset()
		if err != nil {
			log.Printf("Error updating dataset: %v", err)
			errMsgPayload, err := returnErrorJSON("Dataset update error", http.StatusInternalServerError, err.Error())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			fmt.Fprint(w, errMsgPayload)
			return
		}
		versionDataJSON, err := json.MarshalIndent(newVersionInfo, " ", " ")
		if err != nil {
			log.Printf("Error marshalling returned VersionData struct: %v", err)
			errMsgPayload, err := returnErrorJSON("Dataset update error", http.StatusInternalServerError, err.Error())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			fmt.Fprint(w, errMsgPayload)
			return
		}
		w.Header().Add("content-type", "application/json")
		fmt.Fprint(w, string(versionDataJSON))
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
		server.TLSConfig.ServerName = fmt.Sprintf("%s Server", ResourceServiceName)
	}

	//run startup scripts in the data resource
	if err := dr.Startup(); err != nil {
		return fmt.Errorf("Could not run data resource startup scripts before server launch: %+v", err)
	}

	//Ensure shutdown scripts are run
	defer dr.Shutdown()

	//run server
	log.Printf("Running server on port %d\n", port)
	return fmt.Errorf("Could not launch server: %+v", server.ListenAndServe())
}
