# Movies App

[![Develop on Okteto](https://okteto.com/develop-okteto.svg)](https://cloud.okteto.com/deploy?repository=https://github.com/okteto/movies)

This example shows how to leverage [Okteto](https://github.com/okteto/okteto) to develop an application based on microservices directly on Kubernetes. The Movies App is deployed using a Helm Charts. It creates the following components:

- A *React* based [frontend](frontend) service, using [webpack](https://webpack.js.org) as bundler and *hot-reload server* for development
- A Node.js based [catalog](catalog) service to serve the available movies from a MongoDB database
- A Java based [rent](rent) service to receive rent requests and send them to Kafka
- A Golang based [worker](worker) to process rent request from Kafka and update the PostgreSQL database
- A Golang based [api](api) to retrieve the current movies rentals from the PostgresSQL database
- A [MongoDB](https://bitnami.com/stack/mongodb/helm) database
- A [Kafka](https://bitnami.com/stack/kafka/helm) queue
- A [PostgresQL](https://bitnami.com/stack/postgresql/helm) database

![Architecture diagram](architecture-diagram.png)

## Development container demo script

- Deploy the repo from UI
- Rent two movies
- `okteto up worker` + `make build` + `make start`
- Uncomment line 61 in `rentals/cmd/worker/main.go`
- `make build` + `make start`
- Show how the change is applied

## (Optional) Preview demo using a database snapshot

This will demonstrate how to use [Volume Snapshots](https://www.okteto.com/docs/enterprise/administration/volume-snapshots/) with Okteto. The PostgreSQL database has fake "production" user data in it that is loaded on startup. We can configure the preview environments to skip loading that data and use a snapshot with data for development instead.

### Pre-Requisites

- The Okteto cluster is prepared for Volume Snapshots, and the feature is enabled. See [Volume Snapshots documentation](https://www.okteto.com/docs/enterprise/administration/volume-snapshots/).
- A VolumeSnapshot of the postgresql has been created. See [Creating a test database snapshot](docs/creating-db-snapshot.md) for an example.

### Update preview YAML to use a snapshot

- Create a branch
- Modify `.github/workflows/preview.yaml` to include variables to skip data loading, specify the custom db snapshot and namespace.

  Example:

  ```diff
          name: pr-${{ github.event.number }}-cindylopez
          scope: global
  -
  +        variables: "API_LOAD_DATA=false,DB_SNAPSHOT_NAME=dbdata-snapshot,DB_SNAPSHOT_NAMESPACE=movies-test"
  ```

- Push this change and show how the API container did not load the data
- Show the data in the database is from the snapshot, and not the fake "production" data
