package main

import (
	"archive/zip"
	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var files = []struct {
	Name, Body string
}{
	{"readme.txt", "This archive contains some text files."},
	{"gopher.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"},
	{"todo.txt", "Get animal handling licence.\nWrite more examples."},
}

func TestResourceBundle_getWorks(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "fis_test_zip")
	assert.NoError(t, err)
	log.Info(tmpfile.Name())
	defer func() {
		tmpfile.Close()
		err = os.Remove(tmpfile.Name())
		assert.NoError(t, err)
	}()
	w := zip.NewWriter(tmpfile)
	for _, file := range files {
		f, err := w.Create(filepath.Join("weekly", file.Name))
		assert.NoError(t, err)
		_, err = f.Write([]byte(file.Body))
		assert.NoError(t, err)
	}

	assert.NoError(t, w.Close())
	stat, err := tmpfile.Stat()
	assert.NoError(t, err)
	z, err := zip.NewReader(tmpfile, stat.Size())
	assert.NoError(t, err)
	bundle := newResourceBundle(z)

	t.Run("Should read zip file", func(t *testing.T) {
		g, err := bundle.get("readme")
		assert.NoError(t, err)
		defer g.Close()
		bs, err := ioutil.ReadAll(g)
		assert.NoError(t, err)
		assert.Equal(t, "This archive contains some text files.", string(bs[:]))
	})

	t.Run("Should error if not found", func(t *testing.T) {
		_, err := bundle.get("file_that_is_not_there")
		assert.Error(t, err)
	})
}
