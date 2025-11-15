package models

import (
	"database/sql"
	"fmt"
	"strings"
)

type Products []Product

type Product struct {
	ID            int
	Name          string
	Code          string
	PricingMode   string
	PriceCategory sql.NullString
	Price         sql.NullInt64
}

func (p Products) ToProductFormConfig() ProductFormConfig {
	for _, product := range p {
		fmt.Println(product)
	}
	return nil
}

type ProductFormConfig []ProductFormSpec

type ProductFormSpec struct {
	Label           string
	Code            string
	HasCategories   bool
	IsCustomAmount  bool
	PriceCategories []struct{
		Name       string
		PriceCents int
		Checked    bool
	}
}

type PriceKey struct {
	Code     string // e.g. "lunch"
	Category string // "regular", "" for custom
}

type ProductPrices map[PriceKey]int

type ProductModel struct {
	DB *sql.DB
}

func (m *ProductModel) GetProductFormConfig() (ProductFormConfig, error) {
	stmt := `
with product_form_specs as (
       select p.name,
              p.code,
              bool_or(p.price_category_id is not null) as has_categories,
              bool_or(p.pricing_mode = 'custom') as is_custom_amount,
              coalesce(
                array_to_string(
                  array_agg(distinct pc.name order by pc.name)
                    filter (where pc.name is not null),
                  ','
                ),
                ''
              ) as price_categories
         from bellevue.products p
    left join bellevue.price_categories pc
           on pc.id = p.price_category_id
        where p.deleted_at is null
        group by p.name, p.code
)
   select name,
          code,
          has_categories,
          is_custom_amount,
          price_categories
     from product_form_specs
left join product_form_order
    using (code)
 order by sort_order
;
	`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, fmt.Errorf("DB.Query(stmt): %v", err)
	}
	defer rows.Close()

	var pfts ProductFormConfig
	for rows.Next() {
		var pft ProductFormSpec
		var cats string
		err = rows.Scan(
			&pft.Label,
			&pft.Code,
			&pft.HasCategories,
			&pft.IsCustomAmount,
			&cats,
		)
		if err != nil {
			return nil, fmt.Errorf("for rows.Next(): %v", err)
		}

		if cats == "" {
			pft.PriceCategories = nil
		} else {
			pft.PriceCategories = strings.Split(cats, ",")
		}

		pfts = append(pfts, pft)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err(): %v", err)
	}

	return pfts, nil
}

func (m *ProductModel) GetAll() (Products, error) {
	var err error

	stmt := `
	   SELECT p.id, p.name, p.code, p.pricing_mode, cat.name, p.price
	     FROM products p
	LEFT JOIN price_categories cat
	       ON cat.id = p.price_category_id
	;
	`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, fmt.Errorf("DB.Query(stmt): %v", err)
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product
		err = rows.Scan(
			&product.ID,
			&product.Name,
			&product.Code,
			&product.PricingMode,
			&product.PriceCategory,
			&product.Price,
		)
		if err != nil {
			return nil, fmt.Errorf("for rows.Next(): %v", err)
		}
		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err(): %v", err)
	}

	return products, nil
}
