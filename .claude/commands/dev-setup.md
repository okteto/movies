Set up the Okteto development environment for this project. Follow these steps:

1. **Check prerequisites**:
   - Run `okteto version` to confirm the CLI is installed
   - Run `okteto context show` to confirm the user is connected to an Okteto instance
   - If either fails, stop and help the user install/configure Okteto

2. **Deploy all services**:
   - Run `okteto deploy --wait`
   - This builds images and deploys all microservices (frontend, catalog, rent, worker, api) plus infrastructure (PostgreSQL, Kafka, MongoDB)
   - Wait for it to complete successfully

3. **Show the running environment**:
   - Run `okteto endpoints` to display the public URLs
   - Share the URLs with the user so they can open the app in their browser

4. **Guide the user to start development**:
   - Ask which service they want to work on (frontend, catalog, rent, worker, or api)
   - Tell them to run `okteto up <service>` **in their terminal** (this is interactive — do not run it yourself)
   - Once they confirm it's running, remind them of the dev commands for that service:
     - **frontend**: `yarn install && yarn start`
     - **catalog**: starts automatically
     - **rent**: starts automatically
     - **api**: `make build && make start`
     - **worker**: `make build && make start`

5. **Confirm readiness**:
   - Let the user know you can now help with: running tests (`okteto exec -- <cmd>`), debugging errors, reading code, and analyzing logs (`okteto logs <service>`)
