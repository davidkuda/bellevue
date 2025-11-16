package models

import "database/sql"

type Models struct {
	Users              UserModel
	Pages              PageModel
	BellevueActivities BellevueActivityModel
	Invoices           InvoiceModel
	Products           ProductModel
}

func New(db *sql.DB) Models {
	return Models{
		Users:              UserModel{DB: db},
		Pages:              PageModel{DB: db},
		Invoices:           InvoiceModel{DB: db},
		BellevueActivities: BellevueActivityModel{DB: db},
		Products:           ProductModel{DB: db},
	}
}
