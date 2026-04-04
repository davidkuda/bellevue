package models

import (
	"database/sql"
	"fmt"
	"time"
)

type ActivityModel2 struct {
	DB *sql.DB
}

type Activity struct {
	ID        int
	UserID    int
	InvoiceID sql.NullInt32
	Date      time.Time
	Comment   sql.NullString
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m *ActivityModel2) InsertWithTransaction(activity *Activity, tx *sql.Tx) (int, error) {
	var err error

	stmt := `
	INSERT INTO activities (
		user_id, date, comment
	) VALUES (
		$1,      $2,   $3
	)
	RETURNING id;`
	row := tx.QueryRow(
		stmt,
		activity.UserID,
		activity.Date,
		activity.Comment,
	)
	var activityID int
	err = row.Scan(&activityID)
	if err != nil {
		return 0, fmt.Errorf("failed inserting activity: %v", err)
	}

	return activityID, nil
}

func (m *ActivityModel2) UpdateDateAndCommentTx(activity *Activity, tx *sql.Tx) error {
	var err error

	stmt := `
	UPDATE activities
	   SET date = $2,
	       comment = $3,
	       updated_at = NOW()
	 WHERE id = $1;`

	_, err = tx.Exec(
		stmt,
		activity.ID,
		activity.Date,
		activity.Comment,
	)
	if err != nil {
		return fmt.Errorf("failed inserting activity: %v", err)
	}

	return nil
}

func (m *ActivityModel2) Delete(activityID int, tx *sql.Tx) error {
	var err error

	stmt := `
	DELETE FROM activities
	WHERE id = $1;`

	_, err = tx.Exec(stmt, activityID)
	if err != nil {
		return fmt.Errorf("failed inserting activity: %v", err)
	}

	return nil
}

func (m *ActivityModel2) GetActivitiesOfInvoiceForUser(invoiceID int, userID int) ([]Activity, error) {
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
func (m *ActivityModel2) GetUninvoicedActivitiesForUser(userID int) (
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
