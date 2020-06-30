# Movies

This example shows how to leverage Okteto to deploy and develop a Node + React application in Kubernetes.

This chart creates:
- A mongo database
- A deployment to serve the frontend
- A deployment to run the API
- An ingress to serve requests, leveraging Okteto Cloud's automatic SSL endpoints for public access.
