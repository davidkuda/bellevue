begin;

set role developer;

/**
 * restore:
 * - consumptions.user_id
 * - consumptions.date
 * - consumptions.invoice_id
 */
alter table consumptions
add column user_id int references users(id),
add column invoice_id int references invoices_v2(id),
add column "date" date;

UPDATE bellevue.consumptions c
SET
  user_id = a.user_id,
  "date" = a."date",
  invoice_id = a.invoice_id
FROM bellevue.activities a
WHERE c.activity_id = a.id;

alter table consumptions
drop column activity_id;

/**
 * restore: table comments
 */
create table bellevue.comments (
	user_id    INT  not null references bellevue.users(id),
	date       DATE not null,
	comment    TEXT,
	created_at TIMESTAMPTZ default now() not null,
	updated_at TIMESTAMPTZ default now() not null,

	primary key (user_id, date)
);

insert into comments (
	user_id, date, comment, created_at, updated_at
)
 select user_id, date, comment, created_at, updated_at
   from activities
  where activities.comment is not null;

drop table activities;

commit;
