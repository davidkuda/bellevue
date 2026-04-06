package models

import (
	"database/sql"
	"fmt"
	"time"
)

type Consumptions []Consumption

type Consumption struct {
	ID         int
	ActivityID int
	ProductID  int
	UnitPrice  int
	Quantity   int
	TotalPrice int
	CreatedAt  time.Time
}

type ConsumptionModel struct {
	DB *sql.DB
}

func (m *ConsumptionModel) InsertManyWithTransaction(
	activityID int,
	consumptions []Consumption,
	tx *sql.Tx,
) error {
	var err error

	deleteQuery := `
	delete from consumptions
	 where activity_id = $1
	`
	if _, err = tx.Exec(deleteQuery, activityID); err != nil {
		return fmt.Errorf("failed deleting consumptions: %s", err)
	}

	// TODO: Deal with TaxPrice
	for _, c := range consumptions {
		ins := `
        insert into consumptions (
			activity_id,
			product_id,
			unit_price,
			quantity
		)
        values (
			$1,
			$2,
			$3,
			$4
		);
    `
		if _, err = tx.Exec(
			ins,
			c.ActivityID,
			c.ProductID,
			c.UnitPrice,
			c.Quantity,
		); err != nil {
			return fmt.Errorf("failed inserting consumptions: %s", err)
		}
	}

	return nil
}

// TODO: Maybe reuse this in the inserts instead of writing the statement.
func (m *ConsumptionModel) DeleteByActivityID(activityID int, tx *sql.Tx) error {
	var err error

	// We NEVER want to delete consumptions when they have an invoice ID.
	// therefore, check here first:
	stmt := `
	DELETE FROM consumptions
	WHERE activity_id = $1
	  AND activity_id IN (
		SELECT id
		  FROM activities
		 WHERE id = $1
		   AND invoice_id IS NULL
	  );
	`
	if _, err = tx.Exec(stmt, activityID); err != nil {
		return fmt.Errorf("failed deleting consumptions: %s", err)
	}

	return nil
}
