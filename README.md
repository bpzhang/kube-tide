# Kube Tide

A modern Kubernetes multi-cluster management platform based on Go and React, providing an intuitive web interface to simplify Kubernetes resource management and operations.

[中文文档](README.zh-CN.md) | [English](README.md)

## Key Features

### Cluster Management

- Multi-cluster support and management
- Cluster connection testing
- Cluster resource overview

### Node Management

- Node status monitoring and details viewing
- Node resource usage visualization
- Node drain operations
- Scheduling control (Cordon/Uncordon)
- Node taints and labels management
- Node pools creation and management
- Cluster autoscaler configuration

### Workload Management

#### Pod Management

- Pod viewing, details and deletion
- Real-time Pod logs viewing
- Pod terminal connection
- Pod resource monitoring (CPU, Memory, Disk usage)
- Pod metrics historical data visualization
- Pod events viewing
- Pod lifecycle and restart policy management

#### Deployment Management

- Deployment creation and management
- Deployment scaling, restart, rollout history and rollback
- Deployment details with tabbed layout (overview, containers, status, pods, access, events)
- Related Services, Endpoints and Ingress routes in the access tab
- Update strategy, health checks, resource limits and node affinity configuration

#### StatefulSet Management

- StatefulSet basic management, scaling and details

#### Service & Ingress

- Service creation and management
- Service details and endpoints monitoring
- Ingress listing by namespace (shown in Deployment access tab)

### Monitoring & Observability

- Real-time resource monitoring
- Metrics data visualization (Recharts)
- Cluster and node resource overview
- Pod performance metrics history (in-memory cache)

### Internationalization

- Chinese and English multi-language support
- Dynamic language switching

## Technology Stack

### Backend

- **Go 1.26+** — Main language
- **Gin** — Web framework
- **client-go v0.36** — Kubernetes client
- **coder/websocket** — Pod terminal
- **zap + lumberjack** — Logging with rotation
- **viper** — Configuration

### Frontend

- **React 19** — UI framework
- **TypeScript** — Type safety
- **Vite 8** — Build tool
- **Ant Design 6** — UI components
- **React Router 7** — Routing
- **Axios** — HTTP client
- **Recharts** — Charts
- **xterm.js** — Terminal

## System Architecture

Monolithic Go HTTP server with embedded frontend (production build). The backend talks to multiple Kubernetes clusters via client-go; the React SPA uses REST and WebSocket.

See [Architecture Documentation](docs/architecture.md) for details.

**Important for production (ECS / VM):**

- Deploy on a **standalone ECS or VM**, not inside Kubernetes
- No built-in platform authentication — protect with reverse proxy / network isolation
- Cluster registrations are in-memory only — re-register after restart
- Single instance only

See [Operations Guide](docs/operations.md) §2 for ECS deployment.

## Directory Structure

```plaintext
kube-tide/
├── cmd/server/           # Application entry
├── configs/              # Configuration
├── docs/                 # Documentation
├── internal/
│   ├── api/              # HTTP handlers & routes
│   ├── core/k8s/         # Kubernetes business logic
│   └── utils/            # Logger, i18n
├── pkg/embed/            # Embedded frontend (production)
├── web/                  # React frontend
├── scripts/              # Utility scripts
├── Makefile
└── README.md
```

Full tree: [Code Architecture](docs/code_arch.md)

## Installation and Usage

### Prerequisites

- Go 1.26+
- Node.js 18+ (LTS recommended)
- pnpm
- Accessible Kubernetes cluster(s)

### Quick Start (Production)

```bash
git clone https://github.com/bpzhang/kube-tide.git
cd kube-tide

make build-prod
make run-prod
```

Open: `http://localhost:8080`

### Development

**Option A — Makefile (frontend hot reload + backend)**

```bash
make run
# Frontend: http://127.0.0.1:5173
# Backend:  http://127.0.0.1:8080
```

**Option B — Separate terminals**

```bash
# Terminal 1
cd web && pnpm install && pnpm dev

# Terminal 2
go mod download
go run ./cmd/server/main.go
```

### Make Commands

| Command | Description |
|---------|-------------|
| `make build` | Build dev binary + embed frontend |
| `make build-prod` | Build production binary |
| `make run` | Dev mode (Vite + backend) |
| `make run-prod` | Build and run production |
| `make clean` | Remove build artifacts |
| `make docker-build` | Build Docker image (optional on ECS) |
| `make docker-run` | Run container locally |
| `make help` | Show available targets |

## Configuration

- **File:** `configs/config.yaml` — server port, logging
- **Environment:** `K8S_PLATFORM_ENV=production` — force production mode
- **Clusters:** added at runtime via UI or API (not in config file)

See [Operations Guide](docs/operations.md) for production settings.

## Documentation

- [Architecture](docs/architecture.md)
- [Code Structure](docs/code_arch.md)
- [Operations & Deployment](docs/operations.md)
- [Roadmap / TODO](docs/TODO.md)

## Roadmap

- ConfigMap and Secret management
- Storage (PV, PVC, StorageClass)
- Prometheus integration
- Platform RBAC / authentication

Deployment: **ECS / VM** with systemd — see [Operations Guide](docs/operations.md). Optional Docker: `deployments/docker/`.

See [TODO](docs/TODO.md) for the full list.

## Contributing

Pull Requests and Issues are welcome.

- Follow Go official conventions
- Keep frontend TypeScript types accurate
- Update docs when changing behavior

## Dependency Upgrades

```bash
# Backend
./scripts/upgrade-deps.sh

# Or manually
go get -u ./...
go mod tidy

# Frontend
cd web && pnpm update
```

## License

[MIT License](LICENSE)

## Acknowledgments

- [Kubernetes](https://kubernetes.io/)
- [client-go](https://github.com/kubernetes/client-go)
- [Ant Design](https://ant.design/)
- [Gin](https://gin-gonic.com/)
