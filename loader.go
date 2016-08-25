package main

import (
	"errors"
	"github.com/rlmcpherson/s3gof3r"
	"io"
)

type loader interface {
	LoadResource(name string) (io.ReadCloser, error)
}

type s3Loader struct {
	bucket *s3gof3r.Bucket
}

func news3Loader(c s3Config) (s3Loader, error) {
	k, err := s3gof3r.EnvKeys()
	if err != nil {
		return s3Loader{}, err
	}
	s3 := s3gof3r.New(c.domain, k)
	loader := s3Loader{
		bucket: s3.Bucket(c.bucket),
	}
	return loader, nil
}

func (s3Loader *s3Loader) LoadResource(path string) (io.ReadCloser, error) {
	b := s3Loader.bucket
	if b == nil {
		return nil, errors.New("S3 bucket not initialised. Please call news3Loader(c s3Config) function first")
	}
	r, _, err := b.GetReader(path, nil)
	if err != nil {
		return nil, err
	}
	return r, nil
}
