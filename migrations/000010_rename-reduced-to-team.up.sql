set role developer;

begin;

update price_categories
set name = 'team', updated_at = now()
where name = 'reduced';

commit;
