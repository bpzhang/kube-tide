# Kube-Tide 项目文档

欢迎来到 Kube-Tide 项目文档中心！这里包含了项目的完整文档，帮助您快速了解和使用 Kube-Tide。

## 📚 文档导航

### 🚀 快速开始

- [📖 项目概述](../README.md) - 项目介绍、功能特性和快速开始指南
- [⚡ 快速安装](../README.md#installation-and-usage) - 快速部署和运行指南
- [🔧 开发环境设置](../README.md#development-guide) - 开发环境配置指南

### 🏗️ 系统架构

- [🏛️ 系统架构概览](architecture.md) - 整体架构设计和技术栈
- [📁 代码结构说明](../README.md#directory-structure) - 项目目录结构详解

### 🔐 用户角色体系

- [👥 用户角色体系完整指南](user-role-system.md) - **核心功能文档**
  - 认证系统设计
  - 用户管理功能
  - 角色权限控制
  - 审计日志系统
  - API 设计规范
  - 安装部署指南
  - 开发和测试指南

### 🗄️ 数据库集成

- [📊 数据库集成文档](../README_DATABASE_INTEGRATION.md) - 数据库设计和集成方案
- [🔄 数据库迁移整合说明](MIGRATION_CONSOLIDATION.md) - 迁移文件整合详情
- [🔗 API 集成指南](../README_API_INTEGRATION.md) - API 层实现和使用指南

### 📋 开发指南

- [✅ 待办事项清单](TODO.md) - 项目功能规划和开发计划
- [🛠️ 开发最佳实践](../README.md#contributing) - 代码规范和开发流程

## 📖 文档分类

### 🎯 按用户角色分类

#### 👨‍💼 项目管理者

- [项目概述](../README.md) - 了解项目功能和价值
- [系统架构](architecture.md) - 技术架构和设计决策
- [待办事项](TODO.md) - 项目规划和进度跟踪

#### 👨‍💻 开发人员

- [用户角色体系开发文档](user-role-system.md) - 核心功能实现详解
- [数据库集成文档](../README_DATABASE_INTEGRATION.md) - 数据层设计和实现
- [数据库迁移整合说明](MIGRATION_CONSOLIDATION.md) - 迁移文件整合和管理
- [API 集成指南](../README_API_INTEGRATION.md) - API 层开发指南
- [开发环境设置](../README.md#development-guide) - 开发环境配置

#### 🔧 运维人员

- [安装部署指南](user-role-system.md#安装部署) - 生产环境部署
- [配置管理](user-role-system.md#configuration) - 环境变量和配置选项
- [故障排除](user-role-system.md#故障排除) - 常见问题和解决方案

#### 👥 最终用户

- [快速开始](../README.md#quick-start) - 快速上手指南
- [API 文档](user-role-system.md#api-设计) - API 使用说明
- [安全功能](../README.md#security-features) - 安全特性介绍

### 🏷️ 按功能模块分类

#### 🔐 认证授权

- [JWT 认证系统](user-role-system.md#认证系统)
- [角色权限控制](user-role-system.md#权限控制)
- [审计日志系统](user-role-system.md#审计日志)

#### 🗄️ 数据管理

- [数据模型设计](user-role-system.md#数据模型)
- [数据库迁移](../README_DATABASE_INTEGRATION.md)
- [数据持久化](../README_API_INTEGRATION.md)

#### 🌐 API 接口

- [RESTful API 设计](user-role-system.md#api-设计)
- [认证中间件](user-role-system.md#中间件系统)
- [错误处理](user-role-system.md#api-设计)

#### ☸️ Kubernetes 集成

- [集群管理](../README.md#cluster-management)
- [工作负载管理](../README.md#workload-management)
- [资源监控](../README.md#features)

## 🔍 快速查找

### 常用链接

| 需求 | 文档链接 |
|------|----------|
| 🚀 **快速开始** | [安装指南](../README.md#quick-start) |
| 🔐 **用户管理** | [用户角色体系](user-role-system.md) |
| 🗄️ **数据库配置** | [数据库集成](../README_DATABASE_INTEGRATION.md) |
| 🔄 **数据库迁移** | [迁移整合说明](MIGRATION_CONSOLIDATION.md) |
| 🔗 **API 使用** | [API 集成指南](../README_API_INTEGRATION.md) |
| 🐛 **问题排查** | [故障排除](user-role-system.md#故障排除) |
| 🛠️ **开发指南** | [开发环境](../README.md#development-guide) |

### 技术栈文档

| 技术 | 官方文档 | 项目中的使用 |
|------|----------|--------------|
| **Go** | [golang.org](https://golang.org/doc/) | [开发指南](../README.md#development-guide) |
| **Gin** | [gin-gonic.com](https://gin-gonic.com/docs/) | [API 层实现](user-role-system.md#api-设计) |
| **PostgreSQL** | [postgresql.org](https://www.postgresql.org/docs/) | [数据库设计](user-role-system.md#数据模型) |
| **React** | [reactjs.org](https://reactjs.org/docs/) | [前端开发](../README.md#frontend) |
| **Kubernetes** | [kubernetes.io](https://kubernetes.io/docs/) | [集群管理](../README.md#cluster-management) |

## 📊 项目状态

### ✅ 已完成功能

- [x] **用户角色体系** - 完整的 RBAC 实现
- [x] **数据库集成** - 数据持久化和 API 层
- [x] **认证授权** - JWT 认证和权限控制
- [x] **审计日志** - 操作追踪和安全审计
- [x] **集群管理** - 多集群支持和管理
- [x] **工作负载管理** - Pod、Deployment、Service 管理

### 🚧 开发中功能

- [ ] 用户管理界面
- [ ] 角色权限配置界面
- [ ] 审计日志查看界面
- [ ] ConfigMap 和 Secret 管理
- [ ] 存储管理功能

### 📈 项目指标

| 指标 | 状态 |
|------|------|
| **代码覆盖率** | 目标 >80% |
| **文档完整性** | ✅ 完整 |
| **API 稳定性** | ✅ 稳定 |
| **安全性** | ✅ 企业级 |

## 🤝 贡献指南

### 文档贡献

我们欢迎对文档的改进和补充：

1. **发现错误**: 通过 [Issues](https://github.com/your-org/kube-tide/issues) 报告文档错误
2. **改进建议**: 提出文档改进建议
3. **新增内容**: 贡献新的文档内容
4. **翻译工作**: 帮助翻译文档到其他语言

### 文档规范

- 使用 Markdown 格式
- 遵循项目的文档模板
- 包含代码示例和截图
- 保持文档的及时更新

## 📞 获取帮助

### 联系方式

- 📧 **邮件支持**: <support@kube-tide.com>
- 🐛 **问题报告**: [GitHub Issues](https://github.com/your-org/kube-tide/issues)
- 💬 **社区讨论**: [GitHub Discussions](https://github.com/your-org/kube-tide/discussions)
- 📖 **项目 Wiki**: [GitHub Wiki](https://github.com/your-org/kube-tide/wiki)

### 常见问题

1. **如何快速开始？** - 查看 [快速开始指南](../README.md#quick-start)
2. **如何配置数据库？** - 参考 [数据库配置](user-role-system.md#安装部署)
3. **如何管理用户权限？** - 阅读 [权限控制文档](user-role-system.md#权限控制)
4. **如何部署到生产环境？** - 查看 [部署指南](user-role-system.md#安装部署)

---

## 📝 文档更新日志

| 日期 | 版本 | 更新内容 |
|------|------|----------|
| 2024-12-01 | v1.1.0 | 完成数据库迁移文件整合 |
| 2024-01-01 | v1.0.0 | 完成用户角色体系文档 |
| 2024-01-01 | v1.0.0 | 整合项目文档索引 |
| 2024-01-01 | v1.0.0 | 更新主 README 文档 |

---

*最后更新: 2025-05-28*
*维护者: Kube-Tide 开发团队*
