# GPT-Load

[中文文档](README_CN.md) | English

![Docker Build](https://github.com/tbphp/gpt-load/actions/workflows/docker-build.yml/badge.svg)
![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)

A high-performance proxy server for OpenAI-compatible APIs with multi-key rotation and load balancing, built with Go.

## Features

- **Multi-key Rotation**: Automatic API key rotation with load balancing
- **Multi-Target Load Balancing**: Supports round-robin load balancing across multiple upstream API targets
- **Intelligent Blacklisting**: Distinguishes between permanent and temporary errors for smart key management
- **Real-time Monitoring**: Comprehensive statistics, health checks, and blacklist management
- **Flexible Configuration**: Environment-based configuration with .env file support
- **CORS Support**: Full cross-origin request support
- **Structured Logging**: Detailed logging with response times and key information
- **Optional Authentication**: Project-level Bearer token authentication
- **High Performance**: Zero-copy streaming, concurrent processing, and atomic operations
- **Production Ready**: Graceful shutdown, error recovery, and memory management

## Quick Start

### Prerequisites

- Go 1.21+ (for building from source)
- Docker (for containerized deployment)

### Option 1: Using Docker (Recommended)

```bash
# Pull the latest image
docker pull ghcr.io/tbphp/gpt-load:latest

# Create keys.txt file with your API keys (one per line)
echo "sk-your-api-key-1" > keys.txt
echo "sk-your-api-key-2" >> keys.txt

# Run the container
docker run -d -p 7860:7860 \
  -v $(pwd)/keys.txt:/app/keys.txt:ro \
  --name gpt-load \
  ghcr.io/tbphp/gpt-load:latest
```

### Option 2: Using Docker Compose

```bash
# Start the service
docker-compose up -d

# Stop the service
docker-compose down
```

### Option 3: Build from Source

```bash
# Clone and build
git clone https://github.com/tbphp/gpt-load.git
cd gpt-load
go mod tidy

# Create configuration
cp .env.example .env
echo "sk-your-api-key" > keys.txt

# Run
make run
```

## Configuration

### Supported API Providers

This proxy server works with any OpenAI-compatible API, including:

- **OpenAI**: `https://api.openai.com`
- **Azure OpenAI**: `https://your-resource.openai.azure.com`
- **Anthropic Claude**: `https://api.anthropic.com` (with compatible endpoints)
- **Third-party Providers**: Any service implementing OpenAI API format

### Environment Variables

Copy the example configuration file and modify as needed:

```bash
cp .env.example .env
```

### Key Configuration Options

| Setting                 | Environment Variable               | Default                     | Description                                                                                 |
| ----------------------- | ---------------------------------- | --------------------------- | ------------------------------------------------------------------------------------------- |
| Server Port             | `PORT`                             | 7860                        | Server listening port                                                                       |
| Server Host             | `HOST`                             | 0.0.0.0                     | Server binding address                                                                      |
| Keys File               | `KEYS_FILE`                        | keys.txt                    | API keys file path                                                                          |
| Start Index             | `START_INDEX`                      | 0                           | Starting key index for rotation                                                             |
| Blacklist Threshold     | `BLACKLIST_THRESHOLD`              | 1                           | Error count before blacklisting                                                             |
| Max Retries             | `MAX_RETRIES`                      | 3                           | Maximum retry attempts with different keys                                                  |
| Upstream URL            | `OPENAI_BASE_URL`                  | `https://api.openai.com`    | OpenAI-compatible API base URL. Supports multiple, comma-separated URLs for load balancing. |
| Max Concurrent Requests | `MAX_CONCURRENT_REQUESTS`          | 100                         | Maximum number of concurrent requests                                                       |
| Enable Gzip             | `ENABLE_GZIP`                      | true                        | Enable Gzip compression for responses                                                       |
| Auth Key                | `AUTH_KEY`                         | -                           | Optional authentication key                                                                 |
| CORS                    | `ENABLE_CORS`                      | true                        | Enable CORS support                                                                         |
| Allowed Origins         | `ALLOWED_ORIGINS`                  | \*                          | CORS allowed origins (comma-separated, \* for all)                                          |
| Allowed Methods         | `ALLOWED_METHODS`                  | GET,POST,PUT,DELETE,OPTIONS | CORS allowed HTTP methods                                                                   |
| Allowed Headers         | `ALLOWED_HEADERS`                  | \*                          | CORS allowed headers (comma-separated, \* for all)                                          |
| Allow Credentials       | `ALLOW_CREDENTIALS`                | false                       | CORS allow credentials                                                                      |
| Log Level               | `LOG_LEVEL`                        | info                        | Logging level (debug, info, warn, error)                                                    |
| Log Format              | `LOG_FORMAT`                       | text                        | Log format (text, json)                                                                     |
| Enable File Logging     | `LOG_ENABLE_FILE`                  | false                       | Enable logging to file                                                                      |
| Log File Path           | `LOG_FILE_PATH`                    | logs/app.log                | Log file path                                                                               |
| Enable Request Logging  | `LOG_ENABLE_REQUEST`               | true                        | Enable request logging (set to false in production for performance)                         |
| Server Read Timeout     | `SERVER_READ_TIMEOUT`              | 120                         | HTTP server read timeout in seconds                                                         |
| Server Write Timeout    | `SERVER_WRITE_TIMEOUT`             | 1800                        | HTTP server write timeout in seconds                                                        |
| Server Idle Timeout     | `SERVER_IDLE_TIMEOUT`              | 120                         | HTTP server idle timeout in seconds                                                         |
| Graceful Shutdown       | `SERVER_GRACEFUL_SHUTDOWN_TIMEOUT` | 60                          | Graceful shutdown timeout in seconds                                                        |
| Request Timeout         | `REQUEST_TIMEOUT`                  | 30                          | Request timeout in seconds                                                                  |
| Response Timeout        | `RESPONSE_TIMEOUT`                 | 30                          | Response timeout in seconds (TLS handshake & response header)                               |
| Idle Connection Timeout | `IDLE_CONN_TIMEOUT`                | 120                         | Idle connection timeout in seconds                                                          |

### Configuration Examples

#### OpenAI (Default)

```bash
OPENAI_BASE_URL=https://api.openai.com
# Use OpenAI API keys: sk-...
```

#### Azure OpenAI

```bash
OPENAI_BASE_URL=https://your-resource.openai.azure.com
# Use Azure API keys and adjust endpoints as needed
```

#### Third-party Provider

```bash
OPENAI_BASE_URL=https://api.your-provider.com
# Use provider-specific API keys
```

#### Multi-Target Load Balancing

```bash
# Use a comma-separated list of target URLs
OPENAI_BASE_URL=https://gateway.ai.cloudflare.com/v1/.../openai,https://api.openai.com/v1,https://api.another-provider.com/v1
```

## API Key Validation

The project includes a high-performance API key validation tool:

```bash
# Validate keys automatically
make validate-keys

# Or run directly
./scripts/validate-keys.py
```

## Monitoring Endpoints

| Endpoint      | Method | Description                   |
| ------------- | ------ | ----------------------------- |
| `/health`     | GET    | Health check and basic status |
| `/stats`      | GET    | Detailed statistics           |
| `/blacklist`  | GET    | Blacklist information         |
| `/reset-keys` | GET    | Reset all key states          |

## Development

### Available Commands

```bash
# Build
make build      # Build binary
make build-all  # Build for all platforms
make clean      # Clean build files

# Run
make run        # Run server
make dev        # Development mode with race detection

# Test
make test       # Run tests
make coverage   # Generate coverage report
make bench      # Run benchmarks

# Code Quality
make lint       # Code linting
make fmt        # Format code
make tidy       # Tidy dependencies

# Management
make health     # Health check
make stats      # View statistics
make reset-keys # Reset key states
make blacklist  # View blacklist

# Help
make help       # Show all commands
```

### Project Structure

```text
/
├── cmd/
│   └── gpt-load/
│       └── main.go          # Main entry point
├── internal/
│   ├── config/
│   │   └── manager.go       # Configuration management
│   ├── errors/
│   │   └── errors.go        # Custom error types
│   ├── handler/
│   │   └── handler.go       # HTTP handlers
│   ├── keymanager/
│   │   └── manager.go       # Key manager
│   ├── middleware/
│   │   └── middleware.go    # HTTP middleware
│   └── proxy/
│       └── server.go        # Proxy server core
├── pkg/
│   └── types/
│       └── interfaces.go    # Common interfaces and types
├── scripts/
│   └── validate-keys.py     # Key validation script
├── .github/
│   └── workflows/
│       └── docker-build.yml # GitHub Actions CI/CD
├── build/                   # Build output directory
├── .env.example            # Configuration template
├── Dockerfile              # Docker build file
├── docker-compose.yml      # Docker Compose configuration
├── Makefile               # Build scripts
├── go.mod                 # Go module file
├── LICENSE                # MIT License
└── README.md              # Project documentation
```

## Architecture

### Performance Features

- **Concurrent Processing**: Leverages Go's goroutines for high concurrency
- **Memory Efficiency**: Zero-copy streaming with minimal memory allocation
- **Connection Pooling**: HTTP/2 support with optimized connection reuse
- **Atomic Operations**: Lock-free concurrent operations
- **Pre-compiled Patterns**: Regex patterns compiled at startup

### Security & Reliability

- **Memory Safety**: Go's built-in memory safety prevents buffer overflows
- **Concurrent Safety**: Uses sync.Map and atomic operations for thread safety
- **Error Handling**: Comprehensive error handling and recovery mechanisms
- **Resource Management**: Automatic cleanup prevents resource leaks

## 🌟 Star History

[![Star History Chart](https://api.star-history.com/svg?repos=tbphp/gpt-load&type=Date)](https://star-history.com/#tbphp/gpt-load&Date)

## License

MIT License - see [LICENSE](LICENSE) file for details.
