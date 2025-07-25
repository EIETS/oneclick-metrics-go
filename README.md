# OneClick Metrics Exporter (Go)

## 📌 项目简介

OneClick Metrics Exporter 是一个使用 Go 编写的服务，用于从 PostgreSQL 数据库中定时采集 Pull Request（PR）相关指标，并通过 Prometheus 暴露 `/metrics` 接口，供监控系统使用。

## 🚀 功能说明

- 定时从数据库中采集 PR 状态、报告覆盖率等指标
- 使用 Prometheus 客户端注册并暴露指标
- 支持标签化指标（如 project、pr_state、report 等）
- 可配置采集间隔、数据库连接信息、日志等级等

## 🛠️ 安装依赖

```bash
go mod init oneclick-metrics-go
go get github.com/jackc/pgx/v5
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promhttp
```

## ⚙️ 环境变量说明
你可以通过设置以下环境变量来配置数据库连接：

DATABASE_URL：PostgreSQL 连接字符串，例如：
```url
postgres://user:password@localhost:5432/dbname
```

## ▶️ 运行方式
```go
go run cmd/main.go
```

启动后，Prometheus 指标服务将运行在 http://localhost:8000/metrics。

## 📊 示例输出
访问 /metrics 后，你将看到类似输出：
```bash
# HELP oneclick_dummy_metric A dummy metric for testing
# TYPE oneclick_dummy_metric gauge
oneclick_dummy_metric{label="test"} 1
```

## 🧭 总体目标
将 Python 实现的 OneClick Metrics Exporter 重构为 Go 项目，具备以下能力：

定时从 PostgreSQL 拉取 PR 数据
处理并聚合指标（如 PR 状态、报告覆盖率等）
暴露 Prometheus 指标接口
支持配置、日志、模块化结构

## 🗂️ 阶段划分与任务清单
### ✅ 阶段 1：项目初始化与基础搭建（1 天）
#### 📌 任务
- 初始化 Go 模块：go mod init oneclick-metrics-go
- 创建基础目录结构（你已完成）
- 安装依赖：
```go
go get github.com/jackc/pgx/v5
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promhttp
```

#### 🎯 目标
- 能运行 go run cmd/main.go 并在 http://localhost:8000/metrics 输出一个 dummy 指标

### ✅ 阶段 2：数据库连接模块（1 天）
#### 📌 任务
- 在 db/connection.go 中实现连接池（使用 pgxpool）
- 支持从环境变量或配置文件读取连接信息
- 编写测试函数验证连接是否成功
#### 🎯 目标
- 能成功连接 PostgreSQL 并执行简单查询

### ✅ 阶段 3：Prometheus 指标注册与暴露（1 天）
#### 📌 任务
- 在 metrics/exporter.go 中注册多个指标（如 PR 状态、报告数量）
- 使用 GaugeVec 定义指标结构
- 在 main.go 中定时调用采集函数
#### 🎯 目标
- Prometheus 能拉取多个结构化指标（带标签）

### ✅ 阶段 4：实现 SQL 查询与数据采集逻辑（3~4 天）
#### 📌 任务
- 将 Python 中的 SQL 查询迁移为 Go 字符串常量
- 在 db/queries.go 中封装每个查询函数
- 在 metrics/collectors.go 中实现每类指标的采集逻辑
  - open PR
  - closed PR
  - result report
  - check summary
  - missing report
#### 🎯 目标
- 每类指标都能从数据库中获取真实数据并更新 Prometheus 指标

### ✅ 阶段 5：配置与参数支持（1 天）
#### 📌 任务
- 使用 flag 或 viper 支持命令行参数或配置文件
- 支持设置：
  - 数据库连接信息
  - 拉取间隔
  - 日志等级
  - 时区
#### 🎯 目标
- 项目可配置，适配不同部署环境

### ✅ 阶段 6：日志与错误处理（1 天）
#### 📌 任务
- 使用 log 或 zap 记录关键操作与错误
- 所有数据库操作、指标采集都应有日志输出
- 捕获并处理错误，避免程序崩溃
#### 🎯 目标
- 稳定运行，日志清晰，便于调试

### ✅ 阶段 7：打包与部署准备（1 天）
#### 📌 任务
- 编写 README.md，说明项目功能、结构、运行方式
- 添加 Dockerfile（可选）
- 编写部署脚本或 systemd 服务文件（可选）
#### 🎯 目标
- 项目可部署、可运行、可展示
#### 🧠 附加建议
- ✅ 每完成一个阶段就 git commit，保持版本清晰
- ✅ 可以用 Grafana 连接 Prometheus 展示指标图表
- ✅ 最后在简历中写上你实现的模块、技术栈、项目目标和成效
