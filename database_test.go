package backive

import (
	"io/fs"
	"testing"
)

func TestDatabase(t *testing.T) {
	mockOsStat = func(p string) (fs.FileInfo, error) {
		return nil, nil
	}
	db := new(Database)
	mockOsReadFile = func(p string) ([]byte, error) {
		return []byte("{}"), nil
	}
	db.Load()

	mockOsWriteFile = func(p string, data []byte, rights fs.FileMode) error {
		return nil
	}

	db.Save()
}
