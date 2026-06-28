# GitHub Copilot 指令

## 项目背景

本项目是一个基于Go语言的Kubernetes平台。在协助编写代码时，请遵循以下指导原则。

## chat 交互语言
- 使用中文进行交流

## 编码标准
- 遵循Go语言的官方代码规范
- 使用有意义的变量和函数名
- 为所有公共API提供详细注释
- 错误处理必须彻底，避免panic
- 使用接口实现松耦合设计
- 不要出现基本的编译性错误
- 前端改动，需要同步调整后端相应的API

## 架构指南
- 遵循 README.md 与 [docs/architecture.md](../docs/architecture.md) 的内容
- 代码结构目录参考 [code_arch.md](../docs/code_arch.md)（当前实际结构）
- 运维部署参考 [operations.md](../docs/operations.md)
- 使用模块化设计，避免单一职责原则的违反
- 避免过度嵌套的代码结构

## Kubernetes相关
- 使用 client-go 库与 Kubernetes 交互
- 资源操作遵循 Kubernetes API 约定（本项目为 Web 控制台，非 Operator/CRD 控制器）
- 考虑多集群支持

## 测试要求
- 所有代码必须有单元测试
- 集成测试应使用测试容器
- 确保测试覆盖率达到70%以上
- 使用表驱动测试方法

## 文档要求
- 所有功能都应在README.md中描述
- 包含架构图和流程图
- API文档应清晰完整
- 提供部署和配置说明
- 功能更新之后需要更新对应的README.md和README.zh-CN.md

## 安全考虑
- 敏感信息不应硬编码
- 使用适当的认证和授权机制
- 遵循最小权限原则
- 避免常见的安全漏洞