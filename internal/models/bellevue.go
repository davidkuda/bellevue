package models

import (
	"database/sql"
	"time"
)

type SettingsModel struct {
	DB *sql.DB
}

// postgres: bellevue.price_category
type PriceCategory struct {
	ID          int
	Name        string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// postgres: bellevue.taxes
type Tax struct {
	ID          int
	MwstSatz    int16 // 8.1% => 810
	Code        *string
	Name        *string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// postgres: bellevue.financial_accounts
type FinancialAccount struct {
	ID          int
	Code        *int
	Name        string
	Description *string
	TaxID       int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// postgres: bellevue.products
type Product struct {
	ID                 int
	Name               string
	Description        *string
	FinancialAccountID int
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          *time.Time
}

// postgres: bellevue.prices
type Price struct {
	ID              int
	Price           int // in rappen
	ProductID       int
	PriceCategoryID int
	ValidFrom       time.Time
	ValidTo         *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

// postgres: bellevue.consumptions
type Consumption struct {
	ID        int64
	ProductID int
	Price     *int // snapshot price in rappen
	PriceID   int  // for analytics
	MwstID    int
	MwstPrice *int
	Date      *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
