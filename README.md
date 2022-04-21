# Movies Sample app

[![Develop on Okteto](https://okteto.com/develop-okteto.svg)](https://cloud.okteto.com/deploy?repository=https://github.com/okteto/movies)

This example shows how to leverage [Okteto](https://github.com/okteto/okteto) to develop a Node.js + React Sample App directly in Kubernetes. The Node + React Sample App is deployed using a [Helm Chart](https://github.com/okteto/movies/tree/main/chart). It creates the following components:

- A *React* based [frontend](frontend), using [webpack](https://webpack.js.org) as bundler and *hot-reload server* for development.
- A [rentals](rentals) service. It has a Java API, a worker in Go, a Kafka broker, and a Postgres database.
- A [catalog](catalog) service. It has a Node.js API serving data from a MongoDB database.

![Architecture diagram](architecture.png)
