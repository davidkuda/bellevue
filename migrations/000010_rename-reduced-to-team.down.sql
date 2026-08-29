set role developer;

begin;

update price_categories
set name = 'reduced', updated_at = now()
where name = 'team';

commit;
