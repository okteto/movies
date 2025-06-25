# Argo Rollouts Integration

This application has been refactored to use Argo Rollouts instead of standard Kubernetes deployments for advanced deployment strategies.

## Services and Strategies

- **Frontend**: Canary deployment with manual promotion steps
- **Catalog**: Blue-Green deployment with analysis templates
- **API**: Blue-Green deployment with analysis templates  
- **Rent**: Canary deployment with automatic progression
- **Worker**: Simple canary deployment (background service)

## Deployment

Deploy the application as usual:

```bash
okteto deploy --wait
```

This will:
1. Install Argo Rollouts controller
2. Deploy all services with their respective rollout strategies
3. Apply analysis templates for automated quality gates

## Managing Rollouts

### View Rollout Status

```bash
# List all rollouts
kubectl argo rollouts list -n ${OKTETO_NAMESPACE}

# Get detailed status of a specific rollout
kubectl argo rollouts get rollout frontend -n ${OKTETO_NAMESPACE}
kubectl argo rollouts get rollout catalog -n ${OKTETO_NAMESPACE}
kubectl argo rollouts get rollout api -n ${OKTETO_NAMESPACE}
kubectl argo rollouts get rollout rent -n ${OKTETO_NAMESPACE}
kubectl argo rollouts get rollout worker -n ${OKTETO_NAMESPACE}
```

### Manual Promotion

For services with manual promotion (frontend, catalog, api):

```bash
# Promote to next step
kubectl argo rollouts promote frontend -n ${OKTETO_NAMESPACE}
kubectl argo rollouts promote catalog -n ${OKTETO_NAMESPACE}
kubectl argo rollouts promote api -n ${OKTETO_NAMESPACE}
```

### Abort Rollout

If issues are detected:

```bash
kubectl argo rollouts abort frontend -n ${OKTETO_NAMESPACE}
kubectl argo rollouts abort catalog -n ${OKTETO_NAMESPACE}
kubectl argo rollouts abort api -n ${OKTETO_NAMESPACE}
kubectl argo rollouts abort rent -n ${OKTETO_NAMESPACE}
kubectl argo rollouts abort worker -n ${OKTETO_NAMESPACE}
```

### Rollback

```bash
kubectl argo rollouts undo frontend -n ${OKTETO_NAMESPACE}
kubectl argo rollouts undo catalog -n ${OKTETO_NAMESPACE}
kubectl argo rollouts undo api -n ${OKTETO_NAMESPACE}
kubectl argo rollouts undo rent -n ${OKTETO_NAMESPACE}
kubectl argo rollouts undo worker -n ${OKTETO_NAMESPACE}
```

## Rollout Strategies Explained

### Canary Deployment (Frontend, Rent, Worker)
- Gradually shifts traffic to new version
- Frontend: Manual promotion with 20%, 40%, 60%, 80% steps
- Rent: Automatic progression with 25%, 50%, 75% steps (30s pauses)
- Worker: Simple 50% split with 60s pause

### Blue-Green Deployment (Catalog, API)
- Deploys new version alongside current version
- Uses preview services for testing
- Requires manual promotion after analysis
- Automatic rollback on analysis failure

## Analysis Templates

The following analysis templates are available:
- `success-rate`: Monitors HTTP success rate (>95% required)
- `response-time`: Monitors 95th percentile response time (<500ms required)

Note: Analysis templates require Prometheus metrics. If Prometheus is not available, you can remove the analysis sections from the rollout configurations.

## Troubleshooting

### Check Rollout Events
```bash
kubectl describe rollout <service-name> -n ${OKTETO_NAMESPACE}
```

### View Rollout Logs
```bash
kubectl logs -l app.kubernetes.io/name=<service-name> -n ${OKTETO_NAMESPACE}
```

### Check Argo Rollouts Controller
```bash
kubectl logs -n argo-rollouts deployment/argo-rollouts-controller
```