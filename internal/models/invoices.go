package models

import (
	"database/sql"
	"fmt"
	"time"
)

type BellevueModels struct {
	DB *sql.DB
}

type Invoice struct {
	ID            int
	UserID        int
	Period        time.Time
	MonthYear     string
	PeriodInt     string
	TotalPrice    int
	TotalEating   int
	TotalCoffees  int
	TotalLectures int
	TotalKiosk    int
	TotalSaunas   int
	State         string // open or paid
}

func (m *BellevueModels) ToggleState(ID int) {}

func (m *BellevueModels) GetAllInvoicesOfUser(userID int) ([]Invoice, error) {
	var user string
	err := m.DB.QueryRow("SELECT current_user").Scan(&user)
	if err != nil {}

	stmt := `
	SELECT
		id,
		period,
		period_yyyymm,
		total_price_rappen,
		total_eating,
		total_coffee,
		total_lecture,
		total_sauna,
		total_kiosk,
		state
	FROM bellevue.invoices
	WHERE user_id = $1
	ORDER BY period DESC
	`

	rows, err := m.DB.Query(stmt, userID)
	if err != nil {
		return nil, fmt.Errorf("DB.Query(stmt): %v", err)
	}

	defer rows.Close()

	var invoices []Invoice

	for rows.Next() {
		var invoice Invoice
		err = rows.Scan(
			&invoice.ID,
			&invoice.Period,
			&invoice.PeriodInt,
			&invoice.TotalPrice,
			&invoice.TotalEating,
			&invoice.TotalCoffees,
			&invoice.TotalLectures,
			&invoice.TotalSaunas,
			&invoice.TotalKiosk,
			&invoice.State,
		)
		if err != nil {
			return nil, fmt.Errorf("for rows.Next(): %v", err)
		}
		invoice.UserID = userID
		invoice.MonthYear = invoice.Period.Format("January 2006")
		invoices = append(invoices, invoice)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err(): %v", err)
	}

	return invoices, nil
}
