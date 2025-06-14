# GPT-Load

中文文档 | [English](README.md)

![Docker Build](https://github.com/tbphp/gpt-load/actions/workflows/docker-build.yml/badge.svg)
![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)

一个高性能的 OpenAI 兼容 API 多密钥轮询代理服务器，支持负载均衡，使用 Go 语言开发。

## 功能特性

- **多密钥轮询**: 自动 API 密钥轮换和负载均衡
- **多目标负载均衡**: 支持轮询多个上游 API 地址
- **智能拉黑**: 区分永久性和临时性错误，智能密钥管理
- **实时监控**: 全面的统计信息、健康检查和黑名单管理
- **灵活配置**: 基于环境变量的配置，支持 .env 文件
- **CORS 支持**: 完整的跨域请求支持
- **结构化日志**: 详细的日志记录，包含响应时间和密钥信息
- **可选认证**: 项目级 Bearer token 认证
- **高性能**: 零拷贝流式传输、并发处理和原子操作
- **生产就绪**: 优雅关闭、错误恢复和内存管理

## 快速开始

### 环境要求

- Go 1.21+ (源码构建)
- Docker (容器化部署)

### 方式一：使用 Docker（推荐）

```bash
# 拉取最新镜像
docker pull ghcr.io/tbphp/gpt-load:latest

# 创建密钥文件，每行一个 API 密钥
echo "sk-your-api-key-1" > keys.txt
echo "sk-your-api-key-2" >> keys.txt

# 运行容器
docker run -d -p 7860:7860 \
  -v $(pwd)/keys.txt:/app/keys.txt:ro \
  --name gpt-load \
  ghcr.io/tbphp/gpt-load:latest
```

### 方式二：使用 Docker Compose

```bash
# 启动服务
docker-compose up -d

# 停止服务
docker-compose down
```

### 方式三：源码构建

```bash
# 克隆并构建
git clone https://github.com/tbphp/gpt-load.git
cd gpt-load
go mod tidy

# 创建配置
cp .env.example .env
echo "sk-your-api-key" > keys.txt

# 运行
make run
```

## 配置说明

### 支持的 API 提供商

此代理服务器支持任何 OpenAI 兼容的 API，包括：

- **OpenAI**: `https://api.openai.com`
- **Azure OpenAI**: `https://your-resource.openai.azure.com`
- **Anthropic Claude**: `https://api.anthropic.com` (兼容端点)
- **第三方提供商**: 任何实现 OpenAI API 格式的服务

### 环境变量

复制示例配置文件并根据需要修改：

```bash
cp .env.example .env
```

### 主要配置选项

| 配置项         | 环境变量                           | 默认值                      | 说明                                               |
| -------------- | ---------------------------------- | --------------------------- | -------------------------------------------------- |
| 服务器端口     | `PORT`                             | 7860                        | 服务器监听端口                                     |
| 服务器主机     | `HOST`                             | 0.0.0.0                     | 服务器绑定地址                                     |
| 密钥文件       | `KEYS_FILE`                        | keys.txt                    | API 密钥文件路径                                   |
| 起始索引       | `START_INDEX`                      | 0                           | 密钥轮换起始索引                                   |
| 拉黑阈值       | `BLACKLIST_THRESHOLD`              | 1                           | 拉黑前的错误次数                                   |
| 最大重试次数   | `MAX_RETRIES`                      | 3                           | 使用不同密钥的最大重试次数                         |
| 上游地址       | `OPENAI_BASE_URL`                  | `https://api.openai.com`    | OpenAI 兼容 API 基础地址。支持多个地址，用逗号分隔 |
| 最大并发请求数 | `MAX_CONCURRENT_REQUESTS`          | 100                         | 最大并发请求数                                     |
| 启用 Gzip 压缩 | `ENABLE_GZIP`                      | true                        | 启用响应 Gzip 压缩                                 |
| 认证密钥       | `AUTH_KEY`                         | -                           | 可选的认证密钥                                     |
| 启用 CORS      | `ENABLE_CORS`                      | true                        | 启用 CORS 支持                                     |
| 允许的来源     | `ALLOWED_ORIGINS`                  | \*                          | CORS 允许的来源（逗号分隔，\* 表示允许所有）       |
| 允许的方法     | `ALLOWED_METHODS`                  | GET,POST,PUT,DELETE,OPTIONS | CORS 允许的 HTTP 方法                              |
| 允许的头部     | `ALLOWED_HEADERS`                  | \*                          | CORS 允许的头部（逗号分隔，\* 表示允许所有）       |
| 允许凭证       | `ALLOW_CREDENTIALS`                | false                       | CORS 允许凭证                                      |
| 日志级别       | `LOG_LEVEL`                        | info                        | 日志级别（debug, info, warn, error）               |
| 日志格式       | `LOG_FORMAT`                       | text                        | 日志格式（text, json）                             |
| 启用文件日志   | `LOG_ENABLE_FILE`                  | false                       | 启用文件日志                                       |
| 日志文件路径   | `LOG_FILE_PATH`                    | logs/app.log                | 日志文件路径                                       |
| 启用请求日志   | `LOG_ENABLE_REQUEST`               | true                        | 启用请求日志（生产环境可设为 false 以提高性能）    |
| 服务器读取超时 | `SERVER_READ_TIMEOUT`              | 120                         | HTTP 服务器读取超时时间（秒）                      |
| 服务器写入超时 | `SERVER_WRITE_TIMEOUT`             | 1800                        | HTTP 服务器写入超时时间（秒）                      |
| 服务器空闲超时 | `SERVER_IDLE_TIMEOUT`              | 120                         | HTTP 服务器空闲超时时间（秒）                      |
| 优雅关闭超时   | `SERVER_GRACEFUL_SHUTDOWN_TIMEOUT` | 60                          | 服务器优雅关闭超时时间（秒）                       |
| 请求超时       | `REQUEST_TIMEOUT`                  | 30                          | 请求超时时间（秒）                                 |
| 响应超时       | `RESPONSE_TIMEOUT`                 | 30                          | 响应超时时间（秒）- 控制 TLS 握手和响应头接收      |
| 空闲连接超时   | `IDLE_CONN_TIMEOUT`                | 120                         | 空闲连接超时时间（秒）                             |

### 配置示例

#### OpenAI（默认）

```bash
OPENAI_BASE_URL=https://api.openai.com
# 使用 OpenAI API 密钥: sk-...
```

#### Azure OpenAI

```bash
OPENAI_BASE_URL=https://your-resource.openai.azure.com
# 使用 Azure API 密钥，根据需要调整端点
```

#### 第三方提供商

```bash
OPENAI_BASE_URL=https://api.your-provider.com
# 使用提供商特定的 API 密钥
```

#### 多目标负载均衡

```bash
# 使用逗号分隔多个目标地址
OPENAI_BASE_URL=https://gateway.ai.cloudflare.com/v1/.../openai,https://api.openai.com/v1,https://api.another-provider.com/v1
```

## API 密钥验证

项目包含高性能的 API 密钥验证工具：

```bash
# 自动验证密钥
make validate-keys

# 或直接运行
./scripts/validate-keys.py
```

## 监控端点

| 端点          | 方法 | 说明               |
| ------------- | ---- | ------------------ |
| `/health`     | GET  | 健康检查和基本状态 |
| `/stats`      | GET  | 详细统计信息       |
| `/blacklist`  | GET  | 黑名单信息         |
| `/reset-keys` | GET  | 重置所有密钥状态   |

## 开发

### 可用命令

```bash
# 构建
make build      # 构建二进制文件
make build-all  # 构建所有平台版本
make clean      # 清理构建文件

# 运行
make run        # 运行服务器
make dev        # 开发模式（启用竞态检测）

# 测试
make test       # 运行测试
make coverage   # 生成覆盖率报告
make bench      # 运行基准测试

# 代码质量
make lint       # 代码检查
make fmt        # 格式化代码
make tidy       # 整理依赖

# 管理
make health     # 健康检查
make stats      # 查看统计信息
make reset-keys # 重置密钥状态
make blacklist  # 查看黑名单

# 帮助
make help       # 显示所有命令
```

### 项目结构

```text
/
├── cmd/
│   └── gpt-load/
│       └── main.go          # 主入口文件
├── internal/
│   ├── config/
│   │   └── manager.go       # 配置管理
│   ├── errors/
│   │   └── errors.go        # 自定义错误类型
│   ├── handler/
│   │   └── handler.go       # HTTP 处理器
│   ├── keymanager/
│   │   └── manager.go       # 密钥管理器
│   ├── middleware/
│   │   └── middleware.go    # HTTP 中间件
│   └── proxy/
│       └── server.go        # 代理服务器核心
├── pkg/
│   └── types/
│       └── interfaces.go    # 通用接口和类型
├── scripts/
│   └── validate-keys.py     # 密钥验证脚本
├── .github/
│   └── workflows/
│       └── docker-build.yml # GitHub Actions CI/CD
├── build/                   # 构建输出目录
├── .env.example            # 配置文件模板
├── Dockerfile              # Docker 构建文件
├── docker-compose.yml      # Docker Compose 配置
├── Makefile               # 构建脚本
├── go.mod                 # Go 模块文件
├── LICENSE                # MIT 许可证
└── README.md              # 项目文档
```

## 架构

### 性能特性

- **并发处理**: 利用 Go 的 goroutine 实现高并发
- **内存效率**: 零拷贝流式传输，最小化内存分配
- **连接池**: HTTP/2 支持，优化连接复用
- **原子操作**: 无锁并发操作
- **预编译模式**: 启动时编译正则表达式模式

### 安全性与可靠性

- **内存安全**: Go 的内置内存安全防止缓冲区溢出
- **并发安全**: 使用 sync.Map 和原子操作保证线程安全
- **错误处理**: 全面的错误处理和恢复机制
- **资源管理**: 自动清理防止资源泄漏

## 🌟 Star History

[![Star History Chart](https://api.star-history.com/svg?repos=tbphp/gpt-load&type=Date)](https://star-history.com/#tbphp/gpt-load&Date)

## 许可证

MIT 许可证 - 详情请参阅 [LICENSE](LICENSE) 文件。
