# 定义变量
BINARY_NAME=dist/kube-tide
BINARY_NAME_PROD=dist/kube-tide-prod
API_SERVER_BINARY=dist/api-server
MIGRATE_BINARY=dist/migrate
GO=go
GOFLAGS=-v
LDFLAGS=-w -s
# 定义源文件目录
CMD_DIR=./cmd/server
API_SERVER_CMD_DIR=./cmd/api-server
MIGRATE_CMD_DIR=./cmd/migrate
CONFIG_DIR=./configs
INTERNAL_DIR=./internal
PKG_DIR=./pkg
WEB_DIR=./web
DIST_DIR=./dist
WEB_DIST_DIR=$(PKG_DIR)/embed/web
SCRIPTS_DIR=./scripts
DATA_DIR=./data
# 定义Node.js工具
NODE=node
NPM=npm
PNPM=pnpm

# 数据库配置
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=
DB_NAME=kube_tide
SQLITE_FILE=./data/kube_tide.db

# 默认目标
.PHONY: all
all: build

# 重新构建 - 先清理再构建
.PHONY: rebuild
rebuild: clean build-prod
	@echo "重新构建完成"

# 构建前端（生产环境）
.PHONY: build-web-prod
build-web-prod:
	@echo "开始前端生产构建..."
	cd $(WEB_DIR) && $(PNPM) install && $(PNPM) build
	@echo "前端生产构建完成"

# 构建后端（生产环境）
.PHONY: build-backend-prod
build-backend-prod:
	@echo "开始后端生产构建..."
	mkdir -p $(DIST_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS) -X 'kube-tide/configs.BuildMode=production'" -o $(BINARY_NAME_PROD) $(CMD_DIR)/main.go
	@echo "后端生产构建完成"

# 构建生产版本
.PHONY: build-prod
build-prod:
	@echo "开始生产版本构建..."
	$(MAKE) build-web-prod
	@echo "前端生产构建完成"
	mkdir -p $(WEB_DIST_DIR)
	mv $(WEB_DIR)/dist $(WEB_DIST_DIR)
	$(MAKE) build-backend-prod
	@echo "生产版本构建完成"

# 仅构建后端
.PHONY: build-backend
build-backend:
	@echo "仅构建后端..."
	mkdir -p $(DIST_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) $(CMD_DIR)/main.go
	@echo "后端构建完成"

# 构建 API 服务器
.PHONY: build-api-server
build-api-server:
	@echo "构建 API 服务器..."
	mkdir -p $(DIST_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(API_SERVER_BINARY) $(API_SERVER_CMD_DIR)/main.go
	@echo "API 服务器构建完成"

# 构建数据库迁移工具
.PHONY: build-migrate
build-migrate:
	@echo "构建数据库迁移工具..."
	mkdir -p $(DIST_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(MIGRATE_BINARY) $(MIGRATE_CMD_DIR)/main.go
	@echo "数据库迁移工具构建完成"

# 构建所有二进制文件
.PHONY: build-all
build-all: build-backend build-api-server build-migrate
	@echo "所有二进制文件构建完成"

# 运行
.PHONY: run
run: 
	./$(BINARY_NAME)

# 运行生产版本
.PHONY: run-prod
run-prod: build-prod
	./$(BINARY_NAME_PROD)

# 开发模式运行（不嵌入前端资源）
.PHONY: dev
dev: build-backend
	./$(BINARY_NAME)

# 运行 API 服务器
.PHONY: run-api-server
run-api-server: build-api-server
	./$(API_SERVER_BINARY)

# 开发模式运行 API 服务器（使用 SQLite）
.PHONY: dev-api-server
dev-api-server: build-api-server
	mkdir -p $(DATA_DIR)
	DB_TYPE=sqlite DB_SQLITE_PATH=$(SQLITE_FILE) ./$(API_SERVER_BINARY)

# 运行数据库迁移
.PHONY: migrate
migrate: build-migrate
	@echo "运行数据库迁移..."
	./$(MIGRATE_BINARY) -action=migrate -type=$(DB_TYPE) -host=$(DB_HOST) -port=$(DB_PORT) -user=$(DB_USER) -password=$(DB_PASSWORD) -database=$(DB_NAME)

# 运行数据库迁移（SQLite）
.PHONY: migrate-sqlite
migrate-sqlite: build-migrate
	@echo "运行 SQLite 数据库迁移..."
	mkdir -p $(DATA_DIR)
	./$(MIGRATE_BINARY) -action=migrate -type=sqlite -sqlite-file=$(SQLITE_FILE)

# 查看当前迁移版本
.PHONY: migrate-version
migrate-version: build-migrate
	@echo "查看当前迁移版本..."
	./$(MIGRATE_BINARY) -action=version -type=$(DB_TYPE) -host=$(DB_HOST) -port=$(DB_PORT) -user=$(DB_USER) -password=$(DB_PASSWORD) -database=$(DB_NAME)

# 查看当前迁移版本（SQLite）
.PHONY: migrate-version-sqlite
migrate-version-sqlite: build-migrate
	@echo "查看当前 SQLite 迁移版本..."
	./$(MIGRATE_BINARY) -action=version -type=sqlite -sqlite-file=$(SQLITE_FILE)

# 回滚数据库到指定版本（需要设置 VERSION 变量）
.PHONY: migrate-rollback
migrate-rollback: build-migrate
	@if [ -z "$(VERSION)" ]; then echo "请设置 VERSION 变量，例如: make migrate-rollback VERSION=6"; exit 1; fi
	@echo "回滚数据库到版本 $(VERSION)..."
	./$(MIGRATE_BINARY) -action=rollback -version=$(VERSION) -type=$(DB_TYPE) -host=$(DB_HOST) -port=$(DB_PORT) -user=$(DB_USER) -password=$(DB_PASSWORD) -database=$(DB_NAME)

# 回滚 SQLite 数据库到指定版本
.PHONY: migrate-rollback-sqlite
migrate-rollback-sqlite: build-migrate
	@if [ -z "$(VERSION)" ]; then echo "请设置 VERSION 变量，例如: make migrate-rollback-sqlite VERSION=6"; exit 1; fi
	@echo "回滚 SQLite 数据库到版本 $(VERSION)..."
	./$(MIGRATE_BINARY) -action=rollback -version=$(VERSION) -type=sqlite -sqlite-file=$(SQLITE_FILE)

# 初始化数据库（迁移 + 创建默认管理员）
.PHONY: db-init
db-init: migrate
	@echo "数据库初始化完成"
	@echo "默认管理员账户："
	@echo "  用户名: admin"
	@echo "  密码: admin123"
	@echo "  邮箱: admin@kube-tide.com"

# 初始化 SQLite 数据库
.PHONY: db-init-sqlite
db-init-sqlite: migrate-sqlite
	@echo "SQLite 数据库初始化完成"
	@echo "默认管理员账户："
	@echo "  用户名: admin"
	@echo "  密码: admin123"
	@echo "  邮箱: admin@kube-tide.com"

# 测试
.PHONY: test
test:
	$(GO) test -v ./...

# 运行测试并生成覆盖率报告
.PHONY: test-coverage
test-coverage:
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

# 运行 API 测试脚本
.PHONY: test-api
test-api:
	@echo "运行 API 测试..."
	chmod +x $(SCRIPTS_DIR)/test-api.sh
	$(SCRIPTS_DIR)/test-api.sh

# 运行数据库集成测试
.PHONY: test-db
test-db:
	@echo "运行数据库集成测试..."
	chmod +x $(SCRIPTS_DIR)/test-database-integration.sh
	$(SCRIPTS_DIR)/test-database-integration.sh

# 运行基准测试
.PHONY: bench
bench:
	$(GO) test -bench=. -benchmem ./...

# 清理
.PHONY: clean
clean:
	$(GO) clean
	rm -f $(BINARY_NAME) $(BINARY_NAME_PROD) $(API_SERVER_BINARY) $(MIGRATE_BINARY)
	find . -name '*.test' -delete
	find . -name '*.out' -delete
	rm -f coverage.out coverage.html
	rm -rf $(WEB_DIR)/dist
	rm -rf $(WEB_DIST_DIR)
	rm -rf $(DIST_DIR)
	@echo "清理完成"

# 格式化代码
.PHONY: fmt
fmt:
	$(GO) fmt ./...
	@echo "代码格式化完成"

# 代码检查
.PHONY: lint
lint:
	$(GO) vet ./...
	@echo "代码检查完成"

# 安全检查
.PHONY: security
security:
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec 未安装，跳过安全检查"; \
		echo "安装命令: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# 依赖管理
.PHONY: deps
deps:
	$(GO) mod tidy
	$(GO) mod verify
	@echo "依赖管理完成"

# 更新依赖
.PHONY: deps-upgrade
deps-upgrade:
	$(GO) run $(SCRIPTS_DIR)/upgrade_deps.go
	$(GO) mod tidy
	@echo "依赖更新完成"

# 生成文档
.PHONY: doc
doc:
	$(GO) doc -all ./...

# 构建 Docker 镜像
.PHONY: docker-build
docker-build: build
	docker build -t kube-tide:latest .

# 构建生产 Docker 镜像
.PHONY: docker-build-prod
docker-build-prod: build-prod
	docker build -t kube-tide:prod .

# 构建 API 服务器 Docker 镜像
.PHONY: docker-build-api
docker-build-api: build-api-server
	docker build -f Dockerfile.api -t kube-tide-api:latest .

# 运行 Docker 容器
.PHONY: docker-run
docker-run:
	docker run -p 8080:8080 kube-tide:latest

# 运行 API 服务器 Docker 容器
.PHONY: docker-run-api
docker-run-api:
	docker run -p 8080:8080 -v $(PWD)/data:/app/data kube-tide-api:latest

# 快速开始（SQLite 版本）
.PHONY: quick-start
quick-start: db-init-sqlite run-api-server

# 开发环境设置
.PHONY: dev-setup
dev-setup: deps build-all db-init-sqlite
	@echo "开发环境设置完成"
	@echo "可用命令："
	@echo "  make run-api-server    # 运行 API 服务器"
	@echo "  make test-api          # 测试 API"
	@echo "  make migrate-version-sqlite  # 查看数据库版本"

# 生产环境部署准备
.PHONY: prod-setup
prod-setup: build-prod docker-build-prod
	@echo "生产环境部署文件已准备完成"

# 帮助信息
.PHONY: help
help:
	@echo "Kube-Tide 项目管理命令"
	@echo ""
	@echo "🚀 快速开始:"
	@echo "    quick-start         快速开始（SQLite 版本）"
	@echo "    dev-setup          开发环境设置"
	@echo "    prod-setup         生产环境部署准备"
	@echo ""
	@echo "🔨 构建命令:"
	@echo "    build-backend      仅构建后端"
	@echo "    build-api-server   构建 API 服务器"
	@echo "    build-migrate      构建数据库迁移工具"
	@echo "    build-all          构建所有二进制文件"
	@echo "    build-prod         构建生产版本（包括前后端）"
	@echo "    build-web-prod     仅构建前端生产版本"
	@echo ""
	@echo "🏃 运行命令:"
	@echo "    run                运行主服务器"
	@echo "    run-api-server     运行 API 服务器"
	@echo "    dev-api-server     开发模式运行 API 服务器（SQLite）"
	@echo "    run-prod           运行生产版本"
	@echo ""
	@echo "🗄️ 数据库命令:"
	@echo "    migrate            运行数据库迁移（PostgreSQL）"
	@echo "    migrate-sqlite     运行数据库迁移（SQLite）"
	@echo "    migrate-version    查看当前迁移版本"
	@echo "    migrate-rollback   回滚到指定版本（需要 VERSION=N）"
	@echo "    db-init            初始化数据库（PostgreSQL）"
	@echo "    db-init-sqlite     初始化数据库（SQLite）"
	@echo ""
	@echo "🧪 测试命令:"
	@echo "    test               运行所有测试"
	@echo "    test-coverage      运行测试并生成覆盖率报告"
	@echo "    test-api           运行 API 测试脚本"
	@echo "    test-db            运行数据库集成测试"
	@echo "    bench              运行基准测试"
	@echo ""
	@echo "🛠️ 开发工具:"
	@echo "    clean              清理构建产物"
	@echo "    fmt                格式化代码"
	@echo "    lint               代码检查"
	@echo "    security           安全检查"
	@echo "    deps               依赖管理"
	@echo "    deps-upgrade       更新依赖"
	@echo "    doc                生成文档"
	@echo ""
	@echo "🐳 Docker 命令:"
	@echo "    docker-build       构建 Docker 镜像"
	@echo "    docker-build-prod  构建生产 Docker 镜像"
	@echo "    docker-build-api   构建 API 服务器镜像"
	@echo "    docker-run         运行 Docker 容器"
	@echo "    docker-run-api     运行 API 服务器容器"
	@echo ""
	@echo "📝 示例用法:"
	@echo "    make quick-start                    # 快速开始开发"
	@echo "    make migrate-rollback VERSION=6     # 回滚到版本 6"
	@echo "    make test-coverage                  # 运行测试并查看覆盖率"
	@echo "    DB_PASSWORD=secret make migrate     # 使用密码运行迁移"
	@echo ""
	@echo "🔧 环境变量:"
	@echo "    DB_TYPE      数据库类型 (postgres/sqlite, 默认: postgres)"
	@echo "    DB_HOST      数据库主机 (默认: localhost)"
	@echo "    DB_PORT      数据库端口 (默认: 5432)"
	@echo "    DB_USER      数据库用户 (默认: postgres)"
	@echo "    DB_PASSWORD  数据库密码"
	@echo "    DB_NAME      数据库名称 (默认: kube_tide)"
	@echo "    SQLITE_FILE  SQLite 文件路径 (默认: ./data/kube_tide.db)"