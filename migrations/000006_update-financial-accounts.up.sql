begin;

set role developer;

update products set financial_account_id = 3 where code = 'kiosk';
update products set financial_account_id = 5 where code = 'sauna';

alter table financial_accounts
rename column description to view_name;

update financial_accounts
set view_name = 'Kiosk'
where id = 3;

insert into financial_accounts (
	tax_id,
	code,
	name,
	view_name
) values (
	(select id from taxes where mwst_satz = 0),
	3410,
	'Spenden - Diverse',
	'Spenden'
);

update products
set financial_account_id = (
	select id
	from financial_accounts
	where code = 3410
)
where code = 'donations';

commit;
