begin;

alter table consumptions
add column user_id int references users(id),
add column "date" date;

UPDATE bellevue.consumptions c
SET
  user_id = a.user_id,
  "date" = a."date"
FROM bellevue.activities a
WHERE c.activity_id = a.id;

alter table consumptions
drop column activity_id;

drop table activities;

commit;
