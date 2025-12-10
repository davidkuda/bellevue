# TODO: Add a check whether the tables already have been created

psql -f migrations/000001_create_roles.up.sql
psql -f migrations/000002_data_model_products.up.sql
psql -f migrations/000003_populate_product_tables.up.sql