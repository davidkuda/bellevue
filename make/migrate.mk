migrate/newsql:
	@migrate create \
	-seq \
	-ext=.sql \
	-dir=./migrations \
	${name}

migrate/up-all:
	@migrate \
	-path=./migrations \
	-database=${PG_DSN_DEVELOPER} \
	up

migrate/version:
	@migrate \
	-path=./migrations/ \
	-database=${PG_DSN_DEVELOPER} \
	version

# force V: Set version V but don't run migration (ignores dirty state)
migrate/force:
	@migrate \
	-path=./migrations/ \
	-database=${PG_DSN_DEVELOPER} \
	force ${version}

# migrate down one step
migrate/down-1:
	@migrate \
	-path=./migrations/ \
	-database=${PG_DSN_DEVELOPER} \
	down 1

# migrate up one step
migrate/up-1:
	@migrate \
	-verbose \
	-path=./migrations/ \
	-database=${PG_DSN_DEVELOPER} \
	up 1

