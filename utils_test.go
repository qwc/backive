package backive

import (
	"fmt"
	"io/fs"
	"os"
	"testing"
)

var creatingDir = false

func TestCreateDirectoryIfNotExists(t *testing.T) {
	mockOsMkdirAll = func(dir string, mode os.FileMode) error {
		t.Log("Creating directories")
		creatingDir = true
		return nil
	}
	mockOsStat = os.Stat
	CreateDirectoryIfNotExists("/somewhere/which.does/not/exist")
	if !creatingDir {
		t.Log("Should have called MkdirAll")
		t.Fail()
	}
	mockOsStat = func(dir string) (fs.FileInfo, error) {
		return nil, fmt.Errorf("Just some error for testing")
	}
	err := CreateDirectoryIfNotExists("asdfasdfasdf")
	if err == nil {
		t.Log("Should have an error here")
		t.Fail()
	}
}
