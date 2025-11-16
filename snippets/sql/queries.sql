-- show all acitivities
select activity_date, email, total_price
from bellevue_origins b
join auth.users u on u.id = b.user_id;

-- show activities by user
select activity_date, email, total_price
from bellevue_origins b
join auth.users u on u.id = b.user_id
where email = 'test@kuda.ai';


-- show consumptions:
select date, u.first_name, p.name, pc.name as pricecat, c.unit_price, c.quantity, c.total_price
from consumptions c
join products p on product_id = p.id
join auth.users u on c.user_id = u.id
join price_categories pc on c.pricecat_id = pc.id;

