# Autosphere AWX/Ansible Integration Guide

This directory contains Ansible playbooks and resources for managing Autosphere infrastructure through AWX automation.

## 🤖 **Available MCP Tools**

### 1. **launch_awx_job**
Launch AWX job templates for various Autosphere operations.

**Parameters:**
- `job_template`: Job template name (e.g., "autosphere-deploy", "health-check", "autosphere-scale")
- `extra_vars`: Additional variables to pass to the job
- `inventory`: Target inventory (optional)
- `limit`: Limit to specific hosts (optional)
- `tags`: Ansible tags to run (optional)
- `skip_tags`: Ansible tags to skip (optional)

**Example:**
```json
{
  "job_template": "autosphere-deploy",
  "extra_vars": {
    "deployment_version": "v2.1.0",
    "strategy": "rolling"
  },
  "inventory": "production"
}
```

### 2. **check_awx_job**
Monitor AWX job execution status and results.

**Parameters:**
- `job_id`: AWX job ID to check

**Example:**
```json
{
  "job_id": 12345
}
```

### 3. **health_check**
Perform comprehensive health checks on Autosphere components.

**Parameters:**
- `component`: Specific component to check ("api", "database", "cache", "web", "workers", "monitoring", "all")
- `deep`: Perform deep health checks (default: false)

**Example:**
```json
{
  "component": "all",
  "deep": true
}
```

### 4. **autoscale**
Manage autoscaling of Autosphere services.

**Parameters:**
- `action`: Scaling action ("scale_up", "scale_down", "analyze", "auto")
- `service`: Service to scale (optional, defaults to "api")
- `replicas`: Target replica count (for manual scaling)
- `threshold`: Scaling threshold ("cpu_high", "memory_high", "load_high")

**Example:**
```json
{
  "action": "auto",
  "service": "api",
  "threshold": "cpu_high"
}
```

## 📋 **Ansible Playbooks**

### 1. **autosphere-health-check.yml**
Comprehensive health monitoring for all Autosphere components.

**Features:**
- API endpoint health checks
- Database connectivity tests
- Cache (Redis) health verification
- System resource monitoring (CPU, memory, disk)
- Service status verification
- Error log analysis
- Automated health reporting

**Variables:**
- `autosphere_api_url`: Base URL for API health checks
- `autosphere_db_*`: Database connection parameters
- `autosphere_redis_*`: Redis connection parameters

### 2. **autosphere-autoscale.yml**
Kubernetes-based auto-scaling for Autosphere services.

**Features:**
- Metrics-based scaling decisions
- Support for scale-up and scale-down operations
- Configurable thresholds and limits
- Deployment rollout monitoring
- Scaling event logging

**Variables:**
- `scale_up_cpu_threshold`: CPU threshold for scaling up (default: 80%)
- `scale_up_memory_threshold`: Memory threshold for scaling up (default: 85%)
- `scale_down_cpu_threshold`: CPU threshold for scaling down (default: 20%)
- `scale_down_memory_threshold`: Memory threshold for scaling down (default: 30%)
- `min_replicas`: Minimum number of replicas (default: 2)
- `max_replicas`: Maximum number of replicas (default: 10)

### 3. **autosphere-deploy.yml**
Production deployment automation with multiple strategies.

**Features:**
- Rolling, blue-green, and canary deployment strategies
- Automated database migrations
- Health check validation
- Backup creation
- Post-deployment testing
- Notification system

**Variables:**
- `deployment_version`: Version to deploy (default: "latest")
- `strategy`: Deployment strategy ("rolling", "blue_green", "canary")
- `create_backup`: Create backup before deployment (default: true)
- `notification_email`: Email for deployment notifications

## 🔧 **AWX Job Templates**

### Recommended AWX Job Template Configuration:

#### 1. **Autosphere Health Check**
- **Name**: `autosphere-health-check`
- **Playbook**: `autosphere-health-check.yml`
- **Inventory**: Production servers
- **Credentials**: SSH key for server access
- **Survey Variables**:
  - `component`: Choice (api, database, cache, web, workers, monitoring, all)
  - `deep`: Boolean (default: false)

#### 2. **Autosphere Deploy**
- **Name**: `autosphere-deploy`
- **Playbook**: `autosphere-deploy.yml`
- **Inventory**: Production servers
- **Credentials**: SSH + Vault for secrets
- **Survey Variables**:
  - `deployment_version`: Text (default: latest)
  - `strategy`: Choice (rolling, blue_green, canary)
  - `create_backup`: Boolean (default: true)

#### 3. **Autosphere Auto-scale**
- **Name**: `autosphere-scale`
- **Playbook**: `autosphere-autoscale.yml`
- **Inventory**: Kubernetes cluster
- **Credentials**: Kubernetes service account
- **Survey Variables**:
  - `scaling_action`: Choice (scale_up, scale_down, auto)
  - `target_service`: Choice (api, workers, web, all)

## 🚀 **Usage Examples**

### Health Check Example:
```bash
# Using MCP Inspector or Claude Desktop
{
  "tool": "health_check",
  "arguments": {
    "component": "all",
    "deep": true
  }
}
```

### Deployment Example:
```bash
# Launch deployment job
{
  "tool": "launch_awx_job",
  "arguments": {
    "job_template": "autosphere-deploy",
    "extra_vars": {
      "deployment_version": "v2.1.0",
      "strategy": "rolling",
      "create_backup": true
    },
    "inventory": "production"
  }
}

# Check deployment status
{
  "tool": "check_awx_job",
  "arguments": {
    "job_id": 12345
  }
}
```

### Auto-scaling Example:
```bash
# Automatic scaling based on metrics
{
  "tool": "autoscale",
  "arguments": {
    "action": "auto",
    "service": "api"
  }
}

# Manual scale up
{
  "tool": "autoscale",
  "arguments": {
    "action": "scale_up",
    "service": "api",
    "replicas": 5
  }
}
```

## 📊 **Monitoring and Alerts**

The health check system monitors:

### 🔍 **Component Status:**
- **API**: Response time, error rates, throughput
- **Database**: Connection pool, query performance, storage
- **Cache**: Hit ratio, memory usage, eviction rate
- **Web**: Active connections, resource usage
- **Workers**: Queue size, job processing rate, failures
- **Monitoring**: Uptime, active alerts, dashboard status

### ⚠️ **Alert Thresholds:**
- **Critical**: >90% CPU, >95% Memory, API errors >5%
- **Warning**: >80% CPU, >85% Memory, API errors >1%
- **Healthy**: Normal operating parameters

## 🔐 **Security Considerations**

1. **AWX Credentials**: Use separate credentials for different environments
2. **Vault Integration**: Store sensitive variables in Ansible Vault
3. **RBAC**: Configure role-based access control in AWX
4. **Audit Logging**: Enable comprehensive logging for all automation jobs
5. **Network Security**: Restrict AWX access to authorized networks

## 🛠️ **Setup Instructions**

1. **Install AWX**: Deploy AWX in your environment
2. **Configure Inventories**: Set up production, staging, development inventories
3. **Add Credentials**: SSH keys, Kubernetes service accounts, database credentials
4. **Import Playbooks**: Add the playbooks to your AWX project
5. **Create Job Templates**: Configure job templates with survey variables
6. **Test MCP Integration**: Use the MCP tools to trigger and monitor jobs

This integration enables Claude and other AI assistants to perform sophisticated infrastructure automation tasks through natural language interactions!
