package models

import "database/sql"

type Models struct {
	Users           UserModel
	Invoices        InvoiceModel
	InvoicesV2      InvoiceV2Model
	Products        ProductModel
	PriceCategories PriceCategoryModel
	Consumptions    ConsumptionModel
	Comments        CommentModel
	Activities      ActivityModel
}

func New(db *sql.DB) Models {
	return Models{
		Users:           UserModel{DB: db},
		Invoices:        InvoiceModel{DB: db},
		InvoicesV2:      InvoiceV2Model{DB: db},
		Products:        ProductModel{DB: db},
		PriceCategories: PriceCategoryModel{DB: db},
		Consumptions:    ConsumptionModel{DB: db},
		Comments:        CommentModel{DB: db},
		Activities:      ActivityModel{DB: db},
	}
}
