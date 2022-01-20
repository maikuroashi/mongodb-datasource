# Grafana Data Source Backend Plugin For MongoDB

This is an implementation of a Grafana Data Source Backend Plugin to allow Grafana to query and visualize data from a MongoDB database. A data source backend plugin consists of both frontend and backend components. For more information about backend plugins, refer to the documentation on [Backend plugins](https://grafana.com/docs/grafana/latest/developers/plugins/backend/).

## Requirements

* Grafana > 7.x
* MongoDB > 2.6

* Development tools

  * Make > 3.x
  * Node > 16.x
  * NPM  > 8.x
    * [Install](https://nodejs.org/en/download/)
  * Go > go1.17.6
    * [Install](https://go.dev/doc/install)
  * Mage > 1.29.2
    * [Install](https://magefile.org/)
  * Docker > 20.0
    * [Install](https://docs.docker.com/engine/install/)
  * Docker Compose > 1.29
    * [Install](https://docs.docker.com/compose/install/)

## Build

The front end and back end components can be built into the `dist` directory by running:

```sh
make all
```

## Installation

The plugin can be installed in Grafana by copying the contents of the `dist` directory into a sub directory, usual named after the plugin e.g. `mongodb-datasource`, in the Grafana plugin directory.
For security reasons Grafana will, by default, not load an unsigned plugin so it is necessary to define the following environment variable with the plugin identifier:

```sh
GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=maikuroashi-mongodb-datasource
```

## Query Syntax

The plugin aims to support a subset of the query syntax provided by the MongoDb shell. Both `find` and `aggregation` queries are supported. However, one key difference is that the documents passed to the query must be valid JSON with the exception of limited support for literal values e.g. `new Date(1622353314804)`.

Grafana defines a number of global variables that can be substituted into a query using the `${}` syntax before it is passed to the backend plugin. The `$__from` and `$__to` variables allow the dashboard's current date range to be integrated into a query. For further information refer to the Grafana [Global Variables](https://grafana.com/docs/grafana/latest/variables/variable-types/global-variables/) documentation.

The following are some examples of the query syntax, including using Grafana global variables to refer to the dashboard's date range. For further information about MongoDB queries see the [Mongo DB Documentation](https://docs.mongodb.com/manual/tutorial/query-documents/).

```javascript
//Find all documents in the employees collection of the default DB
db.employees.find()

//Find documents in the employees collection of the default DB with a first name of "Bob"
db.employees.find({"firstName": "Bob"})

//Find all documents in the employees collection of the default DB and return just the "lastName"
db.employees.find({}, {"_id": 0, "lastName": 1})

//Find all documents in the employees collection of the default DB and order the results by "lastName"
db.employees.find({}, {}).sort({"lastName": 1});

//Find all documents in the employees collection of the default DB within a given date range.
db.employees.find({"startDate" : { "$gte": new Date($__from), "$lt": new Date($__to) }})

//Find all documents in the products collection of the sales database. 
sales.products.find();

//Find all documents in the employees collection of the default DB with a first name of "Bob" and order results by "lastName"
db.employees.aggregate([
  {
    "$match": {
      "firstName": "Bob",
    }
  },
  {
    "$sort": {
      "$lastName": 1,
    }
  }
])
```

## Development

The `dockerdev` directory contains a `docker-compose.yaml` file which can be used to launch an instance of Grafana with the plugin installed and a MongoDB database instance. The Grafana UI is exposed on the host at port `3000` and MongoDb on the default port of `27017`.

The plugin can be configured to connect to the Docker MongoDb instance by selecting the `Configuration (Cog) / Datasources` page from the side-bar, clicking the `Add data source` button, and then entering the following details in the configuration page:

```properties
Name: MongoDb
URL: mongodb://mongo:27017
Database: test
User: root
Password: example
```

### Remote Debugging

The compose file builds a custom docker image that extends the base Grafana image to include the Go Lang development tools to permit remote debugging of the backend plugin. Port `2345` is mapped to the host to allow access to the Go remote debugger `dvl` which can be launched in the container with the following commands:

```sh
docker exec -it grafana bash
dlv attach --headless --listen=:2345 --log --api-version=2 $(pgrep gpx_mongodb-datasource_linux_amd64)
```

Once the remote debugger is running you can attach vscode and debug using the `Connect to server` launch configuration.

### Unit Tests

Run unit tests and generate a coverage report with the following commands:

```sh
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```
