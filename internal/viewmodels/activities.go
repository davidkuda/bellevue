package viewmodels

import (
	"time"

	"github.com/davidkuda/bellevue/internal/models"
)

type ActivityViewModel struct {
	DBModels models.Models
}

type Invoice struct {
	ID         int
	Activities []Activity
}

type Activity struct {
	Date         time.Time
	Consumptions []Consumption
	TotalPrice   int
}

type Consumption struct {
	ProductName string
	Amount      int
	Count       int
	UnitPrice   int
	TotalPrice  int
}

func (m *ActivityViewModel) GetUninvoicedActivitiesForUser(userID int) (Invoice, error) {

}
