package models

import "database/sql"

type Models struct {
	Users              UserModel
	Invoices           InvoiceModel
	Products           ProductModel
	PriceCategories    PriceCategoryModel
	Consumptions       ConsumptionModel
	Comments           CommentModel
	Activities         ActivityModel
}

func New(db *sql.DB) Models {
	return Models{
		Users:              UserModel{DB: db},
		Invoices:           InvoiceModel{DB: db},
		Products:           ProductModel{DB: db},
		PriceCategories:    PriceCategoryModel{DB: db},
		Consumptions:       ConsumptionModel{DB: db},
		Comments:           CommentModel{DB: db},
		Activities:         ActivityModel{DB: db},
	}
}
