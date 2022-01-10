package backive

import (
	"fmt"
	"os/exec"
	"path"
	"runtime"
	"testing"
)

func getCurrentFilePath() string {
	pc, file, line, ok := runtime.Caller(1)
	fmt.Printf("pc: %d, file: %s, line: %d, ok: %v\n", pc, file, line, ok)
	return file
}

type MockCmd struct{}

func (c *MockCmd) Run() error {
	return nil
}

func TestFindBackupsForDevice(t *testing.T) {
	var testBackups = Backups{}

	testBackups["backup1"] = &Backup{
		Name:         "backup1",
		TargetDevice: "dev1",
	}
	testBackups["backup2"] = &Backup{
		Name:         "backup2",
		TargetDevice: "dev1",
	}
	testBackups["backup3"] = &Backup{
		Name:         "backup3",
		TargetDevice: "dev2",
	}

	var testDevice = Device{
		Name: "dev1",
	}
	bkps, found := testBackups.FindBackupsForDevice(testDevice)

	if !found {
		t.Log("found has to be true")
		t.Fail()
	}
	if len(bkps) != 2 {
		t.Log("Length of the returned backup slice has to be 2")
		t.Fail()
	}
	for _, b := range bkps {
		if b.TargetDevice != testDevice.Name {
			t.Log("All resulting elements of the returned slice have to have the questioned device as TargetDevice!")
			t.Fail()
		}
	}
}

func TestCanRun(t *testing.T) {
	var bkpTargetPathMissing = Backup{
		Name:       "targetPathMissing",
		ScriptPath: "Somethingsomething",
	}
	err := bkpTargetPathMissing.CanRun()
	if err == nil {
		t.Log("Missing targetPath has to fail function 'CanRun()'")
		t.Fail()
	}

	var bkpScriptPathMissing = Backup{
		Name:       "scriptPathMissing",
		TargetPath: "somethingsomething",
	}
	err = bkpScriptPathMissing.CanRun()
	if err == nil {
		t.Log("Missing scriptPath has to fail function 'CanRun()'")
		t.Fail()
	}

	var bkpFrequencyZero = Backup{
		Name:       "testFrequencyZero",
		TargetPath: "somewhere",
		ScriptPath: "somehwere_else",
		Frequency:  0,
	}
	var bkpFrequencySeven = Backup{
		Name:       "testFrequencySeven",
		TargetPath: "somewhere",
		ScriptPath: "somewhere_else",
		Frequency:  7,
	}
	database.Load()
	runs.Load(database)
	runs.RegisterRun(&bkpFrequencyZero)
	err = bkpFrequencyZero.CanRun()
	if err != nil {
		t.Log("Frequency zero can be executed anytime.")
		t.Fail()
	}

	runs.RegisterRun(&bkpFrequencySeven)
	err = bkpFrequencySeven.CanRun()
	if err == nil {
		t.Log("Frequency 7 must give an error about not having reached the interval")
		t.Fail()
	}
}

func setupNewTestEnv(subdir string) {
	file := getCurrentFilePath()
	dir, _ := path.Split(file)
	dir = path.Join(dir, "test", "_workarea", subdir)
	config.Settings.SystemMountPoint = path.Join(dir, "mnt")
	config.Settings.LogLocation = path.Join(dir, "log")
	fmt.Printf("SystemMountPoint: %s, LogLocation: %s\n", config.Settings.SystemMountPoint, config.Settings.LogLocation)
}

func TestPrepareRun(t *testing.T) {
	setupNewTestEnv("preparerun")

	mock_cmd_Run = func(c *exec.Cmd) error {
		return nil
	}
	var testBkp = Backup{
		Name:         "testbkp",
		TargetDevice: "mytarget",
		TargetPath:   "mypath",
	}
	err := testBkp.PrepareRun()
	if err != nil {
		t.Log("When this fails, something's fishy...")
		t.Fail()
	}
}

func TestRun(t *testing.T) {
	setupNewTestEnv("run")
	config.Devices = map[string]*Device{
		"mytarget": new(Device),
	}
	config.Devices["mytarget"].Name = "mytarget"
	config.Devices["mytarget"].UUID = "123-456-789-abc-def"

	mock_cmd_Run = func(c *exec.Cmd) error {
		return nil
	}
	var testBkp = Backup{
		Name:         "testbkp",
		TargetDevice: "mytargets",
		TargetPath:   "mypath",
		ScriptPath:   "/dev/null",
		SourcePath:   "/dev/random",
	}
	err := testBkp.Run()
	if err == nil || err.Error() != "device mytargets not found" {
		if err != nil {
			t.Logf("The error should be 'device mytargets not found', but is '%s'", err.Error())
			t.Fail()
		}
	}
	testBkp.TargetDevice = "mytarget"
	config.Devices["mytarget"].Mount()
	err = testBkp.PrepareRun()
	err = testBkp.Run()
	if err != nil {
		t.Logf("Error which should not occur: %s", err)
		t.Fail()
	}
	mock_cmd_Run = func(c *exec.Cmd) error {
		return nil
	}
}
