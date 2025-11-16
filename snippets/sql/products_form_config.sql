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


with product_form_specs as (
       select p.name,
              p.code,
              bool_or(p.pricing_mode = 'custom') as is_custom_amount,
              bool_or(p.price_category_id is not null) as has_categories,
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
         is_custom_amount,
         has_categories,
         coalesce(categories_json, '[]'::json) as categories_json
    from product_form_specs
order by sort_order nulls last, code
;
