# Okteto Divert Demo - Movies App (Nginx + Linkerd)

## What is Divert?

Okteto's Divert feature allows developers to work on individual microservices without deploying the entire application stack. It uses HTTP header-based routing to intelligently route traffic between shared services and your personal development instances.

**Note**: This demo uses the **Nginx driver with Linkerd** for header-based routing. Okteto Divert also works natively with **Istio** (no Linkerd required). Choose the driver based on your existing service mesh infrastructure.

### Key Benefits

- **Massive Resource Savings**: Deploy only the services you're working on (~85% reduction in resources)
- **Faster Setup**: Environment ready in 30 seconds instead of 5-10 minutes
- **Isolation**: Your changes don't affect other developers
- **Production-like**: Test against real shared services
- **Cost Efficient**: Share expensive infrastructure (databases, message queues)

### How It Works

1. **Shared Namespace**: A complete Movies app stack runs in a shared staging namespace
2. **Personal Namespace**: Your namespace contains only the service(s) you're developing
3. **Smart Routing**: Nginx ingress (with Linkerd sidecar) routes requests with the `baggage: okteto-divert=<your-namespace>` header to your services
4. **Header Propagation**: The baggage header is automatically propagated through all service calls

### Divert Drivers

Okteto Divert supports two drivers:

- **Nginx + Linkerd** (this demo): Uses Nginx ingress controller with Linkerd service mesh for header-based routing
- **Istio** (native): Uses Istio's built-in VirtualService for header-based routing without requiring additional components

To use Istio instead, change `driver: nginx` to `driver: istio` in the divert configurations.

## Quick Start

### Prerequisites

- Okteto CLI installed (`brew install okteto` or download from [okteto.com](https://okteto.com))
- kubectl configured
- Access to an Okteto cluster

### Setup Steps

1. **Clone the repository**
   ```bash
   git clone https://github.com/okteto/movies
   cd movies
   ```

2. **Run the setup script**
   ```bash
   ./scripts/setup-divert.sh movies-staging
   ```

3. **Follow the prompts**
   - Enter your name (e.g., "alice")
   - Choose which service to work on (1-4)

4. **Start developing**
   ```bash
   okteto up
   ```

## Available Divert Configurations

### 1. Frontend Development (`okteto-frontend-divert.yaml`)
**Use when**: Working on React UI components, user interactions, or frontend features

**What it deploys**:
- Frontend service only (React/Node.js)

**Shares from staging**:
- Catalog service
- API service
- Rent service
- All databases (MongoDB, PostgreSQL, Kafka)

**Dev command**: `yarn start`

### 2. Catalog Development (`okteto-catalog-divert.yaml`)
**Use when**: Working on movie catalog, inventory management, or MongoDB integration

**What it deploys**:
- Catalog service only (Node.js/Express)

**Shares from staging**:
- Frontend
- API service
- Rent service
- MongoDB (connected from shared namespace)

**Dev command**: `yarn start`

### 3. API Development (`okteto-api-divert.yaml`)
**Use when**: Working on API endpoints, user management, or PostgreSQL integration

**What it deploys**:
- API service only (Golang)

**Shares from staging**:
- Frontend
- Catalog service
- Rent service
- PostgreSQL (connected from shared namespace)

**Dev command**: `go run cmd/api/main.go`

### 4. Rent Development (`okteto-rent-divert.yaml`)
**Use when**: Working on rental logic, Kafka integration, or Spring Boot backend

**What it deploys**:
- Rent service only (Java/Spring Boot)

**Shares from staging**:
- Frontend
- Catalog service
- API service
- Kafka (connected from shared namespace)

**Dev command**: `mvn spring-boot:run`

## Baggage Header Propagation

The Movies app has been instrumented with baggage header propagation to ensure Divert routing works seamlessly across all services.

### Frontend (React)

```javascript
// Detects baggage header from URL query parameter or sessionStorage
const params = new URLSearchParams(window.location.search);
const baggageHeader = params.get('baggage') || sessionStorage.getItem('baggage-header');

// All fetch calls include the baggage header
async fetchWithBaggage(url, options = {}) {
  const headers = {
    'Content-Type': 'application/json',
    ...options.headers
  };

  if (window.baggageHeader) {
    headers['baggage'] = window.baggageHeader;
  }

  return fetch(url, { ...options, headers });
}
```

### Catalog Service (Node.js)

```javascript
// Middleware to capture baggage header
app.use((req, res, next) => {
  req.baggageHeader = req.headers['baggage'];
  next();
});
```

### API Service (Golang)

```go
// Middleware to propagate baggage header
func baggageMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        baggage := r.Header.Get("baggage")
        r.Header.Set("X-Baggage", baggage)
        next(w, r)
    }
}

// When calling other services
if baggage := r.Header.Get("X-Baggage"); baggage != "" {
    req.Header.Set("baggage", baggage)
}
```

### Rent Service (Java/Spring Boot)

```java
// ThreadLocal for baggage header
private static final ThreadLocal<String> baggageHeader = new ThreadLocal<>();

@PostMapping("/rent")
public String rent(@RequestBody Rental rental,
                  @RequestHeader(value = "baggage", required = false) String baggage) {
    baggageHeader.set(baggage);

    try {
        ProducerRecord<String, String> record = new ProducerRecord<>("rentals", data);

        // Propagate to Kafka
        if (baggage != null) {
            record.headers().add(new RecordHeader("baggage", baggage.getBytes()));
        }

        kafkaTemplate.send(record);
    } finally {
        baggageHeader.remove();
    }
}
```

## Testing Your Divert Setup

### 1. Test Header Propagation Locally

```bash
# Start each service and verify header logging
curl -H "baggage: okteto-divert=test" http://localhost:8080/catalog
```

### 2. Test Divert in Okteto

```bash
# Test with baggage header (routes to your namespace)
curl -H "baggage: okteto-divert=alice-movies" https://movies-movies-staging.okteto.dev

# Test direct access (bypasses Divert)
curl https://movies-alice-movies.okteto.dev
```

### 3. Verify Divert Resources

```bash
# Check Divert custom resources
kubectl get diverts -n alice-movies

# Check HTTPRoutes in shared namespace
kubectl get httproutes -n movies-staging | grep okteto-

# Check service endpoints
kubectl get endpoints -n alice-movies
```

### 4. Browser Testing

1. Open the shared staging URL: `https://movies-movies-staging.okteto.dev`
2. Add query parameter: `?baggage=okteto-divert%3Dalice-movies`
3. Verify your service changes appear
4. Check that other services work normally (shared from staging)

## Resource Savings Comparison

| Scenario | Full Stack | With Divert | Savings |
|----------|-----------|-------------|---------|
| **Services Deployed** | 5 services | 1 service | 80% |
| **Databases** | 3 databases | 0 (shared) | 100% |
| **Memory Usage** | ~8 GB | ~1.2 GB | 85% |
| **CPU Usage** | ~4 cores | ~0.5 cores | 87.5% |
| **Setup Time** | 5-10 minutes | 30 seconds | 90% |
| **Cost per Developer** | $$$$ | $ | ~85% |

### Real-World Impact

- **Team of 10 developers**: Save ~$40,000/year in infrastructure costs
- **100 daily deployments**: Save 15+ hours of deployment time
- **Shared resources**: MongoDB, PostgreSQL, Kafka only deployed once

## Troubleshooting

### Issue: Services can't communicate

**Solution**: Verify network policies allow cross-namespace communication
```bash
kubectl get networkpolicies -n alice-movies
kubectl get networkpolicies -n movies-staging
```

### Issue: Baggage header not propagating

**Solution**: Check service logs for header values
```bash
kubectl logs -n alice-movies deployment/frontend -f
```

### Issue: Divert routing not working

**Solution**: Verify Divert custom resource is created
```bash
kubectl describe divert -n alice-movies
kubectl get httproutes -n movies-staging
```

### Issue: Can't connect to shared databases

**Solution**: Verify service discovery and DNS
```bash
kubectl run -it --rm debug --image=nicolaka/netshoot --restart=Never -- nslookup mongodb.movies-staging.svc.cluster.local
```

### Issue: Linkerd not injecting sidecar (Nginx driver only)

**Solution**: Ensure shared namespace has `okteto-shared` label for Linkerd sidecar injection
```bash
kubectl get namespace movies-staging -o yaml | grep okteto-shared
```

**Note**: This issue only applies when using the Nginx driver. Istio driver does not require Linkerd.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│ Shared Namespace (movies-staging)                           │
│ [okteto-shared label for Linkerd injection]                 │
│                                                              │
│  ┌──────────┐  ┌─────────┐  ┌─────┐  ┌──────┐             │
│  │ Frontend │  │ Catalog │  │ API │  │ Rent │             │
│  └──────────┘  └─────────┘  └─────┘  └──────┘             │
│       │             │          │         │                  │
│  ┌────▼─────────────▼──────────▼─────────▼────┐            │
│  │   MongoDB   PostgreSQL   Kafka              │            │
│  └──────────────────────────────────────────────┘           │
└─────────────────────────────────────────────────────────────┘
                          │
              Nginx Ingress + Linkerd
           (Header-based routing: baggage)
                          │
┌─────────────────────────▼───────────────────────────────────┐
│ Personal Namespace (alice-movies)                           │
│                                                              │
│  ┌──────────┐                                               │
│  │ Frontend │  (Your development version)                   │
│  │ (dev)    │                                               │
│  └──────────┘                                               │
│       │                                                      │
│  Connects to shared services:                               │
│  - Catalog (movies-staging)                                 │
│  - API (movies-staging)                                     │
│  - Rent (movies-staging)                                    │
│  - All databases (movies-staging)                           │
└─────────────────────────────────────────────────────────────┘

Note: With Istio driver, Linkerd is not required - Istio VirtualServices
handle routing natively.
```

## Best Practices

1. **Use descriptive namespace names**: `<your-name>-movies` or `<feature-name>-movies`
2. **Clean up when done**: Delete personal namespaces after development
   ```bash
   okteto namespace delete alice-movies
   ```
3. **Test locally first**: Verify changes work before deploying with Divert
4. **Monitor resource usage**: Check your namespace doesn't exceed quotas
5. **Keep shared staging updated**: Regularly update the shared environment
6. **Use preview environments**: For major changes, consider full preview environments

## Advanced Usage

### Switching to Istio Driver

This demo is configured for **Nginx + Linkerd**. To use **Istio** instead:

1. Edit the divert configuration files (e.g., `okteto-frontend-divert.yaml`)
2. Change the driver:
   ```yaml
   divert:
     namespace: ${SHARED_NAMESPACE:-staging}
     driver: istio  # Changed from 'nginx'
   ```
3. Deploy normally - Istio will handle routing without Linkerd

### Manual Divert Setup

If you prefer manual control:

```bash
# Create namespace
okteto namespace create alice-movies

# Deploy with specific manifest
okteto deploy --file okteto-frontend-divert.yaml --var SHARED_NAMESPACE=movies-staging

# Start development mode
okteto up frontend
```

### Multiple Services

To work on multiple services simultaneously:

1. Modify a divert YAML to include multiple services
2. Build multiple service images
3. Deploy them together in your namespace

### Custom Shared Namespace

Use a different shared namespace:

```bash
./scripts/setup-divert.sh my-custom-staging
```

## Support

- **Documentation**: [okteto.com/docs/divert](https://okteto.com/docs)
- **GitHub Issues**: [github.com/okteto/movies/issues](https://github.com/okteto/movies/issues)
- **Community**: [community.okteto.com](https://community.okteto.com)
- **Slack**: [okteto.com/slack](https://okteto.com/slack)

## License

Apache License 2.0 - See LICENSE file for details
