begin;

set role developer;

update products set financial_account_id = 3 where code = 'kiosk';
update products set financial_account_id = 5 where code = 'sauna';

alter table financial_accounts
rename column description to view_name;

update financial_accounts
set view_name = 'Kiosk'
where id = 3;

commit;
