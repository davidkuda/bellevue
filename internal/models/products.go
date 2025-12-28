package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"maps"
)

type ProductIDMap map[string]int

type Products []Product

type Product struct {
	ID            int
	Name          string
	Code          string
	PricingMode   string
	PriceCategory sql.NullString
	Price         sql.NullInt64
}

type ProductFormConfig struct {
	Prices map[string]int
	Specs  []ProductFormSpec
}

// clones struct (nested map and slices needs recreation, otherwise, updates app.productFormConfig)
// and updates values, e.g. instead of 0 lunch, use 2 lunches, regular.
// intended to be called when editing an ActivityDay.
func (c ProductFormConfig) WithValues(ad *ActivityDay) ProductFormConfig {

	clone := ProductFormConfig{
		Prices: make(map[string]int, len(c.Prices)),
		Specs:  make([]ProductFormSpec, len(c.Specs)),
	}

	// copy map, this is maybe not necessary (yet?)
	maps.Copy(clone.Prices, c.Prices)

	// copy specs + nested slices
	for i, spec := range c.Specs {
		clone.Specs[i] = spec

		if len(spec.PriceCategories) > 0 {
			clone.Specs[i].PriceCategories = make(
				[]PriceCategoryOption,
				len(spec.PriceCategories),
			)
			copy(clone.Specs[i].PriceCategories, spec.PriceCategories)
		}
	}

	for i, spec := range clone.Specs {
		for _, a := range ad.Items {
			if spec.Code == a.Code {
				clone.Specs[i].CountOrAmount = a.Quantity
				for ipc, pc := range clone.Specs[i].PriceCategories {
					clone.Specs[i].PriceCategories[ipc].Checked = false
					if pc.Name == a.PriceCategory {
						clone.Specs[i].PriceCategories[ipc].Checked = true
					}
				}
			}
		}
	}

	return clone
}

type ProductFormSpec struct {
	Label           string
	Code            string
	CountOrAmount   int
	HasCategories   bool
	IsCustomAmount  bool
	PriceCategories []PriceCategoryOption
}

type PriceCategoryOption struct {
	Name    string `json:"name"`
	Price   int    `json:"price"`
	Checked bool   `json:"checked"`
}

type PriceKey struct {
	Code     string // e.g. "lunch"
	Category string // "regular", "" for custom
}

type ProductPrices map[PriceKey]int

type ProductModel struct {
	DB *sql.DB
}

func (m *ProductModel) GetProductIDMap() (ProductIDMap, error) {
	pidm := ProductIDMap{}
	products, err := m.GetAll()
	if err != nil {
		return nil, fmt.Errorf("Products.GetAll: %v", err)
	}
	for _, p := range products {
		var pricecat string
		if p.PriceCategory.Valid {
			pricecat = "/" + p.PriceCategory.String
		}
		pidm[p.Code+pricecat] = p.ID
	}

	return pidm, nil
}

func (m *ProductModel) GetProductFormConfig() (ProductFormConfig, error) {
	stmt := `
with product_form_specs as (
       select p.name,
              p.code,
              bool_or(p.price_category_id is not null) as has_categories,
              bool_or(p.pricing_mode = 'custom') as is_custom_amount,
              json_agg(
                json_build_object(
                  'name',    pc.name,
                  'price',   p.price,
                  'checked', (pc.name = 'regular')
                )
                order by pc.name
                )
                filter (where pc.name is not null and p.price is not null)
              as categories_json,
              min(pfo.sort_order) as sort_order
         from bellevue.products p
    left join bellevue.price_categories pc
           on pc.id = p.price_category_id
    left join bellevue.product_form_order pfo
           on pfo.code = p.code
        where p.deleted_at is null
        group by p.name, p.code
)
  select name,
         code,
         has_categories,
         is_custom_amount,
         coalesce(categories_json, '[]'::json) as categories_json
    from product_form_specs
order by sort_order nulls last, code
;
	`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return ProductFormConfig{}, fmt.Errorf("DB.Query(stmt): %v", err)
	}
	defer rows.Close()

	var pfc ProductFormConfig
	pfc.Prices = make(map[string]int)
	for rows.Next() {

		var catsJSON []byte
		var spec ProductFormSpec
		err = rows.Scan(
			&spec.Label,
			&spec.Code,
			&spec.HasCategories,
			&spec.IsCustomAmount,
			&catsJSON,
		)
		if err != nil {
			return ProductFormConfig{}, fmt.Errorf("for rows.Next(): %v", err)
		}

		var categories []PriceCategoryOption
		if err := json.Unmarshal(catsJSON, &categories); err != nil {
			return ProductFormConfig{}, fmt.Errorf("unmarshal categories for %s: %w", spec.Code, err)
		}
		spec.PriceCategories = categories
		pfc.Specs = append(pfc.Specs, spec)
		for _, category := range categories {
			pfc.Prices[spec.Code+"/"+category.Name] = category.Price
		}
	}

	if err = rows.Err(); err != nil {
		return ProductFormConfig{}, fmt.Errorf("rows.Err(): %v", err)
	}

	return pfc, nil
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
