include ./make/bundle.mk
include ./make/migrations.mk
include ./make/postgres.mk

fmt/ui: fmt/ui/static fmt/ui/templates

fmt/ui/static:
	biome format ./ui/static/css/ --write
	biome format ./ui/static/js/ --write

fmt/ui/templates:
	./node_modules/prettier/bin/prettier.cjs \
	./ui/html/ \
	--write

