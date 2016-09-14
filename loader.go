package main

import (
	"io"
	"github.com/minio/minio-go"
	"regexp"
	"time"
	"github.com/pkg/errors"
	"strings"
)

type loader interface {
	LoadResource(name string) (io.ReadCloser, error)
}

type s3Loader struct {
	client minio.Client
	config s3Config
}

func news3Loader(c s3Config) (s3Loader, error) {
	s3Client, err := minio.New(c.domain, c.accKey, c.secretKey, true)
	if err != nil {
		return s3Loader{}, err
	}
	return s3Loader{*s3Client, c}, nil
}

func (s3Loader *s3Loader) LoadResource(path string) (io.ReadCloser, error) {
	s3Client := &s3Loader.client
	if s3Client == nil {
		return nil, errors.New("S3 bucket not initialised. Please call news3Loader(c s3Config) function first")
	}

	doneCh := make(chan struct{})
	defer close(doneCh)

	r := regexp.MustCompile("[0-9]{4}-(0[1-9]|1[0-2])-(0[1-9]|[1-2][0-9]|3[0-1])")
	var latestObj = &struct {
		object minio.ObjectInfo
		date   time.Time
	}{}

	for object := range s3Client.ListObjects(s3Loader.config.bucket, path, true, doneCh) {
		if object.Err != nil {
			return nil, object.Err
		}
		if (!strings.Contains(object.Key, path) || strings.Contains(object.Key, "md5")) {
			continue
		}

		dateFromName := r.FindStringSubmatch(object.Key)
		if len(dateFromName) == 0 {
			warnLogger.Printf("Ignoring file [%s]. Cannot parse name.", object.Key)
			continue
		}
		date, err := time.Parse("2006-01-02", dateFromName[0])
		if err != nil {
			errorLogger.Println(err)
			continue
		}

		if latestObj == nil || date.After(latestObj.date) {
			latestObj.object = object
			latestObj.date = date
		}
	}
	infoLogger.Printf("Reading from object: [%s] for resource [%s]", latestObj.object.Key, path)
	return s3Client.GetObject(s3Loader.config.bucket, latestObj.object.Key)
}
