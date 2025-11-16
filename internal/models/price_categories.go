package models

import (
	"database/sql"
	"fmt"
)

type PriceCategoryMap map[string]bool

type PriceCategories []PriceCategory

type PriceCategory struct {
	ID   int
	Name string
}

type PriceCategoryModel struct {
	DB *sql.DB
}

func (m *PriceCategoryModel) GetPriceCatMap() (PriceCategoryMap, error) {
	pcm := PriceCategoryMap{}
	pricecats, err := m.GetAll()
	if err != nil {
		return nil, fmt.Errorf("PriceCategoryModel.GetAll: %v", err)
	}
	for _, pricecat := range pricecats {
		pcm[pricecat.Name] = true
	}

	return pcm, nil
}

func (m *PriceCategoryModel) GetAll() (PriceCategories, error) {
	var err error

	stmt := `
	SELECT id, name
	  FROM price_categories
	`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, fmt.Errorf("DB.Query(stmt): %v", err)
	}
	defer rows.Close()

	var pricecats PriceCategories
	for rows.Next() {
		var pricecat PriceCategory
		err = rows.Scan(
			&pricecat.ID,
			&pricecat.Name,
		)
		if err != nil {
			return nil, fmt.Errorf("for rows.Next(): %v", err)
		}
		pricecats = append(pricecats, pricecat)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err(): %v", err)
	}

	return pricecats, nil
}
