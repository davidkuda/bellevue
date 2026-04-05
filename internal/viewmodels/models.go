package viewmodels

import "database/sql"

type Models struct {
	Activities ActivityViewModel
}

func New(db *sql.DB) Models {
	return Models{
		Activities: ActivityViewModel{db},
	}
}
