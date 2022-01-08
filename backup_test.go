package backive

import "testing"

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

func testPrepareRun() {
	/*
		Need to mock:
		- config.Settings.SystemMountPoint (to local test directory)
		- config.Settings.LogLocation (to local test directory)
		- exec.Command! (to NOT really execute something)

	*/
}
