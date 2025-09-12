package models

import "database/sql"

type Models struct {
	Users              UserModel
	Pages              PageModel
	Bellevue           BellevueModels
	BellevueActivities BellevueActivityModel
}

func New(db *sql.DB) Models {
	return Models{
		Users:              UserModel{DB: db},
		Pages:              PageModel{DB: db},
		Bellevue:           BellevueModels{DB: db},
		BellevueActivities: BellevueActivityModel{DB: db},
	}
}
