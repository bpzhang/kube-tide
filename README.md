# K8s Platform Go

A Kubernetes multi-cluster management platform based on Go and React, providing an intuitive web interface to simplify Kubernetes resource management and operations.

## Features

### Cluster Management

- ✅ Multi-cluster support and management
- ✅ Cluster connection testing
- ✅ Cluster resource overview

### Node Management

- ✅ Node status monitoring and details viewing
- ✅ Node resource usage visualization
- ✅ Node drain operations
- ✅ Scheduling control (Cordon/Uncordon)
- ✅ Node taints management
- ✅ Node labels management
- ✅ Node pools creation and management

### Workload Management

- ✅ Pod viewing, details and deletion
- ✅ Real-time Pod logs viewing
- ✅ Pod terminal connection
- ✅ Deployment creation and management
- ✅ Deployment scaling and restart
- ✅ Service management

## Technology Stack

### Backend

- Go language
- Gin Web framework
- client-go Kubernetes client library

### Frontend

- React
- TypeScript
- Ant Design component library
- Vite build tool

## System Architecture

The platform adopts a front-end and back-end separation architecture:

- Backend provides RESTful APIs
- Frontend communicates with backend through APIs
- Backend interacts with Kubernetes clusters via client-go

## Directory Structure

```
kube-tide/
├── cmd/                    # Application entry points
│   └── server/             # Server entry
├── configs/                # Configuration files
├── dist/                   # Build output directory
├── docs/                   # Documentation
├── internal/               # Internal packages
│   ├── api/                # API handlers and routes
│   ├── core/               # Core business logic
│   │   └── k8s/            # Kubernetes resource management
│   └── utils/              # Utility functions
├── pkg/                    # Exportable packages
│   └── embed/              # Embedded resources
├── web/                    # Frontend code
│   ├── public/             # Static resources
│   └── src/                # Source code
│       ├── api/            # API client
│       ├── components/     # React components
│       ├── layouts/        # Page layouts
│       └── pages/          # Page components
└── Makefile                # Build scripts
```

## Installation and Usage

### Prerequisites

- Go 1.16 or higher
- Node.js 14 or higher
- Yarn package manager
- Accessible Kubernetes cluster

### Build and Run

1. Clone the repository

    ```bash
    git clone https://github.com/your-username/kube-tide.git
    cd kube-tide
    ```

2. Install dependencies and build

    ```bash
    # Build production version (frontend and backend)
    make build-prod
    
    # Or build separately
    make build-web      # Build frontend only
    make build-backend  # Build backend only
    ```

3. Run the application

    ```bash
    # Run production version
    make run-prod
    
    # Or run development version
    make dev
    ```

4. Access the web interface

    ```textplain
    http://localhost:8080
    ```

## Development Guide

### Development Environment Setup

1. Backend Development

    ```bash
    # Build backend only and start
    make build-backend
    make dev
    ```

2. Frontend Development

    ```bash
    cd web
    yarn install
    yarn dev
    ```

### Available Make Commands

- `make build` - Build the project (frontend and backend)
- `make build-prod` - Build production version
- `make build-web` - Build frontend only
- `make build-backend` - Build backend only
- `make run` - Run the application
- `make dev` - Run in development mode
- `make test` - Run tests
- `make clean` - Clean build artifacts

## Todo List

- Implement StatefulSet, DaemonSet management
- Add ConfigMap and Secret management
- Implement storage management (PV, PVC, StorageClass)
- Integrate monitoring system (Prometheus)
- Implement RBAC permission management
- Add CI/CD integration

## Contribution Guidelines

Pull Requests or Issues are welcome to improve the project. Please ensure the code follows the official Go language code specifications and includes appropriate tests.

## License

[MIT License](LICENSE)