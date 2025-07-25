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
