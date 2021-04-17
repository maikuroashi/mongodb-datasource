all: frontend backend

clean:
	rm -rf dist

backend: dist/gpx_*

frontend: node_modules dist/module.js

.PHONY: all clean frontend backend

node_modules:
	npm install

dist/module.js: src/**/*.*
	npm run-script build

dist/gpx_*: pkg/**/*.go
	mage

