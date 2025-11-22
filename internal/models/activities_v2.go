package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

type ActivityModel struct {
	DB *sql.DB
}

// used to render in UI:
type ActivityMonth struct {
	Month      time.Time // e.g. 2025-11-01
	TotalPrice int
	Days       []ActivityDay
}

// used to render in UI:
type ActivityDay struct {
	Date       time.Time
	TotalPrice int // rappen
	Items      []LineItem
	Comment    string
}

// one product summary inside a day
type LineItem struct {
	ProductID int    `json:"product_id"`
	Name      string `json:"name"`
	UnitPrice int    `json:"unit_price"` // rappen
	Quantity  int    `json:"quantity"`
	// TODO: include price category name here and in template
}

func (m *ActivityModel) GetActivityMonths(userID int) ([]ActivityMonth, error) {
	days, err := m.GetActivityDays(userID)
	if err != nil {
		return nil, fmt.Errorf("could not get activity days: %e", err)
	}
	byMonth := map[string]*ActivityMonth{}

	for _, day := range days {
		key := day.Date.Format("2006-01") // month key
		monthStart, _ := time.Parse("2006-01-02", key+"-01")

		m, ok := byMonth[key]
		if !ok {
			m = &ActivityMonth{
				Month: monthStart,
			}
			byMonth[key] = m
		}

		m.Days = append(m.Days, day)
		m.TotalPrice += day.TotalPrice
	}

	// convert to slice & sort by month descending
	result := make([]ActivityMonth, 0, len(byMonth))
	for _, m := range byMonth {
		result = append(result, *m)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Month.After(result[j].Month)
	})

	return result, nil
}

func (m *ActivityModel) GetActivityDays(userID int) ([]ActivityDay, error) {
	const stmt = `
WITH per_day_product AS (
      SELECT c.user_id,
             c.date,
             p.id AS product_id,
             p.name AS product_name,
             sum(c.quantity) AS quantity,
             c.unit_price AS unit_price,
             sum(c.total_price) AS line_total
        FROM bellevue.consumptions c
        JOIN bellevue.products p ON p.id = c.product_id
       WHERE c.user_id = $1
    GROUP BY c.user_id, c.date, p.id, p.name, c.unit_price
)
  SELECT date,
         SUM(line_total) AS total_price,
         jsonb_agg(
             jsonb_build_object(
                 'product_id', product_id,
                 'name', product_name,
                 'unit_price', unit_price,
                 'quantity', quantity
             )
             ORDER BY product_name
         ) AS items
    FROM per_day_product
GROUP BY date
ORDER BY date DESC;
`

	rows, err := m.DB.Query(stmt, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var days []ActivityDay

	for rows.Next() {
		var (
			date       time.Time
			totalPrice int
			itemsJSON  []byte
		)

		if err := rows.Scan(&date, &totalPrice, &itemsJSON); err != nil {
			return nil, err
		}

		var items []LineItem
		if err := json.Unmarshal(itemsJSON, &items); err != nil {
			return nil, err
		}

		days = append(days, ActivityDay{
			Date:       date,
			TotalPrice: totalPrice,
			Items:      items,
		})
	}

	return days, rows.Err()
}
