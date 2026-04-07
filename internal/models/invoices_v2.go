package models

import (
	"database/sql"
	"time"
)

type InvoiceV2Model struct {
	DB *sql.DB
}

type InvoiceV2 struct {
	ID        int
	UserID    int
	Status    string // draft sent paid cancelled
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m *InvoiceV2Model) ToggleState(ID int) {}

func (m *InvoiceV2Model) NewInvoiceTx(userID int, tx *sql.Tx) (InvoiceV2, error) {
	stmt := `
	insert into invoices_v2 (
		user_id
	)
	values (
		$1
	)
	returning id, status;
	`

	row := tx.QueryRow(stmt, userID)

	newInvoice := InvoiceV2{
		UserID: userID,
	}

	err := row.Scan(&newInvoice.ID, &newInvoice.Status)
	if err != nil {
		return InvoiceV2{}, err
	}

	return newInvoice, nil
}

// NOTE: there is an idea there to assign activites of other users to an invoice, too
func (m *InvoiceV2Model) AssignOpenActivitiesByMonthToInvoiceForUserTx(
	month time.Time,
	userID int,
	invoceID int,
	tx *sql.Tx,
) (int, error) {
	var err error

	stmt := `
	update activities
	   set invoice_id = $1
	 where user_id = $2
	   AND date >= date_trunc('month', $3::date)
	   AND date <  date_trunc('month', $3::date) + interval '1 month'
	   and invoice_id is null;
	`

	result, err := tx.Exec(stmt, invoceID, userID, month)
	if err != nil {
		return 0, err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(n), nil
}

func (m *InvoiceV2Model) AssignOpenActivitiesByRangeToInviceForUserTx(
	start time.Time,
	end time.Time,
	userID int,
	invoceID int,
	tx *sql.Tx,
) (int, error) {
	var err error

	stmt := `
	update activities
	   set invoice_id = $1
	 where user_id = $2
	   AND date >= date_trunc('month', $3::date)
	   AND date <  date_trunc('month', $4::date)
	   and invoice_id is null;
	`

	result, err := tx.Exec(stmt, invoceID, userID, start, end)
	if err != nil {
		return 0, err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(n), nil
}

func (m *InvoiceV2Model) AssignAllOpenActivitiesToInvoiceTx(userID, invoceID int, tx *sql.Tx) (int, error) {
	var err error

	stmt := `
	update activities
	   set invoice_id = $1
	 where user_id = $2
	   and invoice_id is null;
	`

	result, err := tx.Exec(stmt, invoceID, userID)
	if err != nil {
		return 0, err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(n), nil
}

type InvoiceFinAccSum struct {
	FinancialAccountName string
	Price                int
}

// TODO: Add indexes to tables, e.g. consumptions.invoice_id
func (m *InvoiceV2Model) CalculatePriceCategoriesByInvoiceID(invoiceID int) ([]InvoiceFinAccSum, error) {
	data := make([]InvoiceFinAccSum, 0)
	var err error

	stmt := `
	select f.name, sum(total_price)
	  from consumptions c
	  join products p
	    on c.product_id = p.id
	  join financial_accounts f
	    on p.financial_account_id = f.id
	 where c.invoice_id = $1
	 group by f.name
	`

	rows, err := m.DB.Query(stmt, invoiceID)
	if err != nil {
		return []InvoiceFinAccSum{}, err
	}

	defer rows.Close()

	for rows.Next() {
		var p InvoiceFinAccSum
		err = rows.Scan(
			&p.FinancialAccountName,
			&p.Price,
		)
		data = append(data, p)
	}

	return data, nil
}
