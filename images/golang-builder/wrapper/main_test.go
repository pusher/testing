package main

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/bmizerany/assert"
)

func TestGetJUNitFileName(t *testing.T) {
	tmp, err := ioutil.TempDir("", "wrapper")
	assert.Equal(t, nil, err, "Expected to be able to create a temp directory")
	defer os.RemoveAll(tmp)

	// tmp should be empty, so first file should be indexed 0
	assert.Equal(t, path.Join(tmp, "junit_0.xml"), getJUnitFileName(tmp), "With an empty directory, expected filename is junit_0.xml")

	f, err := os.Create(path.Join(tmp, "junit_0.xml"))
	assert.Equal(t, nil, err, "Expected to be able to create a file in the temp directory")
	f.Close()

	// tmp should contain junit_0.xml, so next file should be indexed 1
	assert.Equal(t, path.Join(tmp, "junit_1.xml"), getJUnitFileName(tmp), "With one file in the directory, expected filename is junit_1.xml")
}
