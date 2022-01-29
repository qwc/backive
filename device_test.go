package backive

import (
	"os/exec"
	"testing"
)

func TestDevice(t *testing.T) {
	testDevice := new(Device)
	testDevice.Name = "Testdevice"
	testDevice.UUID = "123-456-789-abc-def"
	mockCmdRun = func(c *exec.Cmd) error {
		return nil
	}
	err := testDevice.Mount()
	if err != nil {
		t.Log("Should not fail, is mocked.")
		t.Fail()
	}
	if !testDevice.IsMounted() {
		t.Log("Should return true.")
		t.Fail()
	}
	err = testDevice.Unmount()
	if err != nil {
		t.Log("Should not fail, is mocked.")
		t.Fail()
	}
	if testDevice.IsMounted() {
		t.Log("Should return false.")
		t.Fail()
	}
}
