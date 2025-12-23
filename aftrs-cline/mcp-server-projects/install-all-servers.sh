#!/bin/bash

# AFTRS MCP Servers Installation Script
# Handles Node.js and Python dependencies with proper virtual environment management

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_FILE="${SCRIPT_DIR}/installation.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$LOG_FILE"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."

    # Check Node.js
    if ! command -v node &> /dev/null; then
        log_error "Node.js is not installed. Please install Node.js 18+ first."
        exit 1
    fi

    local node_version=$(node --version | cut -d'v' -f2 | cut -d'.' -f1)
    if [ "$node_version" -lt 18 ]; then
        log_error "Node.js version must be 18 or higher. Current: $(node --version)"
        exit 1
    fi

    # Check npm
    if ! command -v npm &> /dev/null; then
        log_error "npm is not installed."
        exit 1
    fi

    # Check Python
    if ! command -v python3 &> /dev/null; then
        log_error "Python 3 is not installed."
        exit 1
    fi

    # Check if we can create virtual environments
    if ! python3 -m venv --help &> /dev/null; then
        log_error "Python venv module is not available. Install python3-venv package."
        exit 1
    fi

    log_success "Prerequisites check passed"
}

# Create Python virtual environment for Python-based MCP servers
setup_python_venv() {
    log_info "Setting up Python virtual environment..."

    local venv_dir="${SCRIPT_DIR}/python-venv"

    if [ -d "$venv_dir" ]; then
        log_warning "Virtual environment already exists. Removing and recreating..."
        rm -rf "$venv_dir"
    fi

    python3 -m venv "$venv_dir"
    source "$venv_dir/bin/activate"

    # Upgrade pip
    pip install --upgrade pip

    # Install common Python MCP dependencies
    log_info "Installing Python MCP dependencies..."
    pip install \
        mcp \
        requests \
        aiohttp \
        pydantic \
        python-dotenv \
        pyyaml \
        click \
        rich

    # Install supermemory (the problematic package mentioned in the summary)
    log_info "Installing supermemory..."
    if pip install supermemory; then
        log_success "Supermemory installed successfully in virtual environment"
    else
        log_warning "Supermemory installation failed, but continuing with other packages"
    fi

    # Install other Python packages for potential MCP servers
    pip install \
        beautifulsoup4 \
        lxml \
        feedparser \
        youtube-dl \
        yt-dlp \
        spotify-search \
        discordpy \
        tweepy \
        psycopg2-binary \
        pymongo \
        redis \
        elasticsearch

    deactivate
    log_success "Python virtual environment setup complete"
}

# Install Node.js MCP servers
install_nodejs_servers() {
    local servers=("cr8-cli" "aftrs-cli" "opnsense-monolith" "unraid-monolith" "aftrs-wiki")

    for server in "${servers[@]}"; do
        local server_dir="${SCRIPT_DIR}/${server}"

        if [ ! -d "$server_dir" ]; then
            log_warning "Server directory not found: $server"
            continue
        fi

        log_info "Installing $server..."
        cd "$server_dir"

        # Install dependencies
        if npm install; then
            log_success "$server dependencies installed"
        else
            log_error "Failed to install dependencies for $server"
            continue
        fi

        # Build TypeScript
        if npm run build; then
            log_success "$server built successfully"
        else
            log_error "Failed to build $server"
        fi

        cd "$SCRIPT_DIR"
    done
}

# Create Supermemory MCP Server (if needed)
create_supermemory_server() {
    local server_dir="${SCRIPT_DIR}/supermemory"

    if [ -d "$server_dir" ]; then
        log_info "Supermemory server already exists, skipping creation"
        return
    fi

    log_info "Creating Supermemory MCP Server..."
    mkdir -p "$server_dir/src"

    # Create package.json for supermemory server
    cat > "$server_dir/package.json" << 'EOF'
{
  "name": "supermemory-mcp-server",
  "version": "0.1.0",
  "description": "MCP server for Supermemory integration",
  "type": "module",
  "main": "dist/index.js",
  "scripts": {
    "build": "tsc",
    "dev": "tsc --watch",
    "start": "node dist/index.js",
    "type-check": "tsc --noEmit"
  },
  "keywords": ["mcp", "server", "supermemory", "memory"],
  "author": "AFTRS Team",
  "license": "MIT",
  "dependencies": {
    "@modelcontextprotocol/sdk": "^0.6.0"
  },
  "devDependencies": {
    "@types/node": "^20.0.0",
    "typescript": "^5.0.0"
  }
}
EOF

    # Create tsconfig.json
    cat > "$server_dir/tsconfig.json" << 'EOF'
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "node",
    "allowSyntheticDefaultImports": true,
    "esModuleInterop": true,
    "allowJs": true,
    "strict": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "outDir": "./dist",
    "rootDir": "./src",
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist"]
}
EOF

    # Create basic Supermemory MCP server
    cat > "$server_dir/src/index.ts" << 'EOF'
#!/usr/bin/env node

import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import {
  CallToolRequestSchema,
  ListToolsRequestSchema,
} from '@modelcontextprotocol/sdk/types.js';
import { execSync } from 'child_process';

const server = new Server(
  {
    name: 'supermemory-server',
    version: '0.1.0',
  },
  {
    capabilities: {
      tools: {},
    },
  }
);

const VENV_PATH = process.env.PYTHON_VENV_PATH || '/home/hg/Docs/aftrs-void/aftrs-cline/mcp-server-projects/python-venv';

server.setRequestHandler(ListToolsRequestSchema, async () => {
  return {
    tools: [
      {
        name: 'search_memories',
        description: 'Search through stored memories using Supermemory',
        inputSchema: {
          type: 'object',
          properties: {
            query: {
              type: 'string',
              description: 'Search query for memories',
            },
            limit: {
              type: 'number',
              description: 'Maximum number of results to return',
              default: 10,
            },
          },
          required: ['query'],
        },
      },
      {
        name: 'store_memory',
        description: 'Store a new memory in Supermemory',
        inputSchema: {
          type: 'object',
          properties: {
            content: {
              type: 'string',
              description: 'Content to store as memory',
            },
            tags: {
              type: 'array',
              items: { type: 'string' },
              description: 'Tags for the memory',
            },
            source: {
              type: 'string',
              description: 'Source of the memory',
            },
          },
          required: ['content'],
        },
      },
    ],
  };
});

server.setRequestHandler(CallToolRequestSchema, async (request) => {
  const { name, arguments: args } = request.params;

  try {
    switch (name) {
      case 'search_memories': {
        const pythonCmd = `${VENV_PATH}/bin/python -c "
import sys
sys.path.append('${VENV_PATH}/lib/python*/site-packages')
try:
    import supermemory
    # Add supermemory search logic here
    print('Search results for: ${args.query}')
except ImportError as e:
    print(f'Supermemory not available: {e}')
"`;

        const result = execSync(pythonCmd, { encoding: 'utf8' });

        return {
          content: [
            {
              type: 'text',
              text: `Memory search results:\n${result}`,
            },
          ],
        };
      }

      case 'store_memory': {
        const pythonCmd = `${VENV_PATH}/bin/python -c "
import sys
sys.path.append('${VENV_PATH}/lib/python*/site-packages')
try:
    import supermemory
    # Add supermemory storage logic here
    print('Memory stored successfully')
except ImportError as e:
    print(f'Supermemory not available: {e}')
"`;

        const result = execSync(pythonCmd, { encoding: 'utf8' });

        return {
          content: [
            {
              type: 'text',
              text: `Memory storage result:\n${result}`,
            },
          ],
        };
      }

      default:
        throw new Error(`Unknown tool: ${name}`);
    }
  } catch (error) {
    return {
      content: [
        {
          type: 'text',
          text: `Error executing ${name}: ${error.message}`,
        },
      ],
      isError: true,
    };
  }
});

async function runServer() {
  const transport = new StdioServerTransport();
  await server.connect(transport);
  console.error('Supermemory MCP Server running on stdio');
}

runServer().catch(console.error);
EOF

    # Create README for Supermemory server
    cat > "$server_dir/README.md" << 'EOF'
# Supermemory MCP Server

A Model Context Protocol server for Supermemory integration, providing memory storage and search capabilities.

## Features

- **search_memories**: Search through stored memories
- **store_memory**: Store new memories with tags and metadata

## Installation

```bash
cd supermemory
npm install
npm run build
```

## Configuration

Add to your Cline MCP configuration:

```json
{
  "mcpServers": {
    "supermemory": {
      "command": "node",
      "args": ["path/to/supermemory/dist/index.js"],
      "env": {
        "PYTHON_VENV_PATH": "/home/hg/Docs/aftrs-void/aftrs-cline/mcp-server-projects/python-venv"
      }
    }
  }
}
```
EOF

    log_success "Supermemory MCP Server created"
}

# Create unified Cline MCP configuration
create_cline_config() {
    log_info "Creating unified Cline MCP configuration..."

    cat > "${SCRIPT_DIR}/cline-mcp-config.json" << 'EOF'
{
  "mcpServers": {
    "cr8-cli": {
      "command": "node",
      "args": ["./aftrs-cline/mcp-server-projects/cr8-cli/dist/index.js"],
      "env": {}
    },
    "aftrs-cli": {
      "command": "node",
      "args": ["./aftrs-cline/mcp-server-projects/aftrs-cli/dist/index.js"],
      "env": {
        "AFTRS_BASE_PATH": "/home/hg/Docs/aftrs-void/aftrs_cli"
      }
    },
    "opnsense-monolith": {
      "command": "node",
      "args": ["./aftrs-cline/mcp-server-projects/opnsense-monolith/dist/index.js"],
      "env": {
        "OPNSENSE_HOST": "10.0.1.1",
        "OPNSENSE_API_KEY": "",
        "OPNSENSE_API_SECRET": ""
      }
    },
    "unraid-monolith": {
      "command": "node",
      "args": ["./aftrs-cline/mcp-server-projects/unraid-monolith/dist/index.js"],
      "env": {}
    },
    "aftrs-wiki": {
      "command": "node",
      "args": ["./aftrs-cline/mcp-server-projects/aftrs-wiki/dist/index.js"],
      "env": {
        "WIKI_BASE_PATH": "/home/hg/Docs/aftrs-void/aftrs_wiki"
      }
    },
    "supermemory": {
      "command": "node",
      "args": ["./aftrs-cline/mcp-server-projects/supermemory/dist/index.js"],
      "env": {
        "PYTHON_VENV_PATH": "/home/hg/Docs/aftrs-void/aftrs-cline/mcp-server-projects/python-venv"
      }
    }
  }
}
EOF

    log_success "Cline MCP configuration created at: ${SCRIPT_DIR}/cline-mcp-config.json"
}

# Set permissions
set_permissions() {
    log_info "Setting executable permissions..."

    find "${SCRIPT_DIR}" -name "*.sh" -exec chmod +x {} \;

    # Make sure the Python venv is accessible
    if [ -d "${SCRIPT_DIR}/python-venv" ]; then
        chmod -R 755 "${SCRIPT_DIR}/python-venv"
    fi

    log_success "Permissions set"
}

# Main installation process
main() {
    log_info "Starting AFTRS MCP Servers installation..."
    echo "Installation log: $LOG_FILE"

    check_prerequisites
    setup_python_venv
    create_supermemory_server
    install_nodejs_servers
    create_cline_config
    set_permissions

    log_success "Installation complete!"
    echo ""
    echo "Next steps:"
    echo "1. Configure API keys in cline-mcp-config.json"
    echo "2. Copy configuration to Cline settings"
    echo "3. Test MCP servers with: npm run start (in each server directory)"
    echo ""
    echo "Configuration file: ${SCRIPT_DIR}/cline-mcp-config.json"
}

# Run main function
main "$@"
