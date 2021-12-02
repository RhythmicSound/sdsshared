package sdsshared

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

//CreateKVStoreKey creates a standard key for use in kv storage dbs. Appends a Unix
// timestamp so that duplicate entries can be stored seperately, sorted easily, and
// deduped using proper protocols without requiring calls to meta
//
//Using a Unix timestamp means just sorted the keys in place will give time sorted
// list
func CreateKVStoreKey(key string, sep string) string {
	if sep == "" {
		sep = "/"
	}
	return fmt.Sprintf("%s%s%d", key, sep, time.Now().UnixNano())
}

//isValidUrl tests a string to determine if it is a well-structured url or not.
func isValidUrl(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}

	u, err := url.Parse(toTest)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

//GetEnvar get an environment variable and if it is black sets the given default.
//
// The value may be set as default string purposefully rather than omitted.
// In this case an empty string will be returned instead of the default string.
func GetEnv(variable, deflt string) string {
	out, set := os.LookupEnv(variable)
	if set {
		return out
	}
	return deflt
}

//NewHTTPClient is the default client to be used instead of default client.
//Ammended timeouts. See  //https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
func NewHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,

		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,

			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,

			ExpectContinueTimeout: 1 * time.Second,

			IdleConnTimeout:     30 * time.Second,
			MaxIdleConnsPerHost: 10,
			MaxIdleConns:        100,

			DisableKeepAlives: false,
			ForceAttemptHTTP2: true,
		},
	}
}

//returnErrorJSON takes the given error details and returns a JSON standard simple data
// struct to return to the client
func returnErrorJSON(errorTitle string, errorCode int, errorMsg string) (string, error) {
	nw := SimpleData{
		ResultCount: 0,
		Meta: Meta{
			Resource: ResourceServiceName,
		},
		Errors: map[string]string{"title": errorTitle, "code": strconv.Itoa(errorCode), "message": errorMsg},
	}

	binjson, err := json.MarshalIndent(nw, " ", " ")
	return string(binjson), fmt.Errorf("Error json marshalling Error message: %v", err)
}
