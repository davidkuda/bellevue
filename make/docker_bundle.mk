bundle/js:
	node_modules/esbuild/bin/esbuild \
	--bundle \
	./ui/static/js/bundle.js \
	--outfile=./ui/static/dist/app.js

bundle/css:
	node_modules/esbuild/bin/esbuild \
	--bundle \
	./ui/static/css/bundle.css \
	--outfile=./ui/static/dist/styles.css
