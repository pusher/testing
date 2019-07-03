package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/bmizerany/assert"
)

func TestGetJUNitFileName(t *testing.T) {
	tmp, err := ioutil.TempDir("", "wrapper")
	assert.Equal(t, nil, err, "Expected to be able to create a temp directory")
	defer os.RemoveAll(tmp)

	// tmp should be empty, so first file should definitely be uniquely named
	fileName := getJUnitFileName(tmp)
	_, err = os.Stat(fileName)
	assert.Equal(t, true, os.IsNotExist(err), fmt.Sprintf("With an empty directory, expected file (%s) to not exist", fileName))

	// Make sure the file starts with <tmp>/junit and ends with .xml
	assert.Equal(t, true, strings.HasPrefix(fileName, path.Join(tmp, "junit")), "Expected filename to start with 'junit'")
	assert.Equal(t, true, strings.HasSuffix(fileName, ".xml"), "Expected filename to end with '.xml'")

	f, err := os.Create(fileName)
	assert.Equal(t, nil, err, "Expected to be able to create a file in the temp directory")
	f.Close()

	// tmp should contain a file, so next file should have a different name
	fileName2 := getJUnitFileName(tmp)
	assert.NotEqual(t, fileName, fileName2, "Expected result of getJUnitFileName to be unique")
	_, err = os.Stat(fileName2)
	assert.Equal(t, true, os.IsNotExist(err), fmt.Sprintf("Expected file (%s) to not exist", fileName))
}
