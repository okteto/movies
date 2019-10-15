# Movies Sample app

This example shows how to leverage [Okteto](https://github.com/okteto/okteto) to develop a Node + React Sample App directly in Kubernetes. The Node + React Sample App is deployed using a [Helm 3  chart](https://github.com/okteto/charts/tree/master/movies). It creates the following components:

- A *React* based front-end, using [webpack](https://webpack.js.org) as bundler and *hot-reload server* for development.
- A very simple Node.js API using [Express](https://expressjs.com).
- A Node.js job to load the initial data into MongoDB.
- A [MongoDB](https://www.mongodb.com) database.

Okteto works in any Kubernetes cluster by reading your local Kubernetes credentials. For a empowered experience, follow this [guide](https://okteto.com/docs/samples/node/) to deploy the Node + React Sample App in [Okteto Cloud](https://cloud.okteto.com), a free-trial Kubernetes cluster.


## Step 1: Install the Okteto CLI

Install the Okteto CLI by following our [installation guides](https://github.com/okteto/okteto/blob/master/docs/installation.md).


## Step 2: Launch the application

Install the latest release oof Helm 3 by following the [instructions available here](https://v3.helm.sh/docs/intro/install/).


Add the Okteto chart repository to your client:

```console
$ helm repo add okteto https://apps.okteto.com/
```

```console
"okteto" has been added to your repositories.
```

And install the movies application:
```console
$ helm install movies okteto/movies
```

```console
NAME: movies
LAST DEPLOYED: Mon Oct 14 16:55:09 2019
NAMESPACE: okteto
STATUS: deployed
REVISION: 1
NOTES:
Success! Your application will be available shortly.

Get the application URL by running this command:
kubectl get ingress movies
```

After a few seconds, your application will be ready to go. 

> If you're using Okteto Cloud, you can deploy this application directly from the UI: Deploy > movies > Deploy 

## Step 3: Create your Okteto Environment for the frontend

Move to the movies front-end code directory:

```console
$ cd frontend
```

And launch your Okteto environment by running  the command below:

```console
$ okteto up
````

```console
 ✓  Okteto Environment activated
 ✓  Files synchronized
 ✓  Your Okteto Environment is ready
    Namespace: cindy
    Name:      movies-frontend
    Forward:   8080 -> 8080

root@movies-frontend-8c8997bd6-h5rq5:/src#
```

The `okteto up` command will automatically start an Okteto Environment. It will also start a *file synchronization service* to keep your changes up to date between your local filesystem and your Okteto Environment, without rebuilding containers (eliminating the docker build/push/pull/redeploy cycle).

Once the Okteto Environment is ready, the Okteto Terminal will automatically open. Use it to run your frontend with the same flow you would have locally:

```console
okteto> yarn start 
```

```console
[1/4] Resolving packages...
[2/4] Fetching packages...
...
...
```

This will compile and run webpack-dev-server listening on port 8080.

The frontend of your application is now ready and in development mode. You can access it at http://localhost:8080.

## Step 4: Develop directly in Kubernetes

Now things get even more exciting. You can now develop *directly in your Kubernetes cluster*. The API service and database will be available at all times. No need to mock services nor use any kind of redirection.
 
In your IDE edit the file `frontend/src/App.jsx` and change the `Okteflix` text in line 92 to `Netflix`. Save your changes.

Go back to the browser, and cool! Your changes are automatically live with no need to refresh your browser. Everything happened in the cluster but no commit or push was required 😎!

<p align="center"><img src="frontend/static/okteflix.gif" width="650" /></p>

## Step 5: Cleanup

Cancel the `okteto up` command by pressing `ctrl + c` + `exit` and run the following commands to remove the resources created by this guide: 

```console
$ okteto down
 ✓  Okteto Environment deactivated
 
```

```console
$ helm uninstall movies
```

```console
release "movies" uninstalled
```
