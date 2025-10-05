package models

import "database/sql"

type Models struct {
	Users              UserModel
	Pages              PageModel
	BellevueActivities BellevueActivityModel
	Invoices           InvoiceModel
}

func New(db *sql.DB) Models {
	return Models{
		Users:              UserModel{DB: db},
		Pages:              PageModel{DB: db},
		Invoices:           InvoiceModel{DB: db},
		BellevueActivities: BellevueActivityModel{DB: db},
	}
}
