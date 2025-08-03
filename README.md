# Kube Tide

![Kube Tide Logo](docs/images/logo.png)

A modern Kubernetes multi-cluster management platform based on Go and React, providing an intuitive web interface to simplify Kubernetes resource management and operations.

[中文文档](README.zh-CN.md) | [English](README.md)

## ✨ Key Features

### 🌐 Cluster Management

- ✅ Multi-cluster support and management
- ✅ Cluster connection testing
- ✅ Cluster resource overview
- ✅ Cluster health monitoring

### 🖥️ Node Management

- ✅ Node status monitoring and details viewing
- ✅ Node resource usage visualization
- ✅ Node drain operations
- ✅ Scheduling control (Cordon/Uncordon)
- ✅ Node taints management
- ✅ Node labels management
- ✅ Node pools creation and management

### 🚀 Workload Management

#### Pod Management

- ✅ Pod viewing, details and deletion
- ✅ Real-time Pod logs viewing
- ✅ Pod terminal connection
- ✅ Pod resource monitoring (CPU, Memory, Disk usage)
- ✅ Pod metrics historical data visualization
- ✅ Pod events viewing

#### Deployment Management

- ✅ Deployment creation and management
- ✅ Deployment scaling and restart
- ✅ Deployment details viewing
- ✅ Deployment update strategy configuration
- ✅ Deployment health check configuration
- ✅ Deployment resource limits configuration
- ✅ Deployment node affinity configuration

#### StatefulSet Management

- ✅ StatefulSet basic management
- ✅ StatefulSet scaling
- ✅ StatefulSet details viewing

#### Service Management

- ✅ Service creation and management
- ✅ Service details viewing
- ✅ Service endpoints monitoring

### 📊 Monitoring & Observability

- ✅ Real-time resource monitoring
- ✅ Metrics data visualization
- ✅ Cluster and node resource overview
- ✅ Pod performance metrics history

### 🌍 Internationalization

- ✅ Chinese and English multi-language support
- ✅ Dynamic language switching

## 🛠️ Technology Stack

### Backend

- **Go** - Main programming language
- **Gin** - Web framework
- **client-go** - Kubernetes client library
- **WebSocket** - Real-time communication
- **Logrus** - Logging

### Frontend

- **React 18** - Frontend framework
- **TypeScript** - Type safety
- **Ant Design** - UI component library
- **Vite** - Build tool
- **React Router** - Routing
- **Axios** - HTTP client
- **ECharts** - Data visualization

## 🏗️ System Architecture

The platform adopts a front-end and back-end separation architecture:

- **Frontend**: React SPA application, communicating with backend through RESTful APIs and WebSocket
- **Backend**: Go microservice, interacting with multiple Kubernetes clusters via client-go
- **Real-time Communication**: WebSocket support for real-time log viewing and terminal connections

### Architecture Features

- 🔄 **Multi-cluster support**
- 🚀 **High-performance caching**
- 🔐 **Secure authentication**
- 📊 **Real-time monitoring**
- 🌐 **Internationalization**

## 📁 Directory Structure

```plaintext
kube-tide/
├── cmd/                    # Application entry points
│   ├── kube-tide/          # CLI entry
│   └── server/             # Server entry
├── configs/                # Configuration files
├── docs/                   # Documentation
│   ├── architecture.md     # Architecture documentation
│   ├── code_arch.md        # Code architecture
│   └── images/             # Documentation images
├── internal/               # Internal packages
│   ├── api/                # API handlers and routes
│   │   └── middleware/     # HTTP middlewares
│   ├── core/               # Core business logic
│   │   └── k8s/            # Kubernetes resource management
│   └── utils/              # Utility functions
│       ├── i18n/           # Internationalization
│       └── logger/         # Logging utilities
├── pkg/                    # Exportable packages
│   └── embed/              # Embedded resources
├── web/                    # Frontend code
│   ├── public/             # Static resources
│   └── src/                # Source code
│       ├── api/            # API client
│       ├── components/     # React components
│       ├── i18n/           # Internationalization
│       ├── layouts/        # Page layouts
│       ├── pages/          # Page components
│       └── utils/          # Utility functions
└── Makefile                # Build scripts
```

## 🚀 Installation and Usage

### Prerequisites

- Go 1.19 or higher
- Node.js 16 or higher
- pnpm package manager
- Accessible Kubernetes cluster

### Quick Start

1. **Clone the repository**

   ```bash
   git clone https://github.com/bpzhang/kube-tide.git
   cd kube-tide
   ```

2. **Build and run**

   ```bash
   # Build production version (frontend and backend)
   make build-prod
   
   # Run the application
   make run-prod
   ```

3. **Access the web interface**

   ```
   http://localhost:8080
   ```

### Development Setup

1. **Backend Development**

   ```bash
   # Install Go dependencies
   go mod download
   
   # Run backend in development mode
   make dev
   ```

2. **Frontend Development**

   ```bash
   cd web
   pnpm install
   pnpm dev
   ```

### Available Make Commands

- `make build` - Build the project (frontend and backend)
- `make build-prod` - Build production version
- `make build-web` - Build frontend only
- `make build-backend` - Build backend only
- `make run` - Run the application
- `make run-prod` - Run production version
- `make dev` - Run in development mode
- `make test` - Run tests
- `make verify` - Run verification (Maven-style)
- `make clean` - Clean build artifacts

## 🔧 Configuration

The application can be configured through:

- **Environment variables**
- **Configuration file** (`configs/config.yaml`)
- **Command line flags**

Key configuration options:
- Server port and host
- Kubernetes cluster configurations
- Logging levels
- Frontend build settings

## 📚 Documentation

- [Architecture Documentation](docs/architecture.md)
- [Code Architecture](docs/code_arch.md)
- [TODO List](docs/TODO.md)

## 🗺️ Roadmap

### Upcoming Features

- ConfigMap and Secret management
- Storage management (PV, PVC, StorageClass)
- Monitoring system integration (Prometheus)
- RBAC permission management
- CI/CD integration
- Helm Chart support

See the complete [TODO list](docs/TODO.md) for detailed planning.

## 🤝 Contributing

We welcome contributions! Please feel free to submit Pull Requests or Issues to improve the project.

### Development Guidelines

- Follow Go official code standards
- Include appropriate tests for new features
- Update documentation when necessary
- Ensure TypeScript type safety for frontend code

## 📄 License

This project is licensed under the [MIT License](LICENSE).

## 🙏 Acknowledgments

- [Kubernetes](https://kubernetes.io/) - The amazing container orchestration platform
- [client-go](https://github.com/kubernetes/client-go) - Official Kubernetes Go client library
- [Ant Design](https://ant.design/) - Excellent React UI library
- [Gin](https://gin-gonic.com/) - High-performance Go web framework
