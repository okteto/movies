# Creating a database snapshot

### Pre-requisite

You must first prepare the Okteto cluster and enable Volume Snapshots. See the official documention for [Volume Snapshots](https://www.okteto.com/docs/enterprise/administration/volume-snapshots/) for more information.

### Prepare data

Before creating a database snapshot you will need to create an instance of the database and modify the data for use in dev environments. This could include adding or removing data needed for tests, or removing sensitive data from the database to make it safe to use in development environments.

1. Create a namespace to work in while creating the source of the snapshot:

   ```
   $ okteto namespace create movies-source
   ```

2. Deploy the environment with `okteto deploy`
3. Create a port-foward to the database service

   ```
   $ kubectl -n movies-source port-forward service/postgresql 5432:5432
   ```

4. In another terminal, connect to database using psql:

   ```
   $ psql -h localhost -p 5432 -U okteto -d votes
   Password for user okteto:  #The password is "okteto"
   psql (14.5 (Ubuntu 14.5-0ubuntu0.22.04.1), server 14.4)
   Type "help" for help.

   votes=>
   ```

5. Modify the data as needed for your purposes. In this example, we'll just update the last-name column to a static value but you can make any needed modifications here.

   ```
   votes=> UPDATE users SET last_name='testing-snapshot';
   ```

### Create VolumeSnapshot

This example YAML will use the source PersistentVolumeClaim for the postgresql service (`data-postgresql-0`) to create a VolumeSnapshot called "dbdata-snapshot" in the "movies-source" namespace.

`snapshot.yml`:

```yml
apiVersion: snapshot.storage.k8s.io/v1beta1
kind: VolumeSnapshot
metadata:
  namespace: movies-source
  name: dbdata-snapshot
spec:
  volumeSnapshotClassName: okteto-snapshot-class
  source:
    persistentVolumeClaimName: data-postgresql-0
```

1. Apply the snapshot YAML to the cluster with the modified database you want to act as a source for your snapshot:

   ```
   $ kubectl apply -f snapshot.yml
   ```

2. Confirm that the VolumeSnapshot exists and that READYTOUSE is true:

   ```
   $ kubectl -n movies-source get VolumeSnapshot
   NAME                       READYTOUSE   SOURCEPVC           SOURCESNAPSHOTCONTENT                  RESTORESIZE   SNAPSHOTCLASS           SNAPSHOTCONTENT                                    CREATIONTIME   AGE
   dbdata-snapshot            true         data-postgresql-0                                          1Gi           okteto-snapshot-class   snapcontent-67f7c5cb-658f-49fc-876d-bb16b7aa38ca   4d             4d
   ```

Your snapshot is now ready to be used as a source for development or preview environments.
