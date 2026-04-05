package viewmodels

import (
	"fmt"
)

// e.g. food, lecture, etc.
type Category struct {
	Name       string
	TotalPrice int
}

// Current Invoice == invoice_id is null
func (m *ActivityViewModel) GetUninvoicedCategoriesForUser(userID int) ([]Category, error) {
	stmt := `
	  SELECT fa.view_name,
	         sum(total_price) AS total_price
	    FROM consumptions c
	    JOIN activities a
	      ON c.activity_id = a.id
	    JOIN products p
	      ON c.product_id = p.id
	    JOIN financial_accounts fa
	      ON p.financial_account_id = fa.id
	   WHERE a.invoice_id is null
	     AND a.user_id = $1
	GROUP BY fa.view_name
	ORDER BY total_price DESC
	;
	`

	rows, err := m.DB.Query(stmt, userID)
	if err != nil {
		return nil, fmt.Errorf("DB.Query(stmt): %v", err)
	}

	defer rows.Close()

	var res []Category

	for rows.Next() {
		var r Category
		err = rows.Scan(
			&r.Name,
			&r.TotalPrice,
		)

		if err != nil {
			return nil, fmt.Errorf("for rows.Next(): %v", err)
		}

		res = append(res, r)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err(): %v", err)
	}

	return res, nil
}

func (m *ActivityViewModel) GetCategoriesByInvoiceIDForUser(invoiceID, userID int) ([]Category, error) {
	stmt := `
	  SELECT fa.view_name,
	         sum(total_price) AS total_price
	    FROM consumptions c
	    JOIN activities a
	      ON c.activity_id = a.id
	    JOIN products p
	      ON c.product_id = p.id
	    JOIN financial_accounts fa
	      ON p.financial_account_id = fa.id
	   WHERE a.invoice_id = $1
	     AND a.user_id = $2
	GROUP BY fa.view_name
	ORDER BY total_price DESC
	;
	`

	rows, err := m.DB.Query(stmt, invoiceID, userID)
	if err != nil {
		return nil, fmt.Errorf("DB.Query(stmt): %v", err)
	}

	defer rows.Close()

	var res []Category

	for rows.Next() {
		var r Category
		err = rows.Scan(
			&r.Name,
			&r.TotalPrice,
		)

		if err != nil {
			return nil, fmt.Errorf("for rows.Next(): %v", err)
		}

		res = append(res, r)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err(): %v", err)
	}

	return res, nil
}
