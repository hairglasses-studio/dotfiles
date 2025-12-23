# OAI CLI - OpenAI Content Archiving System

## Overview
OAI CLI is a sophisticated archiving and backup system for OpenAI videos, images, and chat conversations. It provides comprehensive content management with web dashboard, database tracking, and automated workflows for preserving and organizing AI-generated content.

## 🏗️ Architecture

### Core Components
- **Multi-Content Archiving**: Images, videos, and chat conversations
- **PostgreSQL Database**: Structured storage with full-text search
- **Web Dashboard**: Real-time monitoring and content browsing
- **Docker Containerization**: Production-ready deployment
- **Automated Workflows**: Cron-based scheduling and batch processing
- **Browser Automation**: Chrome/Chromium integration for web scraping

### Key Features
- **Image Archiving**: Automated scraping with metadata extraction
- **Video Archiving**: Sora video backup with prompt extraction
- **Chat Archiving**: Complete ChatGPT conversation export
- **Search & Discovery**: Full-text search across all content types
- **Analytics & Monitoring**: Comprehensive statistics and reporting
- **Security & Reliability**: Non-root execution with structured logging

## 📁 Project Structure

```
oai_cli/
├── 📖 README.md                    # Comprehensive documentation
├── 🚀 install.sh                   # Automated installation script
├── 🔧 oai_cli.py                  # Main CLI application
├── 📋 requirements.txt             # Python dependencies
├── 🐳 Dockerfile                   # Container configuration
├── 📊 docker-compose.yml          # Multi-service deployment
├── 📁 web/                        # Web dashboard
│   └── app.py                     # Flask web application
├── 📁 scripts/                    # Automation scripts
├── 📁 tests/                      # Test suite
├── 📁 monitoring/                 # System monitoring
├── 📁 db/                         # Database schemas
├── 📁 backups/                    # Backup storage
├── 📁 downloads/                  # Content storage
│   ├── images/                    # Archived images
│   ├── videos/                    # Archived videos
│   └── chats/                     # Archived conversations
└── 📁 logs/                       # Application logs
```

## 🚀 Installation

### Prerequisites
- **Python 3.11+** with pip
- **PostgreSQL 12+** database
- **Chrome/Chromium** browser (for web scraping)
- **Docker** (optional, for containerized deployment)

### Automated Installation
```bash
# Clone repository
git clone https://github.com/hairglasses/oai_cli.git
cd oai_cli

# Run automated installation
chmod +x install.sh
./install.sh
```

### Manual Installation
```bash
# Create virtual environment
python3 -m venv venv
source venv/bin/activate

# Install dependencies
pip install -r requirements.txt

# Setup environment
cp .env.example .env
nano .env

# Create directories
mkdir -p downloads/{images,videos,chats} logs backups
```

### Docker Installation
```bash
# Using Docker Compose
docker-compose up -d

# Or build locally
docker build -t oai_cli .
docker run -d --name oai_cli oai_cli
```

## 🔧 Configuration

### Environment Setup
```bash
# Database Configuration
PG_CONN_STR=postgresql://oai_user:oai_password@localhost:5432/oai_archive

# Storage Configuration
DOWNLOAD_DIR=./downloads
IMAGES_DIR=./downloads/images
VIDEOS_DIR=./downloads/videos
CHATS_DIR=./downloads/chats

# Web Interface
WEB_HOST=0.0.0.0
WEB_PORT=8080
WEB_DEBUG=false

# Chrome Configuration
CHROME_HEADLESS=true
CHROME_WINDOW_SIZE=1920,1080
CHROME_DISABLE_GPU=true

# Archive Settings
MAX_RETRIES=3
DOWNLOAD_TIMEOUT=30
BATCH_SIZE=50
```

### Database Setup
```sql
-- Create database and user
CREATE DATABASE oai_archive;
CREATE USER oai_user WITH PASSWORD 'oai_password';
GRANT ALL PRIVILEGES ON DATABASE oai_archive TO oai_user;

-- Run schema
psql -d oai_archive -f sora_archive_schema.sql
```

## 📊 Features

### Core Commands

#### Image Archiving
```bash
# Archive images from OpenAI
oai_cli.py images --url <openai_url>

# Batch image processing
oai_cli.py images --batch --input images.txt

# Metadata extraction
oai_cli.py images --extract-metadata
```

#### Video Archiving
```bash
# Archive Sora videos
oai_cli.py videos --sora --url <sora_url>

# Video metadata extraction
oai_cli.py videos --extract-metadata

# Prompt text extraction
oai_cli.py videos --extract-prompts
```

#### Chat Archiving
```bash
# Archive ChatGPT conversations
oai_cli.py chats --url <chat_url>

# Export conversation history
oai_cli.py chats --export --format json

# Conversation summarization
oai_cli.py chats --summarize
```

#### Database Management
```bash
# Database backup
oai_cli.py db backup

# Database restore
oai_cli.py db restore backup.sql

# Database optimization
oai_cli.py db optimize
```

#### Search & Discovery
```bash
# Full-text search
oai_cli.py search "prompt text"

# Content type filtering
oai_cli.py search --type images "landscape"

# Metadata search
oai_cli.py search --metadata "resolution:4k"
```

#### System Management
```bash
# System health check
oai_cli.py doctor

# Performance monitoring
oai_cli.py monitor

# Cleanup old files
oai_cli.py cleanup
```

### Web Dashboard
```bash
# Start web interface
cd web && python3 app.py

# Access at http://localhost:8080
# Features:
# - Real-time monitoring
# - Advanced search interface
# - Content browsing
# - Statistics dashboard
```

## 🔄 Workflow

### Initial Setup
1. **Install**: `./install.sh`
2. **Configure Database**: Setup PostgreSQL
3. **Configure Environment**: Edit `.env` file
4. **Test Installation**: `oai_cli.py doctor`

### Daily Operations
1. **Archive Content**: `oai_cli.py images/videos/chats`
2. **Monitor Progress**: Web dashboard
3. **Search Content**: `oai_cli.py search`
4. **Backup Data**: `oai_cli.py db backup`

### Automated Workflows
1. **Cron Scheduling**: Automated archiving
2. **Batch Processing**: Large-scale content processing
3. **Health Monitoring**: System status tracking
4. **Cleanup Tasks**: Storage optimization

## 📈 Current Status

### ✅ Completed Features
- [x] Multi-content archiving (images, videos, chats)
- [x] PostgreSQL database integration
- [x] Web dashboard with real-time monitoring
- [x] Full-text search capabilities
- [x] Docker containerization
- [x] Automated installation script
- [x] Comprehensive test suite
- [x] CI/CD pipeline
- [x] Browser automation for web scraping
- [x] Metadata extraction and storage
- [x] Backup and restore functionality
- [x] Performance monitoring and logging

### 🔄 In Progress
- [ ] Advanced analytics dashboard
- [ ] Real-time content streaming
- [ ] Batch processing optimization
- [ ] Mobile-responsive web interface

### 📋 Planned Features
- [ ] Content deduplication
- [ ] Advanced search filters
- [ ] Content categorization
- [ ] Export functionality
- [ ] API for external integrations
- [ ] Multi-user support

## 🛠️ Technical Decisions

### Architecture Design
- **Modular Python**: Clean separation of concerns
- **Database-First**: PostgreSQL for structured data
- **Web Interface**: Flask for real-time monitoring
- **Container Ready**: Docker for deployment flexibility

### Content Processing
- **Multi-Format Support**: Images, videos, and text
- **Metadata Extraction**: Comprehensive content analysis
- **Browser Automation**: Chrome/Chromium for web scraping
- **Batch Processing**: Efficient large-scale operations

### Storage Strategy
- **Structured Database**: PostgreSQL with full-text search
- **File Organization**: Hierarchical storage structure
- **Backup System**: Automated backup and restore
- **Cleanup Automation**: Storage optimization

## 🔍 Troubleshooting

### Common Issues

#### Database Connection
```bash
# Check PostgreSQL status
sudo systemctl status postgresql

# Test database connection
psql -h localhost -U oai_user -d oai_archive

# Check connection string
echo $PG_CONN_STR
```

#### Chrome/Chromium Issues
```bash
# Check Chrome installation
google-chrome --version

# Install ChromeDriver
sudo apt-get install chromium-chromedriver

# Test browser automation
python3 -c "from selenium import webdriver; print('Chrome available')"
```

#### Content Download Issues
```bash
# Check download directory
ls -la downloads/

# Test network connectivity
curl -I https://openai.com

# Check permissions
ls -la downloads/images/
```

#### Web Dashboard Issues
```bash
# Check Flask installation
python3 -c "import flask"

# Start dashboard manually
cd web && python3 app.py

# Check logs
tail -f logs/oai_cli.log
```

### Debug Mode
```bash
# Enable debug logging
export OAI_DEBUG=1
oai_cli.py images --url <url>

# Verbose output
oai_cli.py doctor --verbose
```

## 📚 Related Documentation

- [Docker Deployment Guide](../patterns/docker_deployment.md)
- [Database Management](../patterns/database_management.md)
- [Web Scraping Patterns](../patterns/web_scraping.md)
- [Content Management Workflows](../patterns/content_management.md)

## 🔗 Integration Points

### With Other Systems
- **AgentCTL**: Content metadata integration
- **CR8 CLI**: Media processing workflows
- **UNRAID Scripts**: Storage management
- **AFTRS CLI**: Network storage integration

### External Services
- **OpenAI API**: Content access and metadata
- **PostgreSQL**: Database storage and queries
- **Chrome/Chromium**: Web scraping automation
- **Docker Hub**: Container registry

---

*Last Updated: 2025-01-14*
*Status: Production Ready*
*Maintainer: AFTRS Development Team* 