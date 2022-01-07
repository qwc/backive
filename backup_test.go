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
