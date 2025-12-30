BEGIN;

insert into bellevue.users (
	id,
	first_name,
	last_name,
	email,
	method,
	hashed_password
)
select	id,
	first_name,
	last_name,
	email,
	'password',
	hashed_password
from	auth.users
where	id != 1;

COMMIT;
