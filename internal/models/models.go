package models

import "database/sql"

type Models struct {
	Users              UserModel
	Pages              PageModel
	BellevueActivities BellevueActivityModel
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
		Pages:              PageModel{DB: db},
		Invoices:           InvoiceModel{DB: db},
		BellevueActivities: BellevueActivityModel{DB: db},
		Products:           ProductModel{DB: db},
		PriceCategories:    PriceCategoryModel{DB: db},
		Consumptions:       ConsumptionModel{DB: db},
		Comments:           CommentModel{DB: db},
		Activities:         ActivityModel{DB: db},
	}
}
