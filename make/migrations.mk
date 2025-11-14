PG_DSN_ADMIN = postgres://davidkuda:@${DB_ADDRESS}/${DB_NAME}?sslmode=disable

PG_DSN_APP = postgres://${DB_USER}:${DB_PASSWORD}@${DB_ADDRESS}/${DB_NAME}?sslmode=disable

db/migrate/newsql:
	@migrate create \
	-seq \
	-ext=.sql \
	-dir=./migrations \
	${name}

db/migrate/up-all:
	@migrate \
	-path=./migrations \
	-database=${PG_DSN_ADMIN} \
	up

db/migrate/version:
	migrate \
	-path=./migrations/ \
	-database=${PG_DSN_ADMIN} \
	version

db/migrate/force:
	@migrate \
	-path=./migrations/ \
	-database=${PG_DSN_ADMIN} \
	force ${version}

# migrate down one step
db/migrate/down-1:
	@migrate \
	-path=./migrations/ \
	-database=${PG_DSN_ADMIN} \
	down 1

# migrate up one step
db/migrate/up-1:
	@migrate \
	-verbose \
	-path=./migrations/ \
	-database=${PG_DSN_ADMIN} \
	up 1

