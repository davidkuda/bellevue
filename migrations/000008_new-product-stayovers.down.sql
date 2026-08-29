set role developer;

begin;

delete from products
where code = 'stay/karmayoga';

delete from financial_accounts
where code = 3104;

delete from product_form_order
where code = 'stay/karmayoga';

commit;
