package calendar

import (
	"testing"
	"time"

	"teslamate-calendar/internal/model"
)

func TestDriveUIDStable(t *testing.T) {
	now := time.Now().UTC()
	drives := []model.Drive{{ID: 1, StartDate: &now, EndDate: &now}}
	e1 := DriveEvents("1", "Model 3", drives, "normal", true, "", "")
	e2 := DriveEvents("1", "Model 3", drives, "normal", true, "", "")
	if len(e1) != 1 || len(e2) != 1 || e1[0].UID != e2[0].UID {
		t.Fatal("drive uid is not stable")
	}
}

func TestChargeUIDStable(t *testing.T) {
	now := time.Now().UTC()
	charges := []model.Charge{{ID: 9, StartDate: &now, EndDate: &now}}
	e1 := ChargeEvents("1", "Model 3", charges, "normal", true, "", "")
	e2 := ChargeEvents("1", "Model 3", charges, "normal", true, "", "")
	if len(e1) != 1 || len(e2) != 1 || e1[0].UID != e2[0].UID {
		t.Fatal("charge uid is not stable")
	}
}
