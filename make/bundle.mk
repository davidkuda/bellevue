bundle/js:
	esbuild \
	--bundle \
	./ui/static/js/bundle.js \
	--outfile=./ui/static/dist/app.js

bundle/js/watch:
	esbuild \
	--bundle \
	./ui/static/js/bundle.js \
	--outfile=./ui/static/dist/app.js \
	--watch

bundle/js/minify:
	esbuild \
	--bundle \
	--minify \
	./ui/static/js/bundle.js \
	--outfile=./ui/static/dist/app.js

bundle/css:
	esbuild \
	--bundle \
	./ui/static/css/bundle.css \
	--outfile=./ui/static/dist/styles.css

bundle/css/watch:
	esbuild \
	--bundle \
	./ui/static/css/bundle.css \
	--outfile=./ui/static/dist/styles.css \
	--watch

bundle/css/minify:
	esbuild \
	--bundle \
	--minify \
	./ui/static/css/bundle.css \
	--outfile=./ui/static/dist/styles.css
