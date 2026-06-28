# Kube Tide 运维部署手册

本文档面向 **ECS / 云主机（VM）** 上的生产部署。Kube Tide 是 **管理 Kubernetes 的运维平台**，推荐单独跑在一台 ECS 上，**不要**把平台本身部署进 K8s 集群。

部署前请先阅读 [architecture.md](./architecture.md) 中的**当前限制**。

## 推荐架构（ECS）

```text
运维人员浏览器
      │ HTTPS (443)
      ▼
┌─────────────────────────────────────┐
│  ECS（一台）                          │
│  Nginx :443  →  kube-tide :8080      │
│  systemd 托管 kube-tide-prod 单进程   │
└──────────────┬──────────────────────┘
               │ client-go / HTTPS
               ▼
    待管理的 K8s 集群 A、B、C …
```

| 组件 | 部署位置 | 说明 |
|------|----------|------|
| kube-tide | ECS 本机 `:8080` | 建议只监听 `127.0.0.1`，由 Nginx 对外 |
| Nginx | 同一台 ECS `:443` | TLS + 访问控制 + WebSocket 代理 |
| 安全组 | 云控制台 | 开放 **443**（办公网/VPN）；**不要**对公网开放 8080 |

> **不需要** `deployments/kubernetes/` 那类清单——那是「平台跑在 K8s 里」的场景，与本项目定位不符。

## 1. 部署前检查清单

| 检查项 | 要求 | 说明 |
|--------|------|------|
| 网络隔离 | 必须 | 平台 API **无内置认证**，禁止直接暴露公网 |
| TLS | 强烈建议 | 同一台 ECS 上用 **Nginx** 终止 HTTPS |
| kubeconfig 权限 | 最小权限 | 平台 ServiceAccount 或 kubeconfig 应遵循 least privilege |
| 单实例 | 当前必须 | 集群注册信息存内存，**不支持多副本** |
| 持久化预期 | 了解限制 | 重启后需重新添加集群；无数据库备份 |
| 资源 | 按规模估算 | 见下文「资源规划」 |

## 2. ECS 部署（推荐）

### 2.1 准备 ECS

- 系统：Alibaba Cloud Linux 3 / Ubuntu 22.04 等常见 Linux
- 规格参考：2 vCPU / 2GiB 起（管理 3–5 个中小集群通常够用）
- 安全组：入站仅 **443**（及 SSH 22 限源 IP）；8080 不对公网开放
- ECS 需能 **访问** 各目标 K8s API（公网 Endpoint、专线或 VPC 内网）

### 2.2 在构建机编译

在开发机或 CI 上（需 Node + pnpm + Go 1.26+）：

```bash
git clone https://github.com/bpzhang/kube-tide.git
cd kube-tide
make build-prod
# 产物：dist/kube-tide-prod
```

也可在 ECS 上直接 clone 后 `make build-prod`（需在 ECS 安装 Go、Node、pnpm）。

### 2.3 安装到 ECS

```bash
# 在 ECS 上
sudo useradd -r -s /sbin/nologin kube-tide || true
sudo mkdir -p /opt/kube-tide/{configs,logs}
sudo chown -R kube-tide:kube-tide /opt/kube-tide

# 从构建机上传（示例）
scp dist/kube-tide-prod user@<ecs-ip>:/tmp/
scp deployments/ecs/config.production.yaml user@<ecs-ip>:/tmp/

# 在 ECS 上
sudo mv /tmp/kube-tide-prod /opt/kube-tide/
sudo mv /tmp/config.production.yaml /opt/kube-tide/configs/config.yaml
sudo chmod +x /opt/kube-tide/kube-tide-prod
sudo chown -R kube-tide:kube-tide /opt/kube-tide
```

### 2.4 systemd

```bash
sudo cp deployments/ecs/kube-tide.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now kube-tide
sudo systemctl status kube-tide
```

单元文件见 [deployments/ecs/kube-tide.service](../deployments/ecs/kube-tide.service)。

### 2.5 Nginx（同机）

```bash
sudo cp deployments/ecs/nginx.conf.example /etc/nginx/conf.d/kube-tide.conf
# 修改 server_name、证书路径，并配置访问控制
sudo nginx -t && sudo systemctl reload nginx
```

### 2.6 首次使用

1. 浏览器访问 `https://kube-tide.example.com`
2. 在「集群管理」中添加各 K8s 集群 kubeconfig
3. **进程重启后集群会丢失**，需重新添加（见 §10）

---

## 3. 构建与发布（通用）

### 3.1 从源码构建

```bash
git clone https://github.com/bpzhang/kube-tide.git
cd kube-tide

# 需要 Node.js + pnpm + Go 1.26+
make build-prod

# 产物
ls -la dist/kube-tide-prod
```

`build-prod` 会：安装前端依赖 → Vite 构建 → 复制到 `pkg/embed/web/dist` → 编译带 embed 的二进制。

### 3.2 运行

```bash
# 方式 A：直接运行
./dist/kube-tide-prod

# 方式 B：Makefile
make run-prod

# 方式 C：指定生产环境变量（非 prod 构建也可强制生产模式）
K8S_PLATFORM_ENV=production ./dist/kube-tide
```

默认监听 `configs/config.yaml` 中的 `server.port`（默认 `8080`）。

### 3.3 systemd

生产环境直接使用 [deployments/ecs/kube-tide.service](../deployments/ecs/kube-tide.service)，完整步骤见上文 **§2 ECS 部署**。

简要示例：


```ini
[Unit]
Description=Kube Tide
After=network.target

[Service]
Type=simple
User=kube-tide
WorkingDirectory=/opt/kube-tide
Environment=K8S_PLATFORM_ENV=production
ExecStart=/opt/kube-tide/kube-tide-prod
Restart=on-failure
RestartSec=5

# 优雅关闭：主进程处理 SIGTERM，预留足够时间
TimeoutStopSec=15
KillSignal=SIGTERM

[Install]
WantedBy=multi-user.target
```

建议 WorkingDirectory 为 `/opt/kube-tide`，保留 `configs/config.yaml` 与可写 `logs/`。

## 4. 配置

### 4.1 配置文件

路径：`configs/config.yaml`（也可放在工作目录根下的 `config.yaml`）

ECS 生产模板：[deployments/ecs/config.production.yaml](../deployments/ecs/config.production.yaml)（`host: 127.0.0.1`，仅本机 + Nginx 暴露）。

```yaml
server:
  port: 8080
  host: 127.0.0.1         # ECS 推荐：不对公网监听，由 Nginx 反代

logging:
  level: info            # debug | info | warn | error
  file:
    enabled: true
    path: "./logs/kube-tide.log"
    error_path: "./logs/kube-tide-error.log"
  rotate:
    enabled: true
    max_size: 100        # MB
    max_age: 30          # 天
    max_backups: 10
    compression: "after_days:7"
    local_time: true
    rotation_time: daily
```

字段说明见 `configs/config.go`。若文件缺失，viper 会使用内置默认值并打印 Warning。

### 4.2 环境变量

| 变量 | 作用 |
|------|------|
| `K8S_PLATFORM_ENV=production` | 启用生产模式：embed 静态资源、关闭 dev 静态路径 |

### 4.3 集群配置

- 集群 **不** 通过配置文件注册，而是通过 Web UI「集群管理」或 `POST /api/clusters` 动态添加
- 支持 kubeconfig **文件路径**或**内容**两种方式
- 内容方式会写入系统临时目录（如 `/tmp/kubeconfig-<name>.yaml`），权限 `0600`
- **进程重启后所有已注册集群丢失**，运维需：
  - 运维手册记录各集群 kubeconfig 的安全存储位置，或
  - 等待后续持久化功能（见 TODO）

## 5. 反向代理与 TLS

### 5.1 Nginx 示例

完整示例见 [deployments/ecs/nginx.conf.example](../deployments/ecs/nginx.conf.example)。

```nginx
upstream kube_tide {
    server 127.0.0.1:8080;
    keepalive 32;
}

server {
    listen 443 ssl http2;
    server_name kube-tide.example.com;

    ssl_certificate     /etc/ssl/certs/kube-tide.crt;
    ssl_certificate_key /etc/ssl/private/kube-tide.key;

    # 建议：在此层做 Basic Auth / OAuth2 / IP 白名单
    # auth_basic "Restricted";
    # auth_basic_user_file /etc/nginx/.htpasswd;

    location / {
        proxy_pass http://kube_tide;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Pod Exec WebSocket
    location ~ ^/api/clusters/.+/exec$ {
        proxy_pass http://kube_tide;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }
}
```

### 5.2 安全建议

1. **禁止**将 8080 端口无防护暴露到公网
2. 在反向代理层启用认证（OAuth2 Proxy、Authelia、VPN 等）
3. 限制源 IP（办公网 / 堡垒机）
4. 审计：记录谁访问了平台（代理 access log）
5. kubeconfig 使用专用 SA，避免 cluster-admin

## 6. 健康检查

当前实现（`internal/api/health_handler.go`）：

```http
GET /api/health
→ 200 { "status": "system.healthCheck" }
```

**局限**：仅表示 HTTP 进程存活，**不检查**：

- 已注册集群是否可达
- 指标采集 goroutine 是否正常
- 磁盘 / 日志目录是否可写

### 6.1 ECS 探测

ECS 上可用云监控或 cron 探测 `http://127.0.0.1:8080/api/health`。

> 未来应增加 `/api/health/ready`，探测至少一个集群连通性。

### 6.2 建议的外部监控

| 监控项 | 方式 |
|--------|------|
| 进程存活 | HTTP GET `/api/health` |
| 日志错误率 | 采集 `logs/kube-tide-error.log` |
| 磁盘 | `logs/` 分区使用率 |
| K8s 操作失败 | 应用日志关键字 `error`、`connection test failed` |
| WebSocket | 合成探测 Pod Exec 端点（可选） |

Prometheus 指标端点**尚未内置**（见 TODO）。

## 7. 日志

| 输出 | 路径 / 说明 |
|------|-------------|
| 标准输出 | 开发模式主要输出 |
| 文件 | `logging.file.path`（默认 `./logs/kube-tide.log`） |
| 错误文件 | `logging.file.error_path` |
| 轮转 | lumberjack：按 `max_size`、`max_age`、`max_backups` |

运维建议：

- 生产开启 `file.enabled` 与 `rotate.enabled`
- 使用 logrotate 或日志采集 Agent（Fluent Bit / Vector / Promtail）采集 `logs/`
- 开发排查时可临时设 `logging.level: debug`

## 8. 优雅关闭

收到 `SIGINT` / `SIGTERM` 后：

1. 取消 Pod 指标采集 context
2. `http.Server.Shutdown`（5 秒超时）
3. 退出进程

systemd 已配置 `TimeoutStopSec=15`；升级时用 `systemctl stop kube-tide` 即可触发优雅关闭。

## 9. 资源规划（参考）

| 规模 | CPU | 内存 | 说明 |
|------|-----|------|------|
| 小型（1–3 集群，<500 Pod） | 0.5–1 核 | 512Mi–1Gi | 默认指标采集间隔 1 分钟 |
| 中型（3–10 集群） | 1–2 核 | 1–2Gi | 关注指标内存缓存 |
| 大型 | 需压测 | 需压测 | 当前架构可能需优化采集频率 |

WebSocket 终端连接会占用长连接与 PTY 资源，并发 Exec 较多时需适当扩容。

## 10. Docker（可选，非必需）

若不想在 ECS 上安装 Go/Node，可在 ECS 上 **只装 Docker**，用镜像跑单容器。平台仍是对外管理 K8s 的运维工具，**不是**跑在 K8s 里。

| 路径 | 说明 |
|------|------|
| `deployments/docker/Dockerfile` | 多阶段构建（Node + Go + distroless） |

```bash
# 构建机
make docker-build DOCKER_IMAGE=your-registry/kube-tide:1.0.0
docker push your-registry/kube-tide:1.0.0

# ECS 上（示例）
docker run -d --name kube-tide --restart unless-stopped \
  -p 127.0.0.1:8080:8080 \
  -v /opt/kube-tide/configs:/app/configs:ro \
  -v /opt/kube-tide/logs:/app/logs \
  -e K8S_PLATFORM_ENV=production \
  your-registry/kube-tide:1.0.0
```

仍建议前面加 Nginx 做 HTTPS。**推荐默认路径仍是 §2 的二进制 + systemd**，更简单、易排查。

## 11. 备份与恢复

| 数据 | 是否持久化 | 备份建议 |
|------|------------|----------|
| 已注册集群 | 否（内存） | 安全保管 kubeconfig；记录集群名称与添加方式 |
| Pod 指标缓存 | 否（内存） | 无需备份，重启后重新采集 |
| 配置文件 | 是（文件） | 纳入 Git 或配置管理 |
| 日志 | 是（文件） | 日志平台保留策略 |

**恢复流程（进程崩溃 / 重启）**：

1. 重启 `kube-tide-prod`
2. 重新添加各集群 kubeconfig
3. 验证 `/api/clusters` 与抽样资源列表

## 12. 升级

```bash
# 1. 记录当前版本与已注册集群清单
# 2. 优雅停止旧进程（SIGTERM）
# 3. 构建新版本
make build-prod
# 4. 替换二进制并启动
# 5. 重新注册集群
# 6. 冒烟测试：/api/health、集群列表、Pod 列表、Deployment 详情
```

回滚：保留上一版 `kube-tide-prod` 二进制，重复上述步骤。

## 13. 故障排查

| 现象 | 可能原因 | 处理 |
|------|----------|------|
| 8080 无法访问 | 端口占用 / 绑定地址 | 检查 `server.host/port`、防火墙 |
| 添加集群失败 | kubeconfig 无效 / 网络不通 | 在服务器上 `kubectl --kubeconfig=... cluster-info` |
| 重启后集群消失 | 设计限制 | 重新添加；见 §10 |
| Pod 终端连不上 | 代理未升级 WebSocket | 检查 Nginx `Upgrade` 头；超时是否过短 |
| 指标图表为空 | metrics-server 未装 / 采集未启动 | 集群安装 metrics-server；查看日志「Pod指标收集」 |
| 前端刷新 404 | 非 prod 模式或未 embed | 使用 `make build-prod`；生产勿用 dev 静态路径 |
| 日志文件未生成 | `file.enabled: false` 或无写权限 | 检查配置与 `logs/` 目录权限 |
| Autoscaler API 异常 | 集群未注册或 RBAC 不足 | 确认集群已添加；检查 autoscaler 相关权限 |

### 常用诊断命令

```bash
# 健康检查
curl -s http://127.0.0.1:8080/api/health | jq .

# 集群列表（重启后可能为空）
curl -s http://127.0.0.1:8080/api/clusters | jq .

# 查看最近日志
tail -f logs/kube-tide.log
```

## 14. 生产安全基线（推荐）

- [ ] 反向代理 + TLS + 认证
- [ ] 非 root 运行进程
- [ ] kubeconfig 最小 RBAC
- [ ] 文件日志轮转与集中采集
- [ ] 定期升级 client-go 与 K8s 版本匹配
- [ ] 限制 Pod Exec / 日志查看权限（网络层 + K8s RBAC）
- [ ] 制定集群重新注册 Runbook

## 15. 相关文档

- [系统架构](./architecture.md)
- [代码目录结构](./code_arch.md)
- [功能路线图](./TODO.md)
