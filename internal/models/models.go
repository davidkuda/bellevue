package models

import "database/sql"

type Models struct {
	Users              UserModel
	Pages              PageModel
	BellevueActivities BellevueActivityModel
}

func New(db *sql.DB) Models {
	return Models{
		Users:              UserModel{DB: db},
		Pages:              PageModel{DB: db},
		BellevueActivities: BellevueActivityModel{DB: db},
	}
}
