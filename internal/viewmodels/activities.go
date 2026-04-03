package viewmodels

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/davidkuda/bellevue/internal/models"
)

type ActivityViewModel struct {
	DBModels models.Models
	DB       *sql.DB
}

type Invoice struct {
	ID         int
	Activities []Activity
}

type Activity struct {
	Date         time.Time
	Consumptions []Consumption
	TotalPrice   int
	Comment      string
}

type Consumption struct {
	ProductName  string
	PriceCatName string
	Amount       int
	Count        int
	UnitPrice    int
	TotalPrice   int
}

func (m *ActivityViewModel) GetUninvoicedActivitiesForUser(userID int) ([]Activity, error) {
	type queryResult struct {
		activityID  int
		date        time.Time
		name        string
		pricecat    string
		quantity    int
		unit_price  int
		total_price int
	}

	stmt := `
	   SELECT a.id,
	          a.date,
	          p.name,
	          pc.name,
	          c.quantity,
	          c.unit_price,
	          c.total_price
	     FROM consumptions c
	LEFT JOIN activities a
	       ON a.id = c.activity_id
	LEFT JOIN products p
	       ON p.id = c.product_id
	LEFT JOIN price_categories pc
	       ON pc.id = p.price_category_id
	    WHERE a.invoice_id is null
	      AND user_id = $1
	 ORDER BY a.date DESC
	;
	`

	// TODO: comments: maybe with map date=>comment

	rows, err := m.DB.Query(stmt, userID)
	if err != nil {
		return nil, fmt.Errorf("DB.Query(stmt): %v", err)
	}

	defer rows.Close()

	var res []queryResult

	for rows.Next() {
		var r queryResult
		err = rows.Scan(
			&r.activityID,
			&r.date,
			&r.name,
			&r.pricecat,
			&r.quantity,
			&r.unit_price,
			&r.total_price,
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
