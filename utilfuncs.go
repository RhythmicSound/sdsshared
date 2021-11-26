package sdsshared

import (
	"fmt"
	"net/url"
	"os"
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
	return fmt.Sprintf("%s%s%d", key, sep, time.Now().Unix())
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
