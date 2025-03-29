# 定义变量
BINARY_NAME=dist/kube-tide
BINARY_NAME_PROD=dist/kube-tide-prod
GO=go
GOFLAGS=-v
LDFLAGS=-w -s
# 定义源文件目录
CMD_DIR=./cmd/server
CONFIG_DIR=./configs
INTERNAL_DIR=./internal
PKG_DIR=./pkg
WEB_DIR=./web
DIST_DIR=./dist
WEB_DIST_DIR=$(PKG_DIR)/embed/web
# 定义Node.js工具
NODE=node
NPM=npm
YARN=yarn

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
	cd $(WEB_DIR) && $(YARN) install && npx vite build
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

# 运行
.PHONY: run
run: 
	./$(BINARY_NAME)

# 运行生产版本
.PHONY: run-prod
run-prod: build-prod
	./$(BINARY_NAME)

# 开发模式运行（不嵌入前端资源）
.PHONY: dev
dev: build-backend
	./$(BINARY_NAME)

# 测试
.PHONY: test
test:
	$(GO) test -v ./...

# 清理
.PHONY: clean
clean:
	$(GO) clean
	rm -f $(BINARY_NAME)
	find . -name '*.test' -delete
	find . -name '*.out' -delete
	rm -rf $(WEB_DIR)/dist
	rm -rf $(WEB_DIST_DIR)
	rm -rf $(DIST_DIR)

# 格式化代码
.PHONY: fmt
fmt:
	$(GO) fmt ./...

# 代码检查
.PHONY: lint
lint:
	$(GO) vet ./...

# 依赖管理
.PHONY: deps
deps:
	$(GO) mod tidy
	$(GO) mod verify

# 生成文档
.PHONY: doc
doc:
	$(GO) doc -all ./...

# 构建 Docker 镜像
.PHONY: docker-build
docker-build: build
	docker build -t $(BINARY_NAME):latest .

# 构建生产 Docker 镜像
.PHONY: docker-build-prod
docker-build-prod: build-prod
	docker build -t $(BINARY_NAME_PROD):prod .

# 帮助信息
.PHONY: help
help:
	@echo "管理 $(BINARY_NAME) 的可用命令:"
	@echo ""
	@echo "使用方法:"
	@echo "    make [命令]"
	@echo ""
	@echo "命令:"
	@echo "    build           构建项目（包括前后端）"
	@echo "    build-prod      构建生产版本（优化的前后端）"
	@echo "    build-web       仅构建前端"
	@echo "    build-web-prod  仅构建前端生产版本"
	@echo "    build-backend   仅构建后端"
	@echo "    run             运行项目"
	@echo "    run-prod        运行生产版本"
	@echo "    dev             开发模式运行（不嵌入前端资源）"
	@echo "    test            运行测试"
	@echo "    clean           清理构建产物"
	@echo "    fmt             格式化代码"
	@echo "    lint            运行代码检查"
	@echo "    deps            更新依赖"
	@echo "    doc             生成文档"
	@echo "    docker-build    构建 Docker 镜像"
	@echo "    docker-build-prod 构建生产 Docker 镜像"
	@echo "    help            显示帮助信息"