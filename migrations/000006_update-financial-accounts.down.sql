begin;

set role developer;

alter table financial_accounts
rename column view_name to description;

update financial_accounts
set description = 'Bier, Chips, Schokolade'
where id = 3;

-- update products ID not reversed with intent.
-- this was just a misstake.

delete from financial_accounts
where code = 3410;

update products
set financial_account_id = 1
where code = 'donations';

commit;
