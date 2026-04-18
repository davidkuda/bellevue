package viewmodels

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type ActivityViewModel struct {
	// DBModels models.Models
	DB *sql.DB
}

type Invoice struct {
	ID         int
	Sent       bool
	Status     string
	Date       time.Time
	MinDate    time.Time
	MaxDate    time.Time
	Activities []Activity
	TotalPrice int
	Categories []Category
}

type UninvoicedActivities struct {
	TotalPrice int
	Activities []Activity
}

type Activity struct {
	ID           int
	Date         time.Time
	Consumptions []Consumption
	TotalPrice   int
	Comment      string
}

type Consumption struct {
	ProductCode   string
	ProductName   string
	PriceCategory string
	Quantity      int
	UnitPrice     int
	TotalPrice    int
}

// intermediate representation of query results
type activityConsumptions []activityConsumption
type activityConsumption struct {
	activityID   int
	date         time.Time
	comment      string
	productCode  string
	productName  string
	pricecatName string
	quantity     int
	unit_price   int
	total_price  int
}

func (m *ActivityViewModel) GetUninvoicedActivitiesForUser(userID int) (*Invoice, error) {
	acs, err := m.getUninvoicedActivityConsumptionsForUser(userID)
	// TODO: should we return err on no rows found?
	if err != nil {
		return nil, fmt.Errorf("m.getUninvoicedActivityConsumptionsForUser(%d): %s", userID, err)
	}

	// TODO: what should we return if len(acs) == 0? should we return an error?
	// how to deal with this template wise / business wise?
	if len(acs) == 0 {
		return nil, nil
	}

	activities := acs.toViewModel()

	totalPrice := 0
	for i := range activities {
		totalPrice = totalPrice + activities[i].TotalPrice
	}

	uninvoicedActivities := Invoice{
		TotalPrice: totalPrice,
		Activities: activities,
	}
	uninvoicedActivities.MinDate, uninvoicedActivities.MaxDate = activityDateRange(activities)

	cats, err := m.GetUninvoicedCategoriesForUser(userID)
	if err != nil {
		return nil, fmt.Errorf("could not get uninvoiced categories for user: %v", err)
	}
	uninvoicedActivities.Categories = cats

	uninvoicedActivities.Sent = false

	return &uninvoicedActivities, nil
}

func (m *ActivityViewModel) GetAllInvoicesForUser(userID int) ([]*Invoice, error) {
	type inv struct {
		id     int
		status string
		date   time.Time
	}
	stmt := `
	select id, status, created_at
	from invoices_v2
	where user_id = $1
	order by created_at desc;`
	rows, err := m.DB.Query(stmt, userID)
	if err != nil {
		return nil, fmt.Errorf("could not get all invoice ids: %v", err)
	}
	defer rows.Close()

	var invs []inv

	for rows.Next() {
		var in inv
		err = rows.Scan(&in.id, &in.status, &in.date)
		if err != nil {
			return nil, fmt.Errorf("for rows.Next(): %v", err)
		}
		invs = append(invs, in)
	}

	var sentInvoices []*Invoice
	for _, in := range invs {
		invoice, err := m.GetInvoiceForUser(in.id, userID)
		if err != nil {
			return nil, fmt.Errorf("could not get sent invoice invoiceID=%d userID=%d: %v", userID, in.id, err)
		}
		invoice.Date = in.date
		invoice.Status = in.status
		sentInvoices = append(sentInvoices, invoice)
	}

	return sentInvoices, nil
}

func (m *ActivityViewModel) GetInvoiceForUser(invoiceID, userID int) (*Invoice, error) {
	acs, err := m.getActivityConsumptionsByInvoiceForUser(invoiceID, userID)
	if err != nil {
		return nil, fmt.Errorf("could not get activityConsumptions userID=%d: %s", userID, err)
	}

	if len(acs) == 0 {
		return nil, nil
	}

	activities := acs.toViewModel()

	totalPrice := 0
	for i := range activities {
		totalPrice = totalPrice + activities[i].TotalPrice
	}

	invoice := Invoice{
		TotalPrice: totalPrice,
		Activities: activities,
	}
	invoice.MinDate, invoice.MaxDate = activityDateRange(activities)

	cats, err := m.GetCategoriesByInvoiceIDForUser(invoiceID, userID)
	if err != nil {
		return nil, fmt.Errorf("could not get uninvoiced categories for user: %v", err)
	}
	invoice.Categories = cats

	invoice.ID = invoiceID
	invoice.Sent = true

	return &invoice, nil
}

func (m *ActivityViewModel) GetActivityByIDForUser(activityID, userID int) (*Activity, error) {
	acs, err := m.getActivityByIDForUser(activityID, userID)
	if err != nil {
		return nil, fmt.Errorf("m.getActivityByIDForUser(%d): %s", userID, err)
	}

	if len(acs) == 0 {
		return nil, errors.New("NoActivityFoundError")
	}

	// we expect exactly 1 activity in this slice:
	activities := acs.toViewModel()
	if len(activities) != 1 {
		return nil, errors.New("MoreThanOneActivityFoundError")
	}

	return &activities[0], nil
}

func activityDateRange(activities []Activity) (time.Time, time.Time) {
	if len(activities) == 0 {
		return time.Time{}, time.Time{}
	}

	minDate := activities[0].Date
	maxDate := activities[0].Date

	for i := 1; i < len(activities); i++ {
		if activities[i].Date.Before(minDate) {
			minDate = activities[i].Date
		}
		if activities[i].Date.After(maxDate) {
			maxDate = activities[i].Date
		}
	}

	return minDate, maxDate
}

func (m *ActivityViewModel) getUninvoicedActivityConsumptionsForUser(userID int) (activityConsumptions, error) {
	// NOTE: case when ... would be redundant if price_categories had a category "free_amount"
	// product.price_category_id can be null...
	stmt := `
	   SELECT a.id,
	          a.date,
	          coalesce(a.comment, ''),
	          p.code as product_code,
	          p.name as product_name,
	          case
	             when p.pricing_mode = 'custom' then 'free_amount'
	             else pc.name
	          end as pricecat_name,
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
	 ORDER BY a.date DESC, a.created_at DESC
	;
	`

	rows, err := m.DB.Query(stmt, userID)
	if err != nil {
		return nil, fmt.Errorf("DB.Query(stmt): %v", err)
	}

	defer rows.Close()

	var res []activityConsumption

	for rows.Next() {
		var r activityConsumption
		err = rows.Scan(
			&r.activityID,
			&r.date,
			&r.comment,
			&r.productCode,
			&r.productName,
			&r.pricecatName,
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

func (m *ActivityViewModel) getActivityConsumptionsByInvoiceForUser(invoiceID, userID int) (activityConsumptions, error) {
	// NOTE: case when ... would be redundant if price_categories had a category "free_amount"
	// product.price_category_id can be null...
	stmt := `
	   SELECT a.id,
	          a.date,
	          coalesce(a.comment, ''),
	          p.code as product_code,
	          p.name as product_name,
	          case
	             when p.pricing_mode = 'custom' then 'free_amount'
	             else pc.name
	          end as pricecat_name,
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
	    WHERE a.invoice_id = $1
	      AND user_id = $2
	 ORDER BY a.date DESC, a.created_at DESC
	;
	`

	rows, err := m.DB.Query(stmt, invoiceID, userID)
	if err != nil {
		return nil, fmt.Errorf("DB.Query(stmt): %v", err)
	}

	defer rows.Close()

	var res []activityConsumption

	for rows.Next() {
		var r activityConsumption
		err = rows.Scan(
			&r.activityID,
			&r.date,
			&r.comment,
			&r.productCode,
			&r.productName,
			&r.pricecatName,
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

func (m *ActivityViewModel) getActivityByIDForUser(activityID int, userID int) (activityConsumptions, error) {
	// TODO: the only difference in this fn to the previous is the WHERE statment (and the parameters / signature)
	// it feels kinda verbose to keep two such big functions...
	stmt := `
	   SELECT a.id,
	          a.date,
	          coalesce(a.comment, ''),
	          p.code as product_code,
	          p.name as product_name,
	          case
	             when p.pricing_mode = 'custom' then 'free_amount'
	             else pc.name
	          end as pricecat_name,
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
	    WHERE a.id = $1
	      AND user_id = $2
	 ORDER BY a.date DESC, a.id
	;
	`

	rows, err := m.DB.Query(stmt, activityID, userID)
	if err != nil {
		return nil, fmt.Errorf("DB.Query(stmt): %v", err)
	}

	defer rows.Close()

	var res []activityConsumption

	for rows.Next() {
		var r activityConsumption
		err = rows.Scan(
			&r.activityID,
			&r.date,
			&r.comment,
			&r.productCode,
			&r.productName,
			&r.pricecatName,
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

func (acs activityConsumptions) toViewModel() []Activity {
	activities := make([]Activity, 0)

	groupID := acs[0].activityID
	activity := Activity{
		ID:           acs[0].activityID,
		Date:         acs[0].date,
		Consumptions: make([]Consumption, 0),
		TotalPrice:   0,
		Comment:      acs[0].comment,
	}

	for _, ac := range acs {
		if ac.activityID != groupID {
			activities = append(activities, activity)

			groupID = ac.activityID
			activity = Activity{
				ID:           ac.activityID,
				Date:         ac.date,
				Consumptions: make([]Consumption, 0),
				TotalPrice:   0,
				Comment:      ac.comment,
			}
		}

		consumption := Consumption{
			ProductCode:   ac.productCode,
			ProductName:   ac.productName,
			PriceCategory: ac.pricecatName,
			Quantity:      ac.quantity,
			UnitPrice:     ac.unit_price,
			TotalPrice:    ac.total_price,
		}

		activity.TotalPrice += consumption.TotalPrice
		activity.Consumptions = append(activity.Consumptions, consumption)
	}

	activities = append(activities, activity)

	return activities
}
