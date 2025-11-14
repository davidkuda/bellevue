PG_DSN = postgres://davidkuda:@${DB_ADDRESS}/${DB_NAME}?sslmode=disable

db/migrate/newsql:
	@migrate create \
	-seq \
	-ext=.sql \
	-dir=./migrations \
	${name}

db/migrate/up-all:
	@migrate \
	-path=./migrations \
	-database=${PG_DSN} \
	up

db/migrate/version:
	migrate \
	-path=./migrations/ \
	-database=${PG_DSN} \
	version

# force V: Set version V but don't run migration (ignores dirty state)
db/migrate/force:
	@migrate \
	-path=./migrations/ \
	-database=${PG_DSN} \
	force ${version}

# migrate down one step
db/migrate/down-1:
	@migrate \
	-path=./migrations/ \
	-database=${PG_DSN} \
	down 1

# migrate up one step
db/migrate/up-1:
	@migrate \
	-verbose \
	-path=./migrations/ \
	-database=${PG_DSN} \
	up 1

