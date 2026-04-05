begin;

set role developer;

alter table consumptions
drop column tax_id,
drop column pricecat_id;

commit;
