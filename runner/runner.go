package runner

import (
	"fmt"
	"os"

	"github.com/qwc/backive/backup"
	"github.com/qwc/backive/config"
)

// Run runs the backup script with appropriate rights.
func Run(b backup.Backup) error {
	cfg := config.Get()
	if cfg.Devices[b.Name].IsMounted() {
		checkExistence := func(path string, name string) error {
			if _, err := os.Stat(path); err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("%s does not exist", name)
				} else {
					return fmt.Errorf("Error when checking %s: %w", name, err)
				}
			}
			return nil
		}
		// Check for existence of target dir
		if err := checkExistence(b.TargetDir, "target directory"); err != nil {
			return err
		}
		// check for existence of source dir
		if err := checkExistence(b.SourceDir, "source directory"); err != nil {
			return err
		}
		// check for existence of script path
		if err := checkExistence(b.ScriptPath, "script path"); err != nil {
			return err
		}
		// setup script environment including user to use
		// run script
	}
	// quit with error that the device is not available.
	return fmt.Errorf("The device is not mounted")
}
