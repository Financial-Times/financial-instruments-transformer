package main

import (
	"github.com/minio/minio-go"
	"github.com/pkg/errors"
	"io"
	"strings"
	"time"
)

type loader interface {
	FindLatestResourcesFolder() (string, error)
	LoadResource(pathPrefix string, name string) (io.ReadCloser, error)
	BucketExists() (bool, error)
}

type s3Loader struct {
	client minio.Client
	config s3Config
}

const dateFormat = "2006-01-02"

func news3Loader(c s3Config) (s3Loader, error) {
	s3Client, err := minio.New(c.domain, c.accKey, c.secretKey, true)
	if err != nil {
		return s3Loader{}, err
	}
	return s3Loader{*s3Client, c}, nil
}

func (s3Loader *s3Loader) FindLatestResourcesFolder() (string, error) {
	s3Client := &s3Loader.client
	if s3Client == nil {
		return "", errors.New("S3 bucket not initialised. Please call news3Loader(c s3Config) function first")
	}

	doneCh := make(chan struct{})
	defer close(doneCh)

	var latestDate time.Time

	for object := range s3Client.ListObjects(s3Loader.config.bucket, "", true, doneCh) {

		if object.Err != nil {
			return "", object.Err
		}

		date, err := time.Parse(dateFormat, object.Key[:len(dateFormat)])
		if err != nil {
			infoLogger.Printf("Ignoring file [%s]. Cannot parse name. Error was [%v]", object.Key, err)
			continue
		}

		if latestDate.IsZero() || date.After(latestDate) {
			latestDate = date
		}
	}

	if latestDate.IsZero() {
		return "", errors.New("Could not find any directory that has a date as its name.")
	}

	latestDateFormatted := latestDate.Format(dateFormat)
	infoLogger.Printf("Found latest folder: [%s]", latestDateFormatted)
	return latestDateFormatted, nil
}

func (s3Loader *s3Loader) LoadResource(pathPrefix string, resourceName string) (io.ReadCloser, error) {
	s3Client := &s3Loader.client
	if s3Client == nil {
		return nil, errors.New("S3 bucket not initialised. Please call news3Loader(c s3Config) function first")
	}

	doneCh := make(chan struct{})
	defer close(doneCh)

	for object := range s3Client.ListObjects(s3Loader.config.bucket, pathPrefix, true, doneCh) {
		if object.Err != nil {
			return nil, object.Err
		}

		if strings.Contains(object.Key, resourceName) {
			infoLogger.Printf("Loading object: [%s] for resource [%s]", object.Key, resourceName)
			return s3Client.GetObject(s3Loader.config.bucket, object.Key)
		}
	}

	return nil, errors.New("Cannot find resource with name [" + resourceName + "] and path prefix [" + pathPrefix + "].")
}

func (s3Loader *s3Loader) BucketExists() (bool, error) {
	return s3Loader.client.BucketExists(s3Loader.config.bucket)
}
