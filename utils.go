package backive

import (
	"os"
)

var mockOsStat = os.Stat
var mockOsMkdirAll = os.MkdirAll
var mockOsIsNotExist = os.IsNotExist

// CreateDirectoryIfNotExists Checks for a directory string and creates the directory if it does not exist, must be a absolute path.
func CreateDirectoryIfNotExists(dir string) error {
	if _, err := mockOsStat(dir); err == nil {
		//ignore
	} else if mockOsIsNotExist(err) {
		return mockOsMkdirAll(dir, 0755)
	} else {
		return err
	}
	return nil
}
