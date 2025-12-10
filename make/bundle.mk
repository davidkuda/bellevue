.PHONY: bundle
bundle: bundle/js/minify bundle/css/minify

bundle/js:
	@./node_modules/.bin/esbuild \
	--bundle \
	./ui/static/js/bundle.js \
	--outfile=./ui/static/dist/app.js

bundle/js/watch:
	@./node_modules/.bin/esbuild \
	--bundle \
	./ui/static/js/bundle.js \
	--outfile=./ui/static/dist/app.js \
	--watch

bundle/js/minify:
	@./node_modules/.bin/esbuild \
	--bundle \
	--minify \
	./ui/static/js/bundle.js \
	--outfile=./ui/static/dist/app.js

bundle/css:
	@./node_modules/.bin/esbuild \
	--bundle \
	./ui/static/css/bundle.css \
	--outfile=./ui/static/dist/styles.css

bundle/css/watch:
	@./node_modules/.bin/esbuild \
	--bundle \
	./ui/static/css/bundle.css \
	--outfile=./ui/static/dist/styles.css \
	--watch

bundle/css/minify:
	@./node_modules/.bin/esbuild \
	--bundle \
	--minify \
	./ui/static/css/bundle.css \
	--outfile=./ui/static/dist/styles.css
