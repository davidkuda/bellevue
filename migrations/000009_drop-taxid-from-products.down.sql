set role developer;

begin;

alter table products
add column tax_id int references taxes(id);

update products
set tax_id = fa.tax_id
from financial_accounts fa
where products.financial_account_id = fa.id;

alter table products
alter column tax_id set not null;

commit;
