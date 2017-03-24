package main

import (
	"archive/zip"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/minio/minio-go"
	"github.com/pkg/errors"
	"io"
	"path/filepath"
	"time"
)

const (
	weeklyObjectName = "/weekly.zip"
	weeklyDir        = "weekly"
	dateFormat       = "2006-01-02"
	fileExtension    = ".txt"
)

type resourceBundle interface {
	get(name string) (io.ReadCloser, error)
}

type rb struct {
	z *zip.Reader
}

func newResourceBundle(z *zip.Reader) resourceBundle {
	return &rb{z: z}
}

func (r *rb) get(name string) (io.ReadCloser, error) {
	name = filepath.Join(weeklyDir, name+fileExtension)
	log.Infof("Looking for file[%v]", name)
	for _, zf := range r.z.File {
		if zf.Name == name {
			return zf.Open()
		}
	}
	return nil, errors.New(fmt.Sprintf("Can't find file [%v]", name))
}

type loader interface {
	FindLatestResourcesFolder() (string, error)
	BucketExists() (bool, error)
	GetResourceBundle(pathPrefix string) (resourceBundle, error)
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

func (s3Loader *s3Loader) GetResourceBundle(pathPrefix string) (resourceBundle, error) {
	ob := pathPrefix + weeklyObjectName
	log.Infof("bucket=[%v],objectName=[%v]", s3Loader.config.bucket, ob)
	obj, err := s3Loader.client.GetObject(s3Loader.config.bucket, ob)
	if err != nil {
		log.Errorf("Error getting object[%v], %v", ob, err.Error())
		return nil, err
	}
	s, err := obj.Stat()
	if err != nil {
		log.Errorf("Error getting stat for object[%v], %v", ob, err.Error())
		defer obj.Close()
		return nil, err
	}
	z, err := zip.NewReader(obj, s.Size)
	if err != nil {
		defer obj.Close()
		log.Errorf("Error creating zip reader for object[%v], %v", ob, err.Error())
		return nil, err
	}
	return newResourceBundle(z), nil
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
			log.Infof("Ignoring file [%s]. Cannot parse name. Error was [%v]", object.Key, err)
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
	log.Infof("Found latest folder: [%s]", latestDateFormatted)
	return latestDateFormatted, nil
}

func (s3Loader *s3Loader) BucketExists() (bool, error) {
	return s3Loader.client.BucketExists(s3Loader.config.bucket)
}
