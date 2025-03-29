# Kubernetes客户端

## 功能清单

- [TODO](./docs/TODO.md)

## 技术栈

### 后端技术栈

- **Web框架**: Gin (最流行的Go Web框架，性能优秀，社区活跃)
- **Kubernetes客户端**: client-go (Kubernetes官方Go客户端)
- **数据库**: PostgreSQL (开源、功能丰富、性能稳定)
- **缓存**: Redis (高性能、广泛使用的开源缓存)
- **认证**: JWT + OIDC (轻量级认证方案)
- **日志**: zap (Uber开发的高性能日志库)
- **配置管理**: viper (灵活的配置解决方案)
- **WebSocket**: gorilla/websocket (最流行的Go WebSocket库)
- **依赖注入**: wire (Google开发的依赖注入工具)
- **ORM**: GORM (Go语言最流行的ORM库)
- **测试**: testify (流行的Go测试框架)

### 前端技术栈

- **框架**: React 18 (最流行的前端框架)
- **类型系统**: TypeScript (增强代码可维护性)
- **状态管理**: Redux Toolkit (Redux官方推荐工具)
- **UI库**: Ant Design (成熟的企业级UI组件库)
- **图表库**: ECharts (功能丰富的开源图表库)
- **HTTP客户端**: Axios (可靠的HTTP客户端)

## 详细目录结构

[代码目录结构](./code_arch.md)

## 开发

### 克隆代码

```shell
git clone git@github.com:bpzhang/kube-tide.git
```

### 下载依赖

1. **go**

```shell
go mod tidy
```

1. **react**

```shell
npm install
```

```shell
npm run dev
```

### 构建

1. **使用Makefile构建项目**

   - 清理构建

   ```shell
   make clean
   ```

   - 构建

   ```shell
   make build
   ```

   - 重新构建

   ```shell
   make rebuild
   ```
