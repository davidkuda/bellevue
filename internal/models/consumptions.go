package models

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Consumptions []Consumption

type Consumption struct {
	ID         int
	UserID     int
	ProductID  int
	TaxID      int
	PriceCatID sql.NullInt64
	Date       time.Time
	UnitPrice  int
	Quantity   int
	TotalPrice int
	CreatedAt  time.Time
}

type ConsumptionModel struct {
	DB *sql.DB
}

func (m *ConsumptionModel) InsertMany(
	userID int,
	date time.Time,
	consumptions Consumptions,
) error {
	var err error

	ctx := context.TODO()
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed starting transaction: %e", err)
	}
	defer tx.Rollback()

	deleteQuery := `
	delete from consumptions
	where user_id = $1
	  and date = $2
	`
	if _, err := tx.ExecContext(ctx, deleteQuery, userID, date); err != nil {
		return fmt.Errorf("failed deleting consumptions: %e", err)
	}

	// TODO: Deal with TaxPrice
	for _, c := range consumptions {
		ins := `
        insert into consumptions (
			user_id,
			product_id,
			tax_id,
			pricecat_id,
			date,
			unit_price,
			quantity
		)
        values (
			$1,
			$2,
			(select tax_id from products where id = $2),
			$3,
			$4,
			$5,
			$6
		);
    `
		if _, err := tx.ExecContext(
			ctx,
			ins,
			c.UserID,
			c.ProductID,
			c.PriceCatID,
			c.Date,
			c.UnitPrice,
			c.Quantity,
		); err != nil {
			return fmt.Errorf("failed inserting consumptions: %e", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed committing transaction: %e", err)
	}

	return nil
}
