## Preview demo using a database snapshot

This will demonstrate how to use [Volume Snapshots](https://www.okteto.com/docs/enterprise/administration/volume-snapshots/) with Okteto. The PostgreSQL database has fake "production" user data in it that is loaded on startup. We can configure the preview environments to skip loading that data and use a snapshot with data for development instead.

### Pre-Requisites

- The Okteto cluster is prepared for Volume Snapshots, and the feature is enabled. See [Volume Snapshots documentation](https://www.okteto.com/docs/enterprise/administration/volume-snapshots/).
- A VolumeSnapshot of the postgresql has been created. See [Creating a test database snapshot](creating-db-snapshot.md) for an example.

### Update preview YAML to use a snapshot

- Create a branch
- Make two modifications to `.github/workflows/preview.yaml`:

  - specify the `okteto-with-volumes.yaml` file for deployment
  - include variables to skip data loading, specify the custom db snapshot and namespace

  Example:

  ```diff
          name: pr-${{ github.event.number }}-cindylopez
          scope: global
  -
  +        file: "okteto-with-volumes.yaml"
  +        variables: "API_LOAD_DATA=false,DB_SNAPSHOT_NAME=dbdata-snapshot,DB_SNAPSHOT_NAMESPACE=movies-test"
  ```

- Push this change and show how the API container did not load the data
- Show the data in the database is from the snapshot, and not the fake "production" data
