package models

import (
	"database/sql"
	"fmt"
	"time"
)

type InvoiceModel struct {
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

func (m *InvoiceModel) ToggleState(ID int) {}

func (m *InvoiceModel) GetAllInvoicesOfUser(userID int) ([]Invoice, error) {
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

func (m *InvoiceModel) GetInvoiceOfLastMonth(user User) (Invoice, error) {
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
	WHERE
		user_id = $1
		AND period = date_trunc(
			'month',
			current_date - interval '1 month'
		)::date;
	`

	row := m.DB.QueryRow(stmt, user.ID)

	var invoice Invoice
	err := row.Scan(
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
		return Invoice{}, fmt.Errorf("failed fetching most recent invoice for user with ID %d: %e", user.ID, err)
	}

	return invoice, nil
}
