-- array_agg
-- array_to_string
   select p.code,
          bool_or(p.price_category_id is not null) as has_categories,
          bool_or(p.pricing_mode = 'custom') as is_custom_amount,
          array_agg(distinct pc.name order by pc.name) filter (
            where pc.name is not null
          ) as price_categories,
          coalesce(
            array_to_string(
              array_agg(distinct pc.name order by pc.name)
                filter (where pc.name is not null),
              ','
            ),
            ''
          ) as price_categories_str
     from bellevue.products p
left join bellevue.price_categories pc
       on pc.id = p.price_category_id
    where p.deleted_at is null
    group by p.code
    order by p.code;
