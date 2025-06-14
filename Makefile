# OpenAI 多密钥代理服务器 Makefile (Go版本)

# 变量定义
BINARY_NAME=gpt-load
MAIN_PATH=./cmd/gpt-load
BUILD_DIR=./build
VERSION=2.0.0
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -s -w"

# 默认目标
.PHONY: all
all: clean build

# 构建
.PHONY: build
build:
	@echo "🔨 构建 $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ 构建完成: $(BUILD_DIR)/$(BINARY_NAME)"

# 构建所有平台
.PHONY: build-all
build-all: clean
	@echo "🔨 构建所有平台版本..."
	@mkdir -p $(BUILD_DIR)

	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)

	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)

	# macOS AMD64
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)

	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)

	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

	@echo "✅ 所有平台构建完成"

# 运行
.PHONY: run
run:
	@echo "🚀 启动服务器..."
	go run $(MAIN_PATH)

# 开发模式运行
.PHONY: dev
dev:
	@echo "🔧 开发模式启动..."
	go run -race $(MAIN_PATH)

# 测试
.PHONY: test
test:
	@echo "🧪 运行测试..."
	go test -v -race -coverprofile=coverage.out ./...

# 测试覆盖率
.PHONY: coverage
coverage: test
	@echo "📊 生成测试覆盖率报告..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ 覆盖率报告生成: coverage.html"

# 基准测试
.PHONY: bench
bench:
	@echo "⚡ 运行基准测试..."
	go test -bench=. -benchmem ./...

# 代码检查
.PHONY: lint
lint:
	@echo "🔍 代码检查..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️ golangci-lint 未安装，跳过代码检查"; \
		echo "安装命令: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# 格式化代码
.PHONY: fmt
fmt:
	@echo "🎨 格式化代码..."
	go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "💡 建议安装 goimports: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

# 整理依赖
.PHONY: tidy
tidy:
	@echo "📦 整理依赖..."
	go mod tidy
	go mod verify

# 安装依赖
.PHONY: deps
deps:
	@echo "📥 安装依赖..."
	go mod download

# 清理
.PHONY: clean
clean:
	@echo "🧹 清理构建文件..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# 安装到系统
.PHONY: install
install: build
	@echo "📦 安装到系统..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "✅ 安装完成: /usr/local/bin/$(BINARY_NAME)"

# 卸载
.PHONY: uninstall
uninstall:
	@echo "🗑️ 从系统卸载..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "✅ 卸载完成"

# Docker 构建
.PHONY: docker-build
docker-build:
	@echo "🐳 构建 Docker 镜像..."
	docker build -t gpt-load:$(VERSION) .
	docker tag gpt-load:$(VERSION) gpt-load:latest
	@echo "✅ Docker 镜像构建完成"

# Docker 运行（使用预构建镜像）
.PHONY: docker-run
docker-run:
	@echo "🐳 运行 Docker 容器（预构建镜像）..."
	docker run -d \
		--name gpt-load \
		-p 7860:7860 \
		-v $(PWD)/keys.txt:/app/keys.txt:ro \
		-v $(PWD)/.env:/app/.env:ro \
		--restart unless-stopped \
		ghcr.io/tbphp/gpt-load:latest

# Docker 运行（本地构建）
.PHONY: docker-run-local
docker-run-local:
	@echo "🐳 运行 Docker 容器（本地构建）..."
	docker run -d \
		--name gpt-load-local \
		-p 7860:7860 \
		-v $(PWD)/keys.txt:/app/keys.txt:ro \
		-v $(PWD)/.env:/app/.env:ro \
		--restart unless-stopped \
		gpt-load:latest

# Docker Compose 运行（预构建镜像）
.PHONY: compose-up
compose-up:
	@echo "🐳 使用 Docker Compose 启动（预构建镜像）..."
	docker-compose up -d

# Docker Compose 运行（本地构建）
.PHONY: compose-up-dev
compose-up-dev:
	@echo "🐳 使用 Docker Compose 启动（本地构建）..."
	docker-compose -f docker-compose.dev.yml up -d

# Docker Compose 停止
.PHONY: compose-down
compose-down:
	@echo "🐳 停止 Docker Compose..."
	docker-compose down
	docker-compose -f docker-compose.dev.yml down 2>/dev/null || true

# 密钥验证
.PHONY: validate-keys
validate-keys:
	@echo "🐍 使用 Python 版本验证密钥..."
	python3 scripts/validate-keys.py -c 100 -t 15

# 健康检查
.PHONY: health
health:
	@echo "💚 健康检查..."
	@curl -s http://localhost:7860/health | jq . || echo "请安装 jq 或检查服务是否运行"

# 查看统计
.PHONY: stats
stats:
	@echo "📊 查看统计信息..."
	@curl -s http://localhost:7860/stats | jq . || echo "请安装 jq 或检查服务是否运行"

# 重置密钥
.PHONY: reset-keys
reset-keys:
	@echo "🔄 重置密钥状态..."
	@curl -s http://localhost:7860/reset-keys | jq . || echo "请安装 jq 或检查服务是否运行"

# 查看黑名单
.PHONY: blacklist
blacklist:
	@echo "🚫 查看黑名单..."
	@curl -s http://localhost:7860/blacklist | jq . || echo "请安装 jq 或检查服务是否运行"

# 帮助
.PHONY: help
help:
	@echo "OpenAI 多密钥代理服务器 v$(VERSION) - 可用命令:"
	@echo ""
	@echo "构建相关:"
	@echo "  build      - 构建二进制文件"
	@echo "  build-all  - 构建所有平台版本"
	@echo "  clean      - 清理构建文件"
	@echo ""
	@echo "运行相关:"
	@echo "  run        - 运行服务器"
	@echo "  dev        - 开发模式运行"
	@echo ""
	@echo "测试相关:"
	@echo "  test       - 运行测试"
	@echo "  coverage   - 生成测试覆盖率报告"
	@echo "  bench      - 运行基准测试"
	@echo ""
	@echo "代码质量:"
	@echo "  lint       - 代码检查"
	@echo "  fmt        - 格式化代码"
	@echo "  tidy       - 整理依赖"
	@echo ""
	@echo "安装相关:"
	@echo "  install    - 安装到系统"
	@echo "  uninstall  - 从系统卸载"
	@echo ""
	@echo "Docker 相关:"
	@echo "  docker-build     - 构建 Docker 镜像"
	@echo "  docker-run       - 运行 Docker 容器（预构建镜像）"
	@echo "  docker-run-local - 运行 Docker 容器（本地构建）"
	@echo "  compose-up       - Docker Compose 启动（预构建镜像）"
	@echo "  compose-up-dev   - Docker Compose 启动（本地构建）"
	@echo "  compose-down     - Docker Compose 停止"
	@echo ""
	@echo "管理相关:"
	@echo "  health     - 健康检查"
	@echo "  stats      - 查看统计信息"
	@echo "  reset-keys - 重置密钥状态"
	@echo "  blacklist  - 查看黑名单"
	@echo ""
	@echo "密钥验证:"
	@echo "  validate-keys        - 验证 API 密钥"
