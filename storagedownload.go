package sdsshared

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"cloud.google.com/go/storage"
)

//GCPDownload downloads assets from Google Cloud Storage with the given
//  Object name and within the given Bucket
func GCPDownload(bucket, object string) error {
	ctx := context.Background()
	//gcp client
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()
	//Local download target file
	if err := os.MkdirAll(LocalDownloadDir, 0755); err != nil {
		return err
	}
	file, err := os.Create(path.Join(LocalDownloadDir, "datasetupdate.zip"))
	defer file.Close()
	if err != nil {
		return err
	}
	//Cloud object target
	rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return fmt.Errorf("Object(%q).NewReader: %v", object, err)
	}
	defer rc.Close()
	//Download
	if _, err := io.Copy(file, rc); err != nil {
		return fmt.Errorf("Error downloading file from Cloud Storage: %v", err)
	}

	return nil
}
