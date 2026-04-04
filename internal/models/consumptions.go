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

func (m *ConsumptionModel) CountOpenConsumptionsForUser(userID int) (int, error) {
	var count int
	var err error

	stmt := `
	select count(*)
	from consumptions
	where user_id = $1
	and invoice_id is null;
	`

	row := m.DB.QueryRow(stmt, userID)

	err = row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
