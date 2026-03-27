# 定义变量
BINARY_NAME=dist/kube-tide
BINARY_NAME_PROD=dist/kube-tide-prod
GO=go
GOFLAGS=-v
LDFLAGS=-w -s
CMD_DIR=./cmd/server
PKG_DIR=./pkg
WEB_DIR=./web
DIST_DIR=./dist
WEB_DIST_DIR=$(PKG_DIR)/embed/web
PNPM=pnpm

# 默认目标（测试环境构建）
.PHONY: all
all: build

# 内部步骤：构建前端并复制到 embed 目录
.PHONY: _prepare-web
_prepare-web:
	@echo "开始前端构建..."
	cd $(WEB_DIR) && $(PNPM) install && $(PNPM) build
	mkdir -p $(WEB_DIST_DIR)
	rm -rf $(WEB_DIST_DIR)/dist
	cp -R $(WEB_DIR)/dist $(WEB_DIST_DIR)
	@echo "前端构建完成"

# 测试环境：构建
.PHONY: build
build: _prepare-web
	@echo "开始测试环境后端构建..."
	mkdir -p $(DIST_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) $(CMD_DIR)/main.go
	@echo "测试环境构建完成"

# 测试环境：运行
.PHONY: run
run: build
	@echo "启动测试环境（前端热更新）..."
	@echo "前端地址: http://127.0.0.1:5173"
	@echo "后端地址: http://127.0.0.1:8080"
	@set -e; \
	cd $(WEB_DIR); $(PNPM) install >/dev/null; $(PNPM) dev --host 0.0.0.0 --port 5173 & \
	FRONT_PID=$$!; \
	cd ..; \
	trap 'kill $$FRONT_PID >/dev/null 2>&1 || true' EXIT INT TERM; \
	K8S_PLATFORM_ENV=production ./$(BINARY_NAME)

# 生产环境：构建
.PHONY: build-prod
build-prod: _prepare-web
	@echo "开始生产环境后端构建..."
	mkdir -p $(DIST_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS) -X 'kube-tide/configs.BuildMode=production'" -o $(BINARY_NAME_PROD) $(CMD_DIR)/main.go
	@echo "生产环境构建完成"

# 生产环境：运行
.PHONY: run-prod
run-prod: build-prod
	./$(BINARY_NAME_PROD)

# 清理构建产物
.PHONY: clean
clean:
	$(GO) clean
	rm -rf $(DIST_DIR)
	rm -rf $(WEB_DIR)/dist
	rm -rf $(WEB_DIST_DIR)
	find . -name '*.test' -delete
	find . -name '*.out' -delete

# 帮助信息
.PHONY: help
help:
	@echo "可用命令:"
	@echo "  make build      测试环境构建"
	@echo "  make run        测试环境运行（前端热更新，访问5173）"
	@echo "  make build-prod 生产环境构建"
	@echo "  make run-prod   生产环境运行（自动构建）"
	@echo "  make clean      清理构建产物"