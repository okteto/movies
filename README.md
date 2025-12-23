# Movies App

This example shows how to leverage [Okteto](https://github.com/okteto/okteto) to develop an application based on microservices directly on Kubernetes. The Movies App is deployed using a Helm Charts. It creates the following components:

- A *React* based [frontend](frontend) service, using [webpack](https://webpack.js.org) as bundler and *hot-reload server* for development
- A Node.js based [catalog](catalog) service to serve the available movies from a MongoDB database
- A Java based [rent](rent) service to receive rent requests and send them to Kafka
- A Golang based [worker](worker) to process rent request from Kafka and update the PostgreSQL database
- A Golang based [api](api) to retrieve the current movies rentals from the PostgresSQL database
- A [MongoDB](https://bitnami.com/stack/mongodb/helm) database
- A [Kafka](https://bitnami.com/stack/kafka/helm) queue
- A [PostgresQL](https://bitnami.com/stack/postgresql/helm) database

![Architecture diagram](docs/architecture-diagram.png)


## Setup Auth0 Auth

1. Create an Auth0 account (or login into your existing account)
1. Create an application in Auth0. Set the following values, replacing `okteto.example.com` with your Okteto Instance's domain.

    - `Allowed Call Back URLs`: `https://*.okteto.example.com`
    - `Allowed Logout URLs`: `https://*.okteto.example.com`
    - `Allowed Web Origins`: `https://*.okteto.example.com`

1. Create the following admin variables in Okteto:

    - `AUTH0_CLIENT_ID`: The client ID of your 
    - `AUTHO_DOMAIN`: Your auth0 domain e.g. example.auth0.com
    - `AUTHO_AUDIENCE`: The audience of the token, since we are calling Auth0 API it should be https://<your auth0 domain>/api/v2/

## Deploy the Application

Execute `okteto deploy` to deploy it from the terminal. You can also deploy it from the UI. 


## Development container demo script

- Deploy the repo from UI
- Rent two movies
- `okteto up worker` + `make build` + `make start`
- Uncomment line 61 in `rentals/cmd/worker/main.go`
- `make build` + `make start`
- Show how the change is applied
