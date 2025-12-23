# CLI Design Patterns - AFTRS Development Style

## Overview
This document outlines the established CLI design patterns and conventions used across all AFTRS projects. These patterns ensure consistency, maintainability, and user-friendly interfaces across our command-line tools.

## 🏗️ Core Design Principles

### 1. **Modular Architecture**
- **Library-First**: Core functionality in reusable libraries
- **Script-Based**: Main CLI as thin wrapper around libraries
- **Plugin System**: Extensible architecture for new features
- **Separation of Concerns**: Clear boundaries between components

### 2. **User Experience**
- **Progressive Enhancement**: Basic functionality without dependencies
- **Graceful Degradation**: Fallback mechanisms for missing features
- **Clear Error Messages**: Helpful error messages with next steps
- **Consistent Interface**: Uniform command structure across tools

### 3. **Development Workflow**
- **Git Integration**: Automated version control
- **Configuration-Driven**: YAML/JSON configuration files
- **Environment-Aware**: Automatic environment detection
- **Testing-First**: Comprehensive test coverage

## 📁 Standard Project Structure

```
project_name/
├── 📖 README.md                    # Project documentation
├── 🚀 install.sh                   # Installation script
├── 🔧 main_cli                     # Main CLI executable
├── 📋 requirements.txt             # Python dependencies
├── 📁 lib/                         # Core library functions
│   ├── system.sh                   # System utilities
│   ├── config.sh                   # Configuration management
│   ├── database.sh                 # Database operations
│   └── utils.sh                    # Common utilities
├── 📁 scripts/                     # Automation scripts
├── 📁 tests/                       # Test suite
├── 📁 config/                      # Configuration templates
├── 📁 docs/                        # Documentation
└── 🐳 Dockerfile                   # Container configuration
```

## 🔧 CLI Command Patterns

### 1. **Standard Command Structure**
```bash
# Main command
tool_name <command> [subcommand] [options] [arguments]

# Examples
agentctl status
tsctl doctor
cr8 youtube <url>
oai_cli.py images --url <url>
aftrs_cli dns list
```

### 2. **Common Commands**
```bash
# Status and Health
tool_name status          # Show current status
tool_name health          # Health check
tool_name doctor          # Comprehensive diagnostics

# Configuration
tool_name config          # Show configuration
tool_name setup           # Interactive setup
tool_name configure       # Configure settings

# Management
tool_name list            # List items
tool_name add             # Add item
tool_name remove          # Remove item
tool_name update          # Update item

# System
tool_name backup          # Create backup
tool_name restore         # Restore from backup
tool_name clean           # Cleanup
tool_name reset           # Reset to defaults
```

### 3. **Option Patterns**
```bash
# Standard options
--help, -h               # Show help
--version, -v            # Show version
--verbose, -V            # Verbose output
--debug, -d              # Debug mode
--config, -c <file>      # Config file
--output, -o <format>    # Output format

# Common flags
--force, -f              # Force operation
--dry-run, -n            # Simulate operation
--quiet, -q              # Quiet mode
--json                    # JSON output
--yaml                    # YAML output
```

## 🛠️ Implementation Patterns

### 1. **Bash Script Patterns**

#### Main CLI Script
```bash
#!/bin/bash

# Standard header
set -e
source lib/system.sh
source lib/config.sh

# Help function
show_help() {
    cat << EOF
Usage: $(basename $0) <command> [options]

Commands:
    status      Show current status
    health      Health check
    setup       Interactive setup
    backup      Create backup
    restore     Restore from backup

Options:
    -h, --help      Show this help
    -v, --verbose   Verbose output
    -d, --debug     Debug mode
EOF
}

# Main function
main() {
    case "${1:-}" in
        status) shift; status_command "$@" ;;
        health) shift; health_command "$@" ;;
        setup) shift; setup_command "$@" ;;
        backup) shift; backup_command "$@" ;;
        restore) shift; restore_command "$@" ;;
        -h|--help) show_help ;;
        *) show_help; exit 1 ;;
    esac
}

# Run main
main "$@"
```

#### Library Functions
```bash
# lib/system.sh
check_dependencies() {
    local deps=("$@")
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            echo "❌ Required dependency not found: $dep"
            exit 1
        fi
    done
}

log_info() {
    echo "ℹ️  $1"
}

log_success() {
    echo "✅ $1"
}

log_error() {
    echo "❌ $1" >&2
}

log_warning() {
    echo "⚠️  $1"
}
```

### 2. **Python CLI Patterns**

#### Click-Based CLI
```python
#!/usr/bin/env python3

import click
import yaml
import json
from pathlib import Path

@click.group()
@click.option('--config', '-c', default='config.yaml', help='Config file')
@click.option('--verbose', '-v', is_flag=True, help='Verbose output')
@click.pass_context
def cli(ctx, config, verbose):
    """Main CLI tool"""
    ctx.ensure_object(dict)
    ctx.obj['config'] = load_config(config)
    ctx.obj['verbose'] = verbose

@cli.command()
@click.pass_context
def status(ctx):
    """Show current status"""
    config = ctx.obj['config']
    verbose = ctx.obj['verbose']
    
    if verbose:
        click.echo("Checking system status...")
    
    # Status logic here
    click.echo("✅ System is healthy")

@cli.command()
@click.option('--force', '-f', is_flag=True, help='Force operation')
@click.pass_context
def backup(ctx, force):
    """Create backup"""
    if force or click.confirm('Create backup?'):
        # Backup logic here
        click.echo("✅ Backup created")

if __name__ == '__main__':
    cli()
```

### 3. **Configuration Patterns**

#### YAML Configuration
```yaml
# config.yaml
system:
  name: "AFTRS CLI"
  version: "1.0.0"
  debug: false

database:
  host: "localhost"
  port: 5432
  name: "aftrs_db"
  user: "aftrs_user"

api:
  base_url: "https://api.aftrs.void"
  timeout: 30
  retries: 3

logging:
  level: "INFO"
  file: "logs/aftrs.log"
  format: "json"
```

#### Environment Variables
```bash
# .env
AFTRS_DEBUG=false
AFTRS_CONFIG_PATH=./config.yaml
AFTRS_LOG_LEVEL=INFO
AFTRS_DB_HOST=localhost
AFTRS_DB_PORT=5432
AFTRS_DB_NAME=aftrs_db
AFTRS_DB_USER=aftrs_user
AFTRS_DB_PASSWORD=secret
```

## 🔄 Workflow Patterns

### 1. **Installation Pattern**
```bash
#!/bin/bash
# install.sh

set -e

echo "🚀 Installing AFTRS CLI..."

# Check dependencies
check_dependencies git curl python3

# Create directories
mkdir -p {bin,lib,config,logs,backups}

# Install Python dependencies
pip3 install -r requirements.txt

# Make scripts executable
chmod +x bin/* lib/*.sh

# Setup configuration
cp config/config.example.yaml config/config.yaml

# Create symlink
ln -sf "$(pwd)/bin/main_cli" /usr/local/bin/aftrs_cli

echo "✅ Installation complete!"
```

### 2. **Testing Pattern**
```bash
#!/bin/bash
# tests/test_cli.sh

set -e

echo "🧪 Running CLI tests..."

# Test basic functionality
test_status_command() {
    echo "Testing status command..."
    aftrs_cli status
    assert_exit_code 0
}

# Test error handling
test_error_handling() {
    echo "Testing error handling..."
    aftrs_cli invalid_command
    assert_exit_code 1
}

# Test configuration
test_configuration() {
    echo "Testing configuration..."
    aftrs_cli config show
    assert_exit_code 0
}

# Run tests
test_status_command
test_error_handling
test_configuration

echo "✅ All tests passed!"
```

### 3. **Deployment Pattern**
```bash
#!/bin/bash
# deploy.sh

set -e

echo "🚀 Deploying AFTRS CLI..."

# Build Docker image
docker build -t aftrs_cli .

# Run tests
docker run aftrs_cli tests/test_cli.sh

# Deploy to production
docker-compose up -d

# Health check
sleep 10
curl -f http://localhost:8080/health

echo "✅ Deployment complete!"
```

## 📊 Error Handling Patterns

### 1. **Graceful Error Handling**
```bash
# Error handling in bash
handle_error() {
    local exit_code=$?
    local line_number=$1
    
    echo "❌ Error occurred in line $line_number"
    echo "Exit code: $exit_code"
    
    # Cleanup
    cleanup_temp_files
    
    exit $exit_code
}

trap 'handle_error $LINENO' ERR
```

### 2. **User-Friendly Messages**
```bash
# Clear error messages
if ! command -v required_tool &> /dev/null; then
    echo "❌ Required tool 'required_tool' not found"
    echo "💡 Install it with: sudo apt install required-tool"
    echo "📖 Or visit: https://example.com/install"
    exit 1
fi
```

### 3. **Debug Information**
```bash
# Debug mode
if [[ "${DEBUG:-}" == "true" ]]; then
    set -x
    echo "🔍 Debug mode enabled"
    echo "📁 Working directory: $(pwd)"
    echo "👤 User: $(whoami)"
    echo "🐧 System: $(uname -a)"
fi
```

## 🔧 Integration Patterns

### 1. **Docker Integration**
```dockerfile
# Dockerfile
FROM python:3.11-slim

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    git \
    && rm -rf /var/lib/apt/lists/*

# Copy application
COPY . /app
WORKDIR /app

# Install Python dependencies
RUN pip install -r requirements.txt

# Make scripts executable
RUN chmod +x scripts/*.sh

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Default command
CMD ["python", "main_cli.py"]
```

### 2. **CI/CD Integration**
```yaml
# .github/workflows/ci.yml
name: CI/CD Pipeline

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run tests
        run: |
          chmod +x tests/test_cli.sh
          ./tests/test_cli.sh
      - name: Build Docker image
        run: docker build -t aftrs_cli .
      - name: Test Docker image
        run: docker run aftrs_cli tests/test_cli.sh
```

## 📈 Performance Patterns

### 1. **Caching Strategy**
```bash
# Cache configuration
CACHE_DIR="$HOME/.cache/aftrs_cli"
CACHE_FILE="$CACHE_DIR/config.cache"

# Cache management
cache_get() {
    local key="$1"
    if [[ -f "$CACHE_FILE" ]]; then
        jq -r ".$key" "$CACHE_FILE" 2>/dev/null
    fi
}

cache_set() {
    local key="$1"
    local value="$2"
    mkdir -p "$CACHE_DIR"
    jq ".$key = \"$value\"" "$CACHE_FILE" 2>/dev/null > "$CACHE_FILE.tmp" \
        && mv "$CACHE_FILE.tmp" "$CACHE_FILE"
}
```

### 2. **Parallel Processing**
```bash
# Parallel execution
run_parallel() {
    local max_jobs=4
    local jobs=()
    
    for task in "$@"; do
        # Limit concurrent jobs
        while [[ ${#jobs[@]} -ge $max_jobs ]]; do
            for i in "${!jobs[@]}"; do
                if ! kill -0 "${jobs[$i]}" 2>/dev/null; then
                    unset "jobs[$i]"
                fi
            done
            sleep 0.1
        done
        
        # Start new job
        eval "$task" &
        jobs+=($!)
    done
    
    # Wait for all jobs
    wait
}
```

## 🔮 Future Patterns

### 1. **AI Integration**
```bash
# AI-assisted CLI
ai_suggest() {
    local context="$1"
    local query="$2"
    
    # Use local AI model for suggestions
    ollama run llama2 "Context: $context\nQuery: $query\nSuggest next command:"
}

# Interactive AI mode
ai_mode() {
    while true; do
        read -p "🤖 AI Assistant: " query
        suggestion=$(ai_suggest "$(pwd)" "$query")
        echo "💡 Suggestion: $suggestion"
    done
}
```

### 2. **Plugin System**
```bash
# Plugin management
load_plugins() {
    local plugin_dir="$HOME/.aftrs_cli/plugins"
    
    for plugin in "$plugin_dir"/*.sh; do
        if [[ -f "$plugin" ]]; then
            source "$plugin"
        fi
    done
}

# Plugin interface
plugin_register() {
    local name="$1"
    local command="$2"
    local description="$3"
    
    eval "cmd_$name() { $command \"\$@\"; }"
    echo "Plugin registered: $name - $description"
}
```

## 📚 Related Documentation

- [Docker Deployment Guide](./docker_deployment.md)
- [Git Workflow](./git_workflow.md)
- [AI Integration](./ai_integration.md)
- [Testing Strategies](./testing_strategies.md)

---

*Last Updated: 2025-01-14*
*Status: Production*
*Maintainer: AFTRS Development Team* 

## 🚀 Bulk GitHub Repo Management Pattern

A modern CLI pattern for managing many GitHub repos at once:
- Uses [ghorg](https://github.com/gabrie30/ghorg) for bulk org/user cloning and syncing
- Uses [gum](https://github.com/charmbracelet/gum) for interactive, prettified UX
- Supports both interactive and headless (YAML) automation
- See [Bulk Repo Management Guide](../../aftrs_cli/docs/bulk_repo_management.md) for implementation and usage 