# Movies App - AI Agent Guide

## Project Overview
Microservices-based Movies rental application demonstrating Okteto development on Kubernetes. Deployed via Helm charts. The agent can deploy all services with a single `okteto deploy`, then help the user debug any individual service.

## Architecture

### Services
- **frontend**: React app with webpack hot-reload
- **catalog**: Node.js service serving movies from MongoDB
- **rent**: Java service handling rent requests → Kafka
- **worker**: Go service processing Kafka messages → PostgreSQL
- **api**: Go service retrieving rentals from PostgreSQL

### Data Stores
- MongoDB: Movie catalog
- Kafka: Message queue for rent requests
- PostgreSQL: Rental records

## Key Files
- `okteto.yaml`: Okteto manifest (source of truth for build/deploy/dev)
- Each service has its own directory with independent build setup
- Helm charts in each service's `chart/` directory

## For AI Agents

### Setting Up the Environment

When a user asks to set up or deploy their development environment:

1. **Check prerequisites**: Run `okteto version` and `okteto context show`
2. **Deploy**: Run `okteto deploy --wait` to build images and deploy all services
3. **Show endpoints**: Run `okteto endpoints` so the user can access the app
4. **Guide the user** to start development on a specific service with `okteto up <service>`

### The `okteto up` Rule

**`okteto up` is interactive and must be run by the user in their terminal.** It opens a shell inside the development container with live file sync. Never run it as a background task. Instead, tell the user which command to run:

```
Run this in your terminal: okteto up worker
```

### Collaborative Workflow

Once the user has `okteto up <service>` running in their terminal:

- **Run diagnostics**: Use `okteto exec -- <command>` to execute commands in the dev container and see output
- **Read synced files**: Use the Read tool to examine code that's syncing to the cluster
- **Analyze pasted output**: When the user hits an error, they can paste terminal output for analysis
- **Check logs**: `okteto logs <service>` shows container logs (separate from the interactive session)

You're facilitating their development workflow, not trying to observe their terminal session.

### Example Interaction

```
Developer: "I'm working on the worker service, my tests are failing"

Agent actions:
  1. okteto exec -- make build        → builds the Go binary in the container
  2. okteto exec -- go test ./...     → runs tests, captures output
  3. Reads the failing test file and source code
  4. Identifies the bug and suggests a fix

Developer applies the fix (auto-syncs to container), agent re-runs tests to confirm.
```

### Debugging Patterns

- **User asks to run tests**: `okteto exec -- make test` or `okteto exec -- go test ./...`
- **User pastes an error**: Read relevant code files, analyze, suggest fix
- **User asks "why is this failing?"**: Run diagnostic commands via `okteto exec`
- **User makes code changes**: Changes auto-sync; help them understand what to run next

## Service-Specific Dev Commands

Once inside a development container (`okteto up <service>`):

| Service | Language | Build & Run |
|---------|----------|-------------|
| **frontend** | Node.js | `yarn install && yarn start` |
| **catalog** | Node.js | `yarn start` (auto-starts via dev command) |
| **rent** | Java | `mvn spring-boot:run` (auto-starts via dev command) |
| **api** | Go | `make build && make start` |
| **worker** | Go | `make build && make start` |

## Okteto CLI Quick Reference

| Command | Who runs it | Purpose |
|---------|-------------|---------|
| `okteto deploy --wait` | Agent | Build images and deploy all services |
| `okteto build` | Agent | Build and push images only |
| `okteto up <service>` | **User** | Start interactive dev container (never run as agent) |
| `okteto down` | Agent/User | Stop dev mode, restore deployment |
| `okteto exec -- <cmd>` | Agent | Run a command in the dev container |
| `okteto logs <service>` | Agent | View container logs |
| `okteto endpoints` | Agent | List public URLs |
| `okteto test` | Agent | Run e2e tests in the cluster |
| `okteto destroy` | User | Tear down all resources |

## Common Mistakes

- **Forgetting to deploy first**: Run `okteto deploy` before `okteto up` if services aren't already deployed
- **Not specifying the service**: With multiple services, always specify which one: `okteto up worker`
- **Using kubectl/helm directly**: Always use `okteto deploy` so Okteto can track resources
- **Building Docker images locally**: Use `okteto build` to leverage the Okteto Build Service

## Troubleshooting

- `okteto doctor`: Generate a diagnostic bundle
- `okteto status`: Check file sync progress
- `okteto logs --since 5m`: Recent service logs
- `okteto context show`: Verify cluster and namespace
