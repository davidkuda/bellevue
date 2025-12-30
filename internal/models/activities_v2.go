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
	ProductID     int    `json:"product_id"`
	Name          string `json:"name"`
	Code          string `json:"code"`
	UnitPrice     int    `json:"unit_price"` // rappen
	Quantity      int    `json:"quantity"`
	PriceCategory string `json:"price_category"`
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
WITH
per_day_product AS (
   SELECT c.user_id,
          c.date,
          p.id AS product_id,
          p.name AS product_name,
          p.code AS product_code,
          sum(c.quantity) AS quantity,
          c.unit_price AS unit_price,
          pc.name AS pricecat,
          sum(c.total_price) AS line_total
     FROM consumptions c
     JOIN products p ON p.id = c.product_id
LEFT JOIN price_categories pc ON pc.id = c.pricecat_id
    WHERE c.user_id = $1
 GROUP BY c.user_id, c.date, p.id, p.name, c.unit_price, pricecat
),
jsonagg AS (
   SELECT date,
          SUM(line_total) AS total_price,
          jsonb_agg(
              jsonb_build_object(
                  'product_id', product_id,
                  'name', product_name,
                  'code', product_code,
                  'unit_price', unit_price,
                  'quantity', quantity,
                  'price_category', pricecat
              )
              ORDER BY pfo.sort_order, product_name
          ) AS items
     FROM per_day_product p
LEFT JOIN product_form_order pfo
       ON pfo.code = product_code
 GROUP BY p.date
)
   SELECT p.date,
          c.comment,
          total_price,
          items
     FROM jsonagg p
LEFT JOIN comments c
       ON c.date = p.date
      AND c.user_id = $1
 ORDER BY p.date DESC;
`

	rows, err := m.DB.Query(stmt, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var days []ActivityDay

	for rows.Next() {
		var itemsJSON []byte
		var comment sql.NullString
		day := ActivityDay{}

		if err := rows.Scan(
			&day.Date,
			&comment,
			&day.TotalPrice,
			&itemsJSON,
		); err != nil {
			return nil, err
		}

		if comment.Valid {
			day.Comment = comment.String
		}

		var items []LineItem
		if err := json.Unmarshal(itemsJSON, &items); err != nil {
			return nil, err
		}
		day.Items = items

		days = append(days, day)
	}

	return days, rows.Err()
}

func (m *ActivityModel) GetActivityDayForUser(t time.Time, userID int) (*ActivityDay, error) {
	var stmt string
	stmt = `
   select p.id,
          p.name,
          p.code,
          pc.name,
          c.unit_price,
          c.quantity
     from consumptions c
     join products p
       on p.id = c.product_id
left join price_categories pc
       on pc.id = c.pricecat_id
left join product_form_order pfo
       on pfo.code = p.code
    where date = $1
      and user_id = $2
	order by pfo.sort_order;
	`

	rows, err := m.DB.Query(stmt, t, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var day ActivityDay
	day.Date = t
	day.Items = make([]LineItem, 0)

	var priceCatName sql.NullString

	for rows.Next() {
		li := LineItem{}
		if err := rows.Scan(
			&li.ProductID,
			&li.Name,
			&li.Code,
			&priceCatName,
			&li.UnitPrice,
			&li.Quantity,
		); err != nil {
			return nil, err
		}

		if priceCatName.Valid {
			li.PriceCategory = priceCatName.String
		} else {
			li.PriceCategory = ""
		}

		day.Items = append(day.Items, li)
		day.TotalPrice += li.Quantity * li.UnitPrice
	}

	stmt = `
	select comment
	from comments
	where date = $1
	and user_id = $2
	`

	row := m.DB.QueryRow(stmt, t, userID)
	row.Scan(&day.Comment)

	return &day, nil
}
