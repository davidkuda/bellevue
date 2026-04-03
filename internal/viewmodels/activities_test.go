package viewmodels

import (
	"testing"

	"github.com/davidkuda/bellevue/internal/envcfg"
)


// this test needs a database connection and at least one uninvoiced activity
func TestGetUninvoicedActivitiesForUser(t *testing.T) {
	db, err := envcfg.DB()
	if err != nil {
		t.Fatalf("could not open DB: %v\n", err)
	}
	defer db.Close()

	model := ActivityViewModel{db}

	userID := 1

	acs, err := model.getUninvoicedActivityConsumptionsForUser(userID)
	if err != nil {
		t.Fatalf("m.getUninvoicedActivityConsumptionsForUser(%d): %s", userID, err)
	}

	if len(acs) <= 0 {
		t.Fatal("something wrong here")
	}

	t.Log(len(acs))
}
