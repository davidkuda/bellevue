-- show all acitivities
select activity_date, email, total_price
from bellevue_origins b
join auth.users u on u.id = b.user_id;

-- show activities by user
select activity_date, email, total_price
from bellevue_origins b
join auth.users u on u.id = b.user_id
where email = 'test@kuda.ai';

