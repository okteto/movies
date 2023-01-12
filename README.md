# Movies Sample app

[![Develop on Okteto](https://okteto.com/develop-okteto.svg)](https://cloud.okteto.com/deploy?repository=https://github.com/okteto/movies)

This example shows how to leverage [Okteto](https://github.com/okteto/okteto) to develop a Node.js + React Sample App directly in Kubernetes. The Node + React Sample App is deployed using a [Helm Chart](https://github.com/okteto/movies/tree/main/chart). It creates the following components:

- A *React* based front-end, using [webpack](https://webpack.js.org) as bundler and *hot-reload server* for development.
- A very simple Node.js API using [Express](https://expressjs.com).
- A [MongoDB](https://www.mongodb.com) database.

This is the application used for the [Getting Started on Okteto Cloud in 5 minutes](https://www.okteto.com/docs/getting-started/) tutorial.

d
