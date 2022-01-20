all: frontend backend

clean:
	rm -rf dist
	rm -rf node_modules

backend: dist/gpx_*

frontend: node_modules dist/module.js

.PHONY: all clean frontend backend

node_modules: package.json
	npm install

dist/module.js: src/**/*.* README.md
	npm run-script build

dist/gpx_*: pkg/**/*.go go.mod go.sum
	mage
