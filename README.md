# Autosphere MCP Golang Server - Clean Architecture

A production-ready Model Context Protocol (MCP) server implementation using Go with clean architecture principles, featuring AWX/Ansible automation tools for Autosphere project self-healing and management.

## 🏗️ **Architecture Overview**

This project follows **Clean Architecture** principles with proper separation of concerns:

```
┌─────────────────────────────────────────┐
│                 cmd/                    │ ← Entry Point
│           (main.go)                     │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│              server/                    │ ← Server Layer
│        (MCP Server Setup)               │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│             handlers/                   │ ← Handler Layer
│      (Request/Response Logic)           │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│             services/                   │ ← Business Layer
│         (Business Logic)                │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│              models/                    │ ← Data Layer
│          (Data Structures)              │
└─────────────────────────────────────────┘
```

## 📁 **Project Structure**

```
autosphere-mcp-golang/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/                    # Private application packages
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── interfaces/
│   │   ├── basic.go             # Basic service interfaces
│   │   └── automation.go        # Automation service interfaces
│   ├── models/
│   │   ├── basic.go             # Basic tool data models
│   │   └── automation.go        # Automation tool data models
│   ├── services/
│   │   ├── basic.go             # Basic operations business logic
│   │   ├── basic_test.go        # Unit tests for basic service
│   │   ├── automation.go        # Automation business logic
│   │   └── health.go            # Health check business logic
│   ├── handlers/
│   │   ├── basic.go             # Basic tool MCP handlers
│   │   └── automation.go        # Automation tool handlers
│   └── server/
│       └── server.go            # MCP server setup and configuration
├── awx-resources/               # Ansible playbooks and resources
│   ├── README.md                # AWX integration guide
│   ├── autosphere-health-check.yml  # Health monitoring playbook
│   ├── autosphere-autoscale.yml     # Auto-scaling playbook
│   └── autosphere-deploy.yml        # Deployment automation playbook
├── Dockerfile                   # Container build configuration
├── docker-compose.yml           # Multi-container setup
├── Makefile                     # Build automation
├── build.sh                     # Build script
├── test.sh                      # Testing script
├── ARCHITECTURE.md              # Detailed architecture documentation
├── go.mod                       # Go module definition
├── go.sum                       # Go dependency checksums
└── README.md                    # This file
```

## 🌟 **Features**

### **Basic Tools:**
1. **greet** - Greets a person by name with a friendly message
2. **calculate** - Performs basic mathematical operations (add, subtract, multiply, divide)
3. **get_time** - Gets the current time in a specified timezone

### **🤖 AWX/Ansible Automation Tools:**
4. **launch_awx_job** - Launch AWX job templates for Autosphere automation
5. **check_awx_job** - Monitor AWX job execution status and results
6. **health_check** - Comprehensive health monitoring of Autosphere components
7. **autoscale** - Intelligent autoscaling of Autosphere services

## 🚀 **Quick Start**

### **Prerequisites**
- Go 1.23 or later
- Docker (optional)
- Make (optional)

### **1. Build and Run**

#### Using Make (Recommended)
```bash
# Setup development environment
make dev-setup

# Build the server
make build

# Run with STDIO transport (for Claude Desktop)
make run

# Run with HTTP transport (for web integration)
make run-http

# Run on custom port
make run-http PORT=3000
```

#### Using Build Script
```bash
# Build
./build.sh

# Run STDIO mode
./autosphere-mcp-server

# Run HTTP mode
./autosphere-mcp-server -http localhost:8080
```

#### Using Go Directly
```bash
# Build and run from source
go run ./cmd/server

# With HTTP transport
go run ./cmd/server -http localhost:8080

# With debug logging
go run ./cmd/server -debug -http localhost:8080
```

### **2. Docker Deployment**

```bash
# Build and run with Docker Compose
docker-compose up --build

# Or build Docker image manually
docker build -t autosphere-mcp .
docker run -p 8080:8080 autosphere-mcp
```

### **3. Testing**

```bash
# Run unit tests
make test

# Format and lint code
make fmt
make vet
make lint

# Test build without running
make test-build

# Test with MCP Inspector
./test.sh stdio
./test.sh http 8080
./test.sh inspect 8080  # In another terminal
```

## 🎯 **Design Patterns Applied**

1. **Dependency Injection** - Services injected through interfaces
2. **Repository Pattern** - Services act as business logic repositories
3. **Interface Segregation** - Small, focused interfaces
4. **Single Responsibility** - Each package has one responsibility
5. **Factory Pattern** - Constructor functions for proper initialization
6. **Command Pattern** - Each tool represents a command

## 🧪 **Development Workflow**

### **Adding a New Tool**

1. **Define Models** (`internal/models/`)
   ```go
   type NewToolArgs struct {
       Input string `json:"input" jsonschema:"description"`
   }
   
   type NewToolOutput struct {
       Result string `json:"result" jsonschema:"description"`
   }
   ```

2. **Add Service Interface** (`internal/interfaces/`)
   ```go
   type NewService interface {
       ProcessNewTool(ctx context.Context, args models.NewToolArgs) (models.NewToolOutput, error)
   }
   ```

3. **Implement Service** (`internal/services/`)
   ```go
   func (s *NewService) ProcessNewTool(ctx context.Context, args models.NewToolArgs) (models.NewToolOutput, error) {
       // Business logic here
   }
   ```

4. **Create Handler** (`internal/handlers/`)
   ```go
   func (h *NewHandler) HandleNewTool(ctx context.Context, req *mcp.CallToolRequest, input models.NewToolArgs) (*mcp.CallToolResult, models.NewToolOutput, error) {
       return h.service.ProcessNewTool(ctx, input)
   }
   ```

5. **Register Tool** (`internal/server/server.go`)
   ```go
   mcp.AddTool(s.server, &mcp.Tool{
       Name:        "new_tool",
       Description: "Description of new tool",
   }, s.newHandler.HandleNewTool)
   ```

### **Testing**

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/services

# Run with verbose output
go test -v ./...
```

## 🔧 **Configuration**

The server supports various configuration options:

```bash
# Command line flags
./autosphere-mcp-server \
  -http localhost:8080 \        # Enable HTTP transport
  -debug \                      # Enable debug logging
  -awx-url https://awx.local    # AWX base URL
```

## 🐳 **Docker Support**

### **Multi-stage Dockerfile**
- Optimized build process
- Minimal runtime image
- Non-root user execution
- Health checks included

### **Docker Compose**
- Ready-to-use setup
- Environment configuration
- Network isolation
- Health monitoring

## 📊 **Monitoring and Health Checks**

The server includes comprehensive health monitoring:

### **Component Health Status**
- 🟢 **Healthy**: All systems operational
- 🟡 **Warning**: Performance degradation detected
- 🔴 **Critical**: Immediate attention required
- ⚫ **Unknown**: Component status unavailable

### **Monitored Components**
- API endpoints (response time, error rates)
- Database connectivity and performance
- Cache systems (Redis hit ratios, memory usage)
- Web servers (active connections, resources)
- Background workers (queue sizes, job processing)
- System metrics (CPU, memory, disk usage)

## 🔗 **Integration with Claude Desktop**

### **STDIO Transport (Default)**
```json
{
  "mcpServers": {
    "autosphere-mcp-golang": {
      "command": "/path/to/autosphere-mcp-server",
      "args": [],
      "env": {}
    }
  }
}
```

### **StreamableHTTP Transport**
```json
{
  "mcpServers": {
    "autosphere-mcp-golang-http": {
      "type": "streamable-http",
      "url": "http://localhost:8080"
    }
  }
}
```

## 💬 **Natural Language Examples**

With Claude Desktop integration, you can use natural language:

- *"Check the health of all Autosphere components"*
- *"Scale up the API service if CPU usage is high"*
- *"Deploy version 2.1.0 using rolling strategy"*
- *"Show me the status of the latest deployment job"*
- *"What's the current system performance?"*
- *"Calculate 15% of 1250"*
- *"What time is it in Tokyo?"*

## 🏆 **Benefits of This Architecture**

### **Development Benefits**
- **Testability**: Each layer can be tested independently
- **Maintainability**: Clear separation of concerns
- **Scalability**: Easy to add new features and tools
- **Code Reusability**: Shared interfaces and models

### **Operational Benefits**
- **Self-Healing Infrastructure**: Automated issue detection and resolution
- **AI-Powered Operations**: Natural language infrastructure management
- **Cost Optimization**: Intelligent resource management
- **Enhanced Reliability**: Continuous monitoring and health checks

## 📈 **Performance Considerations**

- Minimal memory footprint
- Fast startup time
- Concurrent request handling
- Efficient resource utilization
- Optional HTTP/2 support

## 🔐 **Security Features**

- Non-root container execution
- Minimal attack surface
- Input validation and sanitization
- Secure configuration management
- Audit logging capabilities

## 📚 **Documentation**

- [ARCHITECTURE.md](ARCHITECTURE.md) - Detailed architecture documentation
- [awx-resources/README.md](awx-resources/README.md) - AWX integration guide
- Inline code documentation
- Comprehensive examples

## 🤝 **Contributing**

1. Fork the repository
2. Create a feature branch
3. Follow the established architecture patterns
4. Add tests for new functionality
5. Run `make fmt vet lint test` before committing
6. Submit a pull request

## 📝 **License**

This project is provided as an example for building production-ready MCP servers with infrastructure automation capabilities. Please refer to the official Go SDK license for usage terms.

---

**🚀 Ready to build self-healing, AI-powered infrastructure with clean, maintainable code!**

For detailed architecture information, see [ARCHITECTURE.md](ARCHITECTURE.md).
