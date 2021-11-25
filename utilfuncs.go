package sdsshared

import (
	"fmt"
	"net/url"
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
