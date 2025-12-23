# AgentCTL - AI Wiki System & Network Management

## Overview
AgentCTL is a comprehensive, git-based, AI-friendly documentation system that serves as a persistent memory store for development projects, with integrated network management capabilities. The system enables both humans and AI agents to maintain structured knowledge about projects, decisions, and progress with support for multiple AI models and comprehensive network infrastructure management.

## 🏗️ Architecture

### Core Components
- **Flask Web Application**: Modern web interface with Bootstrap UI
- **PostgreSQL Database**: Structured data storage for metadata and content
- **Nginx Reverse Proxy**: Web server and load balancing
- **Supervisor Process Manager**: Keeps all services running
- **Docker Containerization**: Single container deployment
- **Git Integration**: Version control for all documentation
- **Network Management**: Integrated aftrs_cli for comprehensive network infrastructure management

### AI Model Support
- **OpenAI**: GPT-4, GPT-3.5-turbo integration
- **Anthropic**: Claude-3, Claude-2 integration
- **Ollama**: Local models (llama2, codellama)
- **Auto-selection**: Intelligent model selection based on availability

### Network Management Integration
- **Asset-Driven Management**: YAML-based asset inventory with relationships and metadata
- **Real-time Monitoring**: Tailscale session tracking and health diagnostics
- **Advanced Testing**: Comcast DMZ analysis and NAT chain visualization
- **Web Dashboard**: Real-time monitoring interface with custom widgets
- **Performance Optimization**: Parallel execution, intelligent caching, adaptive timeouts
- **Automation Pipeline**: Scheduled tasks and automated remediation

> **Bulk GitHub Repo Management:**
> All install scripts now ensure [ghorg](https://github.com/gabrie30/ghorg) and [gum](https://github.com/charmbracelet/gum) are present for automated, prettified bulk repo management. **All repos are always cloned and synced to `~/Docs` (your home Docs directory).** The script will create and populate this directory if it doesn't exist or is empty. See [Bulk Repo Management Guide](../../aftrs_cli/docs/bulk_repo_management.md) for details.

## 📁 Project Structure

```
agentctl/
├── 🐳 Dockerfile                    # Multi-stage container
├── 🐙 docker-compose.yml            # Production deployment
├── 🚀 deploy.sh                     # Automated deployment
├── 🧪 quick-start.sh               # Testing environment
├── 📖 README.md                    # Comprehensive documentation
├── 📋 requirements.txt             # Python dependencies
├── 📊 docs/                        # Wiki documentation
│   ├── projects/                   # Individual project docs
│   ├── templates/                  # Documentation templates
│   └── global/                     # Cross-project knowledge
├── 🔧 scripts/                     # Automation scripts
│   ├── ai_integration.py          # Multi-model AI integration
│   ├── ai_brain.py                # Semantic search and analytics
│   ├── ai_wiki_cli.py             # Command-line interface
│   ├── git_hooks/                 # Git automation
│   └── tailscale_manager.sh       # Network management
├── aftrs_cli/                     # Network management system
│   ├── aftrs.sh                   # Main CLI dispatcher
│   ├── network_assets/            # Asset management
│   │   ├── assets.yaml           # Asset inventory
│   │   ├── asset_loader.py       # Asset query engine
│   │   ├── asset_merger.py       # Import/merge utilities
│   │   └── generate_docs_from_assets.py
│   ├── scripts/                   # Advanced modules
│   │   ├── tailscale_manager.sh  # Tailscale management
│   │   ├── tailscale_monitor.sh  # Real-time monitoring
│   │   ├── comcast_dmz_checker.sh # DMZ analysis
│   │   ├── nat_chain_visualizer.sh # NAT visualization
│   │   ├── router_tester.sh      # Router testing
│   │   ├── remote_site_tester.sh # Remote site testing
│   │   ├── speed_tester.sh       # Performance testing
│   │   ├── performance_optimizer.sh # Optimization engine
│   │   ├── dynamic_test_generator.sh # Asset-driven testing
│   │   ├── asset_import_scheduler.sh # Scheduled imports
│   │   └── asset_monitoring_pipeline.sh # Monitoring pipeline
│   ├── web_dashboard/             # Web interface
│   │   ├── dashboard.py          # Flask application
│   │   └── templates/            # HTML templates
│   ├── logs/                      # Log files
│   ├── docs/                      # Documentation
│   └── install.sh                 # Installation script
└── 📁 config/                      # Configuration files
    ├── supervisor.conf            # Process management
    └── nginx.conf                # Web server config
```

## 🚀 Deployment

### Production Deployment
```bash
# One-command deployment
./deploy.sh

# Access web interface
http://localhost:8080

# Database access
localhost:5432 (agentctl_wiki/agentctl_user)

# Network management dashboard
http://localhost:8080/dashboard
```

### Development Environment
```bash
# Test environment
./quick-start.sh start

# Validate setup
./validate-docker.sh

# Update system
./update.sh

# Setup network management
cd aftrs_cli
./install.sh
```

### Backup and Recovery
```bash
# Create backup
./backup.sh

# Restore from backup
docker exec -i agentctl_wiki psql -U agentctl_user agentctl_wiki < backup.sql

# Backup network assets
cd aftrs_cli
./network_assets/asset_merger.py backup
```

## 🔧 Configuration

### Environment Variables
```yaml
# Database
DB_HOST=postgres
DB_PORT=5432
DB_NAME=agentctl_wiki
DB_USER=agentctl_user
DB_PASSWORD=agentctl_password

# Application
SECRET_KEY=your-secret-key
FLASK_ENV=production

# AI Models
OPENAI_API_KEY=your-openai-key
ANTHROPIC_API_KEY=your-anthropic-key
OLLAMA_BASE_URL=http://localhost:11434

# Network Management
AFTRS_DASHBOARD_PORT=8080
AFTRS_DEBUG=false
AFTRS_LOG_LEVEL=INFO
```

### Network Management Configuration
```yaml
# Network Management Configuration
network_management:
  enabled: true
  assets_file: "aftrs_cli/network_assets/assets.yaml"
  dashboard_port: 8080
  monitoring_interval: 30
  alert_threshold: 100
  parallel_workers: 4
  cache_enabled: true
  cache_ttl: 300
```

### Port Configuration
- **Web Interface**: 8080 → 80 (container)
- **PostgreSQL**: 5432 → 5432 (container)
- **Network Dashboard**: 8080 → 8080 (container)
- **Monitoring API**: 8082 → 8082 (container)

## 📊 Features

### Web Dashboard
- **Real-time Statistics**: Projects, documents, decisions, progress logs
- **Search Functionality**: Cross-document search with highlighting
- **Project Management**: Interface for browsing and managing projects
- **Health Monitoring**: System health endpoints
- **Secure Login**: Admin authentication (default: admin/agentctl)
- **Network Dashboard**: Real-time asset monitoring and management

### CLI Integration
- **Unified Interface**: `./ai-wiki` command for all operations
- **AI Model Queries**: Direct access to OpenAI, Anthropic, Ollama
- **Documentation Generation**: AI-assisted content creation
- **Decision Logging**: Structured decision tracking
- **Progress Analytics**: Completion metrics and insights
- **Network Management**: `./aftrs.sh` for comprehensive network management

### AI Capabilities
- **Multi-model Support**: Automatic model selection
- **Semantic Search**: TF-IDF based search with ML libraries
- **Auto-cross-referencing**: Automatic linking between documents
- **Smart Commit Messages**: AI-generated meaningful commits
- **Template-driven**: Consistent documentation structure

### Network Management Capabilities
- **Asset-Driven Management**: Centralized asset inventory with relationships and metadata
- **Real-time Monitoring**: Live asset status and health checks
- **Advanced Diagnostics**: Comcast DMZ analysis and NAT chain visualization
- **Performance Optimization**: Parallel execution, intelligent caching, adaptive timeouts
- **Web Dashboard**: Real-time monitoring interface with custom widgets
- **Automation Pipeline**: Scheduled tasks and automated remediation

## 🔄 Development Workflow

### Creating New Projects
```bash
# Create project documentation
./ai-wiki new-project my-project

# Generate README with AI
./ai-wiki generate my-project README "Web application for task management"

# Create decision log
./ai-wiki decision my-project "Database Choice" "PostgreSQL vs MongoDB"
```

### Documentation Management
```bash
# Search across all docs
./ai-wiki search "authentication"

# Update documentation
./scripts/update_doc.sh docs/projects/my-project/README.md "Updated architecture"

# Query AI models
./ai-wiki query "What should I work on next?" --model openai
```

### Network Management
```bash
# Asset management
./aftrs.sh assets list all
./aftrs.sh assets list critical
./aftrs.sh assets query dashboard

# Network testing
./aftrs.sh tests generate router
./aftrs.sh tests run critical

# Real-time monitoring
./aftrs.sh monitor status
./aftrs.sh monitor topology

# Web dashboard
./aftrs.sh dashboard start
./aftrs.sh dashboard stop

# Documentation generation
./aftrs.sh docs table
./aftrs.sh docs diagram
```

### Git Integration
- **Automated Hooks**: Pre-commit and post-commit hooks
- **Meaningful Commits**: AI-generated commit messages
- **Progress Tracking**: Automated completion metrics
- **Cross-referencing**: Automatic linking between related docs

## 📈 Current Status

### ✅ Completed Features
- [x] Docker containerization with PostgreSQL
- [x] Flask web application with Bootstrap UI
- [x] Multi-model AI integration (OpenAI, Anthropic, Ollama)
- [x] Git-based version control with hooks
- [x] Structured markdown documentation
- [x] Semantic search functionality
- [x] Automated deployment scripts
- [x] Health monitoring and logging
- [x] Backup and recovery system
- [x] Asset-driven network management
- [x] Real-time network monitoring
- [x] Advanced diagnostic modules
- [x] Web dashboard with custom widgets
- [x] Performance optimization engine
- [x] Dynamic test generation
- [x] Automated documentation generation
- [x] Asset import/export capabilities
- [x] Monitoring pipeline with alerting
- [x] Remote site management
- [x] UNRAID integration

### 🔄 In Progress
- [ ] Advanced search analytics
- [ ] Progress dependency graphs
- [ ] AI model performance optimization
- [ ] Web interface enhancements
- [ ] Machine learning predictive analysis
- [ ] Advanced analytics and trend prediction

### 📋 Planned Features
- [ ] Real-time collaboration
- [ ] Advanced analytics dashboard
- [ ] API for external integrations
- [ ] Mobile-responsive design
- [ ] Advanced AI model fine-tuning
- [ ] Zero-touch provisioning
- [ ] Security analysis and threat detection
- [ ] Capacity planning and analysis
- [ ] Disaster recovery automation
- [ ] Multi-site geographic management

## 🛠️ Technical Decisions

### Container Architecture
- **Single Container**: All services in one container for simplicity
- **PostgreSQL External**: Database as separate service for scalability
- **Nginx Reverse Proxy**: Efficient web serving and load balancing
- **Supervisor Process Management**: Reliable service orchestration

### AI Integration Strategy
- **Progressive Enhancement**: System works without AI models
- **Auto-selection**: Intelligent model selection based on task
- **Graceful Degradation**: Fallback mechanisms for model failures
- **Unified Interface**: Same API for all AI models

### Documentation Design
- **Git as Backbone**: Leverages existing version control
- **Markdown with YAML**: Human-readable, AI-parseable
- **Template-driven**: Ensures consistency across projects
- **Metadata-rich**: Enables intelligent search and linking

### Network Management Design
- **Asset-First**: YAML-based asset inventory as the foundation
- **CLI-Centric**: Command-line automation with web dashboard
- **Configuration-Driven**: Asset-driven configuration and testing
- **Real-time Monitoring**: Live network health and performance tracking

## 🔍 Troubleshooting

### Common Issues

#### Database Connection
```bash
# Check PostgreSQL health
docker-compose exec postgres pg_isready -U agentctl_user

# View database logs
docker-compose logs postgres
```

#### Web Application
```bash
# Check application health
curl -f http://localhost:8080/health

# View application logs
docker-compose logs agentctl-wiki
```

#### AI Model Issues
```bash
# Test OpenAI connection
python3 -c "import openai; print('OpenAI available')"

# Test Ollama connection
curl http://localhost:11434/api/tags

# Check model availability
./ai-wiki list-models
```

#### Network Management Issues
```bash
# Check network management status
cd aftrs_cli
./aftrs.sh diag

# Test asset loader
python3 network_assets/asset_loader.py --help

# Check dashboard status
./aftrs.sh dashboard status

# Test monitoring pipeline
./scripts/asset_monitoring_pipeline.sh --test
```

### Performance Optimization
- **Database Indexing**: Automatic index creation for search
- **Caching**: Redis integration for frequently accessed data
- **Load Balancing**: Nginx configuration for high traffic
- **Resource Monitoring**: CPU and memory usage tracking
- **Network Optimization**: Parallel execution and intelligent caching
- **Asset-Driven Optimization**: Performance based on asset relationships

## 📚 Related Documentation

- [Docker Deployment Guide](../patterns/docker_deployment.md)
- [AI Integration Patterns](../patterns/ai_integration.md)
- [CLI Design Patterns](../patterns/cli_design.md)
- [Network Architecture](../infrastructure/network.md)
- [AFTRS CLI Integration](./aftrs_cli.md)
- [Asset Management Guide](../projects/asset_management.md)
- [Performance Optimization Guide](../projects/optimization_guide.md)

---

*Last Updated: 2025-01-15*
*Status: Production Ready with Network Management*
*Maintainer: AFTRS Development Team* 