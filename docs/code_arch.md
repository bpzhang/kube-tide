
**代码结构目录**

```textplain

kube-tide/
├── cmd/
│   └── server/
│       └── main.go                # 应用入口点
├── configs/
│   ├── config.go                  # 配置加载
│   └── config.yaml                # 默认配置
├── docs/                          # 项目文档
│   ├── architecture.md            # 系统架构设计
│   ├── code_arch.md               # 代码结构文档
│   └── TODO.md                    # 待办事项清单
├── internal/
│   ├── api/                       # API层
│   │   ├── handlers/              # 请求处理器
│   │   │   ├── auth.go            # 认证处理
│   │   │   ├── cluster.go         # 集群处理
│   │   │   ├── deployment.go      # Deployment处理
│   │   │   ├── node.go            # 节点处理
│   │   │   ├── pod.go             # Pod处理
│   │   │   ├── service.go         # Service处理
│   │   │   └── ws.go              # WebSocket处理
│   │   ├── middleware/            # 中间件
│   │   │   ├── auth.go            # JWT认证
│   │   │   ├── cors.go            # CORS处理
│   │   │   └── logger.go          # 请求日志
│   │   ├── response/              # 响应处理
│   │   │   ├── error.go           # 错误响应
│   │   │   └── success.go         # 成功响应
│   │   └── router.go              # 路由配置
│   ├── core/                      # 核心逻辑
│   │   ├── auth/                  # 认证服务
│   │   │   ├── jwt.go             # JWT实现
│   │   │   └── oidc.go            # OIDC实现
│   │   ├── k8s/                   # K8s交互服务
│   │   │   ├── client.go          # K8s客户端管理
│   │   │   ├── cluster.go         # 集群服务
│   │   │   ├── deployment.go      # Deployment服务
│   │   │   ├── node.go            # 节点服务
│   │   │   ├── pod.go             # Pod服务
│   │   │   └── service.go         # Service服务
│   │   └── ws/                    # WebSocket服务
│   │       ├── client.go          # WS客户端
│   │       ├── hub.go             # WS连接管理
│   │       └── message.go         # WS消息定义
│   ├── models/                    # 数据模型
│   │   ├── cluster.go             # 集群模型
│   │   ├── user.go                # 用户模型
│   │   └── k8s_resource.go        # K8s资源模型
│   ├── repository/                # 数据存储层
│   │   ├── postgres/              # PostgreSQL实现
│   │   │   ├── cluster.go         # 集群存储
│   │   │   └── user.go            # 用户存储
│   │   └── redis/                 # Redis实现
│   │       ├── cache.go           # 缓存服务
│   │       └── ws_session.go      # WS会话存储
│   └── utils/                     # 工具函数
│       ├── logger/                # 日志工具
│       │   └── logger.go          # zap日志配置
│       ├── conversion/            # 转换工具
│       │   └── k8s_resource.go    # K8s资源转换
│       └── errors/                # 错误处理
│           └── error.go           # 错误定义
├── pkg/                           # 公共包
│   ├── constants/                 # 常量定义
│   ├── k8s/                       # K8s工具包
│   │   └── client_helper.go       # K8s客户端辅助函数
│   └── version/                   # 版本信息
├── web/                           # 前端应用
│   ├── public/                    # 静态资源
│   ├── src/                       # 源代码
│   │   ├── api/                   # API请求
│   │   │   ├── auth.ts            # 认证API
│   │   │   ├── cluster.ts         # 集群API
│   │   │   └── ws.ts              # WebSocket客户端
│   │   ├── components/            # 组件
│   │   │   ├── common/            # 通用组件
│   │   │   ├── k8s/               # K8s相关组件
│   │   │   │   ├── ClusterCard.tsx  # 集群卡片
│   │   │   │   ├── PodList.tsx      # Pod列表
│   │   │   │   └── NodeStatus.tsx   # 节点状态
│   │   │   └── charts/            # 图表组件
│   │   │       ├── ResourceUsage.tsx   # 资源使用图表
│   │   │       └── ClusterOverview.tsx # 集群概览图表
│   │   ├── layouts/               # 布局组件
│   │   │   ├── MainLayout.tsx     # 主布局
│   │   │   └── NavigationMenu.tsx # 导航菜单
│   │   ├── pages/                 # 页面组件
│   │   │   ├── Dashboard.tsx      # 仪表盘页面
│   │   │   ├── Clusters.tsx       # 集群管理页面
│   │   │   ├── Nodes.tsx          # 节点管理页面
│   │   │   ├── Workloads/         # 工作负载页面
│   │   │   │   ├── Pods.tsx       # Pod管理页面
│   │   │   │   └── Deployments.tsx # Deployment管理页面
│   │   │   └── Settings.tsx       # 设置页面
│   │   ├── store/                 # Redux状态管理
│   │   │   ├── slices/            # Redux切片
│   │   │   │   ├── authSlice.ts   # 认证状态
│   │   │   │   ├── clusterSlice.ts # 集群状态
│   │   │   │   └── uiSlice.ts     # UI状态
│   │   │   └── store.ts           # Redux配置
│   │   ├── utils/                 # 工具函数
│   │   │   ├── kubernetes.ts      # K8s工具函数
│   │   │   └── format.ts          # 格式化工具
│   │   ├── App.tsx                # 应用入口
│   │   └── index.tsx              # 渲染入口
│   ├── package.json               # 依赖配置
│   └── tsconfig.json              # TypeScript配置
├── deployments/                   # 部署配置
│   ├── docker/                    # Docker配置
│   │   └── Dockerfile             # 构建文件
│   └── kubernetes/                # K8s部署配置
│       ├── deployment.yaml        # 部署定义
│       └── service.yaml           # 服务定义
├── scripts/                       # 脚本工具
│   ├── build.sh                   # 构建脚本
│   └── deploy.sh                  # 部署脚本
├── go.mod                         # Go模块定义
├── go.sum                         # 依赖版本校验
├── Makefile                       # 构建命令
└── README.md                      # 项目说明
```
