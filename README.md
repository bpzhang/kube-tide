# Kube Tide

<!-- ![Kube Tide Logo](docs/images/logo.png) -->

A Kubernetes multi-cluster management platform based on Go and React, providing an intuitive web interface to simplify Kubernetes resource management and operations.

## Features

### 🔐 User Role System (NEW!)

- ✅ **JWT Authentication**: Secure token-based authentication
- ✅ **User Management**: Complete CRUD operations for users
- ✅ **Role-Based Access Control (RBAC)**: Flexible role and permission system
- ✅ **Multi-Scope Permissions**: Global, cluster, and namespace-level permissions
- ✅ **Audit Logging**: Comprehensive operation tracking and security auditing
- ✅ **Session Management**: Secure user session handling
- ✅ **Password Security**: bcrypt encryption and secure password policies

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
- ✅ Pod resource monitoring (CPU, Memory, Disk usage)
- ✅ Pod metrics historical data visualization
- ✅ Deployment creation and management
- ✅ Deployment scaling and restart
- ✅ StatefulSet management
- ✅ Service management

### 📊 Database Integration

- ✅ **Dual-Track Architecture**: Kubernetes API + Database API
- ✅ **Data Persistence**: Historical data storage and analysis
- ✅ **Async Synchronization**: Non-blocking database operations
- ✅ **Multi-Database Support**: SQLite (development) and PostgreSQL (production)
- ✅ **Smart Updates**: Automatic data synchronization and conflict resolution

## Technology Stack

### Backend

- **Language**: Go 1.19+
- **Web Framework**: Gin
- **Database**: PostgreSQL 13+ / SQLite
- **Authentication**: JWT (JSON Web Tokens)
- **Password Encryption**: bcrypt
- **Logging**: Zap (structured logging)
- **Kubernetes Client**: client-go
- **UUID Generation**: Google UUID
- **Data Validation**: Validator v10

### Frontend

- **Framework**: React 18
- **Language**: TypeScript 4.9+
- **UI Library**: Ant Design 5+
- **State Management**: Redux Toolkit
- **HTTP Client**: Axios
- **Build Tool**: Vite 4+

## System Architecture

The platform adopts a modern layered architecture with comprehensive security:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │───▶│   API Gateway   │───▶│  Kubernetes     │
│   React UI      │    │   Gin Router    │    │   Clusters      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │  Auth Middleware│
                       │  Permission     │
                       │  Validation     │
                       └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │  Business Logic │
                       │  Service Layer  │
                       └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │  Data Access    │
                       │  Repository     │
                       └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │   Database      │
                       │  PostgreSQL     │
                       └─────────────────┘
```

## Directory Structure

```plaintext
kube-tide/
├── cmd/                    # Application entry points
│   ├── server/             # Main server
│   └── migrate/            # Database migration tool
├── configs/                # Configuration files
├── dist/                   # Build output directory
├── docs/                   # Documentation
│   ├── user-role-system.md # User role system documentation
│   ├── api-integration.md  # API integration guide
│   └── database.md         # Database documentation
├── internal/               # Internal packages
│   ├── api/                # API handlers and routes
│   │   ├── middleware/     # Authentication & authorization middleware
│   │   ├── auth_handler.go # Authentication endpoints
│   │   └── ...
│   ├── core/               # Core business logic
│   │   ├── auth_service.go # Authentication service
│   │   ├── user_service.go # User management service
│   │   ├── role_service.go # Role management service
│   │   └── k8s/            # Kubernetes resource management
│   ├── repository/         # Data access layer
│   │   ├── user_repository.go
│   │   ├── role_repository.go
│   │   └── ...
│   ├── database/           # Database models and migrations
│   │   ├── models/         # Data models
│   │   └── migrations/     # Database migration files
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

- Go 1.19 or higher
- PostgreSQL 13+ (or SQLite for development)
- Node.js 16 or higher
- Yarn package manager
- Accessible Kubernetes cluster

### Quick Start

1. **Clone the repository**

    ```bash
    git clone https://github.com/your-username/kube-tide.git
    cd kube-tide
    ```

2. **Setup Database**

    ```bash
    # For PostgreSQL (Production)
    createdb kube_tide
    export DB_HOST=localhost
    export DB_PORT=5432
    export DB_USER=postgres
    export DB_PASSWORD=your_password
    export DB_NAME=kube_tide
    
    # For SQLite (Development)
    export DB_TYPE=sqlite
    export DB_SQLITE_PATH=./data/kube_tide.db
    
    # Set JWT secret
    export JWT_SECRET=your_jwt_secret_key
    ```

3. **Run Database Migrations**

    ```bash
    go run cmd/migrate/main.go up
    ```

4. **Build and Run**

    ```bash
    # Build production version (frontend and backend)
    make build-prod
    
    # Or build separately
    make build-web      # Build frontend only
    make build-backend  # Build backend only
    ```

5. **Run the application**

    ```bash
    # Run production version
    make run-prod
    
    # Or run development version
    make dev
    ```

6. **Access the web interface**

    ```textplain
    http://localhost:8080
    ```

### Default Admin Account

After running migrations, a default admin account is created:

- **Username**: `admin`
- **Password**: `admin123`
- **Email**: `admin@kube-tide.local`

⚠️ **Important**: Change the default password immediately after first login!

## Development Guide

### Development Environment Setup

1. **Backend Development**

    ```bash
    # Install Go dependencies
    go mod tidy
    
    # Run database migrations
    go run cmd/migrate/main.go up
    
    # Start backend server
    go run cmd/server/main.go
    ```

2. **Frontend Development**

    ```bash
    cd web
    npm install
    npm run dev
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

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run integration tests
go test -tags=integration ./...

# Test API endpoints
chmod +x scripts/test-api.sh
./scripts/test-api.sh
```

## API Documentation

### Authentication Endpoints

```
POST   /api/auth/login           # User login
POST   /api/auth/logout          # User logout
POST   /api/auth/register        # User registration
POST   /api/auth/change-password # Change password
GET    /api/auth/me              # Get current user info
```

### User Management Endpoints

```
GET    /api/users                # List users
POST   /api/users                # Create user
GET    /api/users/:id             # Get user details
PUT    /api/users/:id             # Update user
DELETE /api/users/:id             # Delete user
GET    /api/users/:id/roles       # Get user roles
POST   /api/users/:id/roles       # Assign role
DELETE /api/users/:id/roles/:role # Remove role
```

### Role & Permission Endpoints

```
GET    /api/roles                     # List roles
POST   /api/roles                     # Create role
GET    /api/roles/:id                 # Get role details
PUT    /api/roles/:id                 # Update role
DELETE /api/roles/:id                 # Delete role
GET    /api/roles/:id/permissions     # Get role permissions
POST   /api/roles/:id/permissions     # Assign permissions
DELETE /api/roles/:id/permissions     # Remove permissions

GET    /api/permissions               # List permissions
POST   /api/permissions/check         # Check permission
```

For detailed API documentation, see [docs/user-role-system.md](docs/user-role-system.md).

## Security Features

### Authentication & Authorization

- **JWT Token Authentication**: Secure, stateless authentication
- **Role-Based Access Control**: Flexible permission system
- **Multi-Scope Permissions**: Global, cluster, and namespace levels
- **Session Management**: Secure session handling with expiration
- **Password Security**: bcrypt encryption with configurable complexity

### Permission Scopes

1. **Global**: Access to all clusters and namespaces
2. **Cluster**: Access to specific cluster resources
3. **Namespace**: Access to specific namespace resources

### Predefined Roles

- **System Admin**: Full system access (`*:*` globally)
- **Cluster Admin**: Full cluster access (`*:*` cluster-scoped)
- **Developer**: Deployment and service management (namespace-scoped)
- **Viewer**: Read-only access (`*:read`)

## Configuration

### Environment Variables

```bash
# Database Configuration
DB_TYPE=postgres                    # Database type: postgres or sqlite
DB_HOST=localhost                   # Database host
DB_PORT=5432                       # Database port
DB_USER=postgres                   # Database user
DB_PASSWORD=your_password          # Database password
DB_NAME=kube_tide                  # Database name
DB_SSL_MODE=disable                # SSL mode for PostgreSQL

# For SQLite
DB_SQLITE_PATH=./data/kube_tide.db # SQLite database file path

# Authentication
JWT_SECRET=your_jwt_secret_key     # JWT signing secret
JWT_EXPIRES_HOURS=24               # JWT token expiration (hours)

# Server Configuration
PORT=8080                          # Server port
LOG_LEVEL=info                     # Log level: debug, info, warn, error
ENABLE_CORS=true                   # Enable CORS for development

# Database Connection Pool
DB_MAX_OPEN_CONNS=25               # Maximum open connections
DB_MAX_IDLE_CONNS=5                # Maximum idle connections
DB_CONN_MAX_LIFETIME=5m            # Connection maximum lifetime
```

## Docker Deployment

### Using Docker Compose

```yaml
version: '3.8'
services:
  postgres:
    image: postgres:13
    environment:
      POSTGRES_DB: kube_tide
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  kube-tide:
    build: .
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: postgres
      DB_PASSWORD: password
      DB_NAME: kube_tide
      JWT_SECRET: your_jwt_secret
    ports:
      - "8080:8080"
    depends_on:
      - postgres

volumes:
  postgres_data:
```

Start with: `docker-compose up -d`

## Documentation

- 📚 [User Role System Guide](docs/user-role-system.md) - Comprehensive user management documentation
- 🔗 [API Integration Guide](docs/api-integration.md) - API usage and integration examples
- 🗄️ [Database Documentation](docs/database.md) - Database schema and migration guide
- 🏗️ [Architecture Overview](docs/architecture.md) - System architecture and design patterns

## Todo List

### Core Features

- [x] ✅ **User Role System** - Complete RBAC implementation
- [x] ✅ **Database Integration** - Data persistence and API layer
- [ ] Implement ConfigMap and Secret management
- [ ] Add storage management (PV, PVC, StorageClass)
- [ ] Integrate monitoring system (Prometheus)
- [ ] Add CI/CD integration

### Security Enhancements

- [x] ✅ **JWT Authentication** - Token-based authentication
- [x] ✅ **RBAC Permissions** - Role-based access control
- [x] ✅ **Audit Logging** - Operation tracking and security auditing
- [ ] Multi-factor authentication (MFA)
- [ ] OAuth2/OIDC integration
- [ ] API rate limiting

### UI/UX Improvements

- [ ] User management interface
- [ ] Role and permission configuration UI
- [ ] Audit log viewer
- [ ] Dashboard with security metrics

## Contributing

Pull Requests or Issues are welcome to improve the project. Please ensure the code follows the official Go language code specifications and includes appropriate tests.

### Development Guidelines

1. Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
2. Write unit tests with >80% coverage
3. Use conventional commit messages
4. Update documentation for new features
5. Ensure all tests pass before submitting PR

## License

[MIT License](LICENSE)

---

## 📞 Support

- 📧 **Email**: <support@kube-tide.com>
- 🐛 **Issues**: [GitHub Issues](https://github.com/your-org/kube-tide/issues)
- 📖 **Documentation**: [Project Wiki](https://github.com/your-org/kube-tide/wiki)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/your-org/kube-tide/discussions)

---

*Last Updated: 2024-01-01*
*Version: v1.0.0 with User Role System*
