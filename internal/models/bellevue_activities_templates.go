package models

import (
	"fmt"
	"time"
)


// used to render tables with activities
type BellevueActivityOverviews []BellevueActivityOverview

func (m *BellevueActivityModel) NewBellevueActivityOverviews(userID int) (BellevueActivityOverviews, error) {

	BAs, err := m.GetAllByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed GetAllByUser(%d): %v", userID, err)
	}

	if len(BAs) == 0 {
		return BellevueActivityOverviews{}, nil
	}

	BAOs := make(BellevueActivityOverviews, 0)
	var BAO BellevueActivityOverview // buffer
	var monthYear, trackMonthYear string
	var BA BellevueActivity

	// commit first
	BA = BAs[0]
	monthYear = BA.Date.Format("January 2006")
	trackMonthYear = monthYear
	BAO.BellevueActivities = append(BAO.BellevueActivities, BA)

	for i := 1; i < len(BAs); i++ {
		BA = BAs[i]

		monthYear = BA.Date.Format("January 2006")

		// same month: add BA to buffer:
		if monthYear == trackMonthYear {
			BAO.BellevueActivities = append(BAO.BellevueActivities, BA)
		}

		// new month: commit buffer, reset buffer, add BA to buffer:
		if monthYear != trackMonthYear {
			// commit
			BAO.CalculateTotalPrice()
			BAO.MonthYear = trackMonthYear
			BAOs = append(BAOs, BAO)

			// reset
			BAO = BellevueActivityOverview{}
			trackMonthYear = monthYear

			// add
			BAO.BellevueActivities = append(BAO.BellevueActivities, BA)
		}
	}
	// commit last:
	BAO.CalculateTotalPrice()
	BAO.MonthYear = trackMonthYear
	BAOs = append(BAOs, BAO)

	return BAOs, nil
}

type BellevueActivityOverview struct {
	BellevueActivities []BellevueActivity
	TotalPrice         int
	MonthYear          string
}

func (b *BellevueActivityOverview) CalculateTotalPrice() {
	var sum int
	for _, a := range b.BellevueActivities {
		sum += a.TotalPrice
	}
	b.TotalPrice = sum
}

type Item struct {
	Activity string
	Count    int
}

func NewItem(s string, n int) Item {
	return Item{s, n}
}

type BellevueOfferings []Offer

// TODO: I don't like these structs ... needs refactoring
// TODO: REFACTOR: Maybe combine with Item? Will see when implementing edit and open the form.
type Offer struct {
	Label string
	Price int
	ID    string
	Count int
}

func (b *BellevueActivity) NewBellevueOfferings() BellevueOfferings {
	return BellevueOfferings{
		Offer{
			Label: "Breakfast (8.00 CHF):",
			Price: 800,
			ID:    "breakfasts",
			Count: b.Breakfasts,
		},
		Offer{
			Label: "Lunch (11.00 CHF):",
			Price: 1100,
			ID:    "lunches",
			Count: b.Lunches,
		},
		Offer{
			Label: "Dinner (11.00 CHF):",
			Price: 1100,
			ID:    "dinners",
			Count: b.Dinners,
		},
		Offer{
			Label: "Coffee (1.00 CHF):",
			Price: 100,
			ID:    "coffees",
			Count: b.Coffees,
		},
		Offer{
			Label: "Sauna (7.50 CHF):",
			Price: 750,
			ID:    "saunas",
			Count: b.Saunas,
		},
		Offer{
			Label: "Lectures (12.00 CHF):",
			Price: 1200,
			ID:    "lectures",
			Count: b.Lectures,
		},
	}
}

func NewBellevueActivity() *BellevueActivity {
	return &BellevueActivity{
		Date: time.Now(),
	}
}

func (b *BellevueActivity) PopulateItems() {
	b.Items = make([]Item, 0)
	b.addItem(b.Breakfasts, "Breakfast", "Breakfasts", "8.00 ")
	b.addItem(b.Lunches, "Lunch", "Lunches", "11.00")
	b.addItem(b.Dinners, "Dinner", "Dinners", "11.00")
	b.addItem(b.Coffees, "Coffee", "Coffees", "1.00")
	b.addItem(b.Saunas, "Sauna", "Saunas", "7.50")
	b.addItem(b.Lectures, "Lecture", "Lectures", "12.00")
}

func (b *BellevueActivity) addItem(count int, singular, plural, price string) {
	if count <= 0 {
		return
	} else if count == 1 {
		b.Items = append(b.Items, Item{fmt.Sprintf("%s à %s CHF", singular, price), count})
	} else {
		b.Items = append(b.Items, Item{fmt.Sprintf("%s à %s CHF", plural, price), count})
	}
}

func (a *BellevueActivity) CalculatePrice() {
	// TODO: Define prices elsewhere, WebForm->DB? YAML?
	//       But keep it simple for now x)
	prices := map[string]int{
		"breakfast": 800,
		"lunch":     1100,
		"dinner":    1100,
		"coffee":    100,
		"sauna":     750,
		"lecture":   1200,
	}

	a.TotalPrice = (a.Breakfasts*prices["breakfast"] +
		a.Lunches*prices["lunch"] +
		a.Dinners*prices["dinner"] +
		a.Coffees*prices["coffee"] +
		a.Saunas*prices["sauna"] +
		a.Lectures*prices["lecture"])

	a.TotalPrice += a.SnacksCHF
}
