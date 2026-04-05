begin;

set role developer;

alter table consumptions
add column tax_id  INT references taxes(id),
add column pricecat_id INT references price_categories(id);

update consumptions c
set tax_id = p.tax_id,
    pricecat_id = p.price_category_id
from products p
where c.product_id = p.id;

alter table consumptions
alter column tax_id set not null;

commit;
