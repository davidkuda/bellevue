package models

import (
	"database/sql"
	"fmt"
	"time"
)

type ActivityModel struct {
	DB *sql.DB
}

type Activity struct {
	ID        int
	UserID    int
	Date      time.Time
	Comment   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m *ActivityModel) GetActivitiesOfInvoiceForUser(invoiceID int, userID int) ([]Activity, error) {
	stmt := `
	SELECT id,
	       user_id,
	       date,
	       comment,
	       created_at,
	       updated_at
	  FROM activities
	 WHERE user_id = $1
	   AND invoice_id = $2
	`

	rows, err := m.DB.Query(stmt, invoiceID, userID)
	if err != nil {
		return nil, fmt.Errorf("DB.Query(stmt): %v", err)
	}

	defer rows.Close()

	var activities []Activity

	for rows.Next() {
		var activity Activity
		err = rows.Scan(
			&activity.ID,
			&activity.UserID,
			&activity.Date,
			&activity.Comment,
			&activity.CreatedAt,
			&activity.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("for rows.Next(): %v", err)
		}

		activities = append(activities, activity)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err(): %v", err)
	}

	return activities, nil
}

// activities can have a null on invoice_id, which means that the activity was
// not invoiced yet or that we haven't created yet an invoice for the user.
func (m *ActivityModel) GetUninvoicedActivitiesForUser(userID int) (
	[]Activity,
	error,
) {
	stmt := `
	SELECT id,
	       user_id,
	       date,
	       comment,
	       created_at,
	       updated_at
	  FROM activities
	 WHERE user_id = $1
	   AND invoice_id is null
	`

	rows, err := m.DB.Query(stmt, userID)
	if err != nil {
		return nil, fmt.Errorf("DB.Query(stmt): %v", err)
	}

	defer rows.Close()

	var activities []Activity

	for rows.Next() {
		var activity Activity
		err = rows.Scan(
			&activity.ID,
			&activity.UserID,
			&activity.Date,
			&activity.Comment,
			&activity.CreatedAt,
			&activity.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("for rows.Next(): %v", err)
		}

		activities = append(activities, activity)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err(): %v", err)
	}

	return activities, nil
}
