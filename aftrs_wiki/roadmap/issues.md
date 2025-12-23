# Known Issues - AFTRS Ecosystem

## Overview
This document tracks known issues, problems, and workarounds across all AFTRS projects and infrastructure. It serves as a central reference for troubleshooting and planning fixes.

## 🚨 Critical Issues

### 1. **OPNSense Router Optimization**
**Status**: 🔄 In Progress  
**Priority**: High  
**Affects**: Network performance, security

#### Issues
- **Comcast Modem Configuration**: Not fully optimized for bridge mode
- **Firewall Rules**: Some rules may be overly permissive
- **DNS Performance**: Unbound DNS could be optimized for local queries
- **QoS Configuration**: Quality of Service not fully configured

#### Workarounds
```bash
# Temporary DNS fix
echo "nameserver 1.1.1.1" >> /etc/resolv.conf
echo "nameserver 8.8.8.8" >> /etc/resolv.conf

# Check firewall rules
pfctl -sr | grep -v "block"

# Monitor DNS performance
dig @192.168.1.1 google.com
```

#### Planned Fixes
- [ ] Optimize Comcast modem bridge mode configuration
- [ ] Review and tighten firewall rules
- [ ] Configure DNS caching and optimization
- [ ] Implement QoS for critical traffic
- [ ] Add comprehensive network monitoring

### 2. **UNRAID Server Performance**
**Status**: 🔄 In Progress  
**Priority**: High  
**Affects**: Storage performance, media processing

#### Issues
- **NVMe Storage**: Not fully utilized for caching
- **Docker Performance**: Some containers may be resource-constrained
- **Backup Performance**: Automated backups could be optimized
- **Media Processing**: FFmpeg operations could be parallelized

#### Workarounds
```bash
# Check NVMe usage
df -h /mnt/cache

# Monitor Docker performance
docker stats

# Optimize backup schedule
crontab -l | grep backup
```

#### Planned Fixes
- [ ] Optimize NVMe cache configuration
- [ ] Implement Docker resource limits
- [ ] Parallelize backup operations
- [ ] Add media processing queue
- [ ] Implement storage tiering

## ⚠️ Medium Priority Issues

### 3. **AgentCTL Wiki System**
**Status**: 🔄 In Progress  
**Priority**: Medium  
**Affects**: Documentation, AI context

#### Issues
- **Search Functionality**: Semantic search needs improvement
- **AI Integration**: Model selection could be more intelligent
- **Web Interface**: Dashboard needs mobile responsiveness
- **Performance**: Large documentation sets may be slow

#### Workarounds
```bash
# Use basic search
./ai-wiki search "term" --basic

# Force specific AI model
./ai-wiki query "question" --model ollama

# Clear cache
rm -rf .cache/agentctl
```

#### Planned Fixes
- [ ] Implement advanced search algorithms
- [ ] Add intelligent model selection
- [ ] Improve web interface responsiveness
- [ ] Optimize database queries
- [ ] Add caching layer

### 4. **CR8 CLI Media Processing**
**Status**: 🔄 In Progress  
**Priority**: Medium  
**Affects**: Media organization, DJ workflows

#### Issues
- **Batch Processing**: Large media sets may timeout
- **Metadata Extraction**: Some files may have incomplete metadata
- **Rekordbox Integration**: XML export format may need updates
- **Authentication**: YouTube/SoundCloud tokens may expire

#### Workarounds
```bash
# Process in smaller batches
cr8 youtube <url> --batch-size 10

# Force metadata extraction
cr8 metadata --force <file>

# Re-authenticate
cr8 auth youtube
cr8 auth soundcloud
```

#### Planned Fixes
- [ ] Implement robust batch processing
- [ ] Improve metadata extraction accuracy
- [ ] Update Rekordbox XML format
- [ ] Add token refresh automation
- [ ] Implement retry mechanisms

### 5. **OAI CLI Content Archiving**
**Status**: 🔄 In Progress  
**Priority**: Medium  
**Affects**: Content backup, OpenAI integration

#### Issues
- **Sora Access**: Manual login required for Sora videos
- **Rate Limiting**: API calls may be rate-limited
- **Storage Management**: Large video files may fill storage
- **Browser Automation**: Chrome/Chromium may be unstable

#### Workarounds
```bash
# Manual Sora login
oai_cli.py sora --manual-login

# Check storage space
df -h downloads/videos

# Restart browser automation
pkill chrome
oai_cli.py videos --restart-browser
```

#### Planned Fixes
- [ ] Implement automated Sora authentication
- [ ] Add rate limiting and backoff
- [ ] Implement storage management
- [ ] Improve browser automation stability
- [ ] Add content deduplication

## 🔧 Low Priority Issues

### 6. **TSCTL Network Management**
**Status**: 📋 Planned  
**Priority**: Low  
**Affects**: Tailscale network management

#### Issues
- **Tag Management**: Tags may not sync properly
- **Configuration**: Some hosts may have inconsistent configs
- **Monitoring**: Limited network analytics
- **Automation**: Some tasks require manual intervention

#### Workarounds
```bash
# Re-sync tags
tsctl check --force

# Reconfigure host
tsctl setup --force

# Check network status
tsctl status --verbose
```

#### Planned Fixes
- [ ] Implement robust tag synchronization
- [ ] Add configuration validation
- [ ] Improve network monitoring
- [ ] Add automation for common tasks

### 7. **AFTRS CLI Network Management**
**Status**: 📋 Planned  
**Priority**: Low  
**Affects**: Network infrastructure management

#### Issues
- **Bootstrap Process**: May fail on some OPNsense versions
- **DNS Management**: Some overrides may not persist
- **DHCP Leases**: Static leases may not be reliable
- **Firewall Rules**: Rule management could be improved

#### Workarounds
```bash
# Manual bootstrap
aftrs_cli bootstrap --manual

# Re-apply DNS overrides
aftrs_cli dns apply

# Check DHCP leases
aftrs_cli dhcp list --verbose
```

#### Planned Fixes
- [ ] Improve bootstrap reliability
- [ ] Add DNS override persistence
- [ ] Implement reliable DHCP management
- [ ] Add firewall rule validation

### 8. **UNRAID Scripts Automation**
**Status**: 📋 Planned  
**Priority**: Low  
**Affects**: Server automation, maintenance

#### Issues
- **Cron Jobs**: Some jobs may not run reliably
- **Error Handling**: Limited error recovery
- **Logging**: Log rotation may not work properly
- **Performance**: Some scripts may be inefficient

#### Workarounds
```bash
# Check cron jobs
crontab -l

# Manual script execution
./maintenance_scripts/run_sora_backup.sh

# Check logs
tail -f logs/unraid_scripts.log
```

#### Planned Fixes
- [ ] Improve cron job reliability
- [ ] Add comprehensive error handling
- [ ] Implement proper log rotation
- [ ] Optimize script performance

## 🔍 Infrastructure Issues

### 9. **Network Performance**
**Status**: 🔄 In Progress  
**Priority**: Medium  
**Affects**: Overall network performance

#### Issues
- **Bandwidth**: 1Gbps may be limiting for large transfers
- **Latency**: Some network paths may have high latency
- **WiFi Performance**: 5GHz optimization needed
- **VLAN Configuration**: Network segmentation could be improved

#### Workarounds
```bash
# Check network performance
speedtest-cli

# Monitor latency
ping -c 10 8.8.8.8

# Check WiFi channels
iwlist wlan0 channel
```

#### Planned Fixes
- [ ] Upgrade to 2.5Gbps network
- [ ] Optimize WiFi channel selection
- [ ] Implement proper VLAN segmentation
- [ ] Add network performance monitoring

### 10. **Security Vulnerabilities**
**Status**: 🔄 In Progress  
**Priority**: High  
**Affects**: Overall system security

#### Issues
- **Default Passwords**: Some services may have default credentials
- **SSL Certificates**: Some certificates may be self-signed
- **Firewall Rules**: Some rules may be too permissive
- **Access Control**: Some services may lack proper authentication

#### Workarounds
```bash
# Change default passwords
passwd
# Change service passwords

# Check SSL certificates
openssl s_client -connect host:port

# Review firewall rules
iptables -L -n
```

#### Planned Fixes
- [ ] Implement password rotation
- [ ] Add proper SSL certificates
- [ ] Tighten firewall rules
- [ ] Implement proper access control

## 🛠️ Development Issues

### 11. **Code Quality**
**Status**: 📋 Planned  
**Priority**: Low  
**Affects**: Code maintainability

#### Issues
- **Documentation**: Some functions lack proper documentation
- **Error Handling**: Some error cases not handled
- **Testing**: Limited test coverage
- **Code Style**: Inconsistent coding style

#### Workarounds
```bash
# Run linting
flake8 *.py
shellcheck *.sh

# Run tests
python -m pytest tests/
```

#### Planned Fixes
- [ ] Add comprehensive documentation
- [ ] Implement proper error handling
- [ ] Increase test coverage
- [ ] Standardize code style

### 12. **Deployment Issues**
**Status**: 📋 Planned  
**Priority**: Low  
**Affects**: System deployment

#### Issues
- **Docker Images**: Some images may be outdated
- **Dependencies**: Some dependencies may have security issues
- **Configuration**: Some configs may be hardcoded
- **Backup**: Backup procedures may not be comprehensive

#### Workarounds
```bash
# Update Docker images
docker pull image:latest

# Update dependencies
pip install --upgrade -r requirements.txt

# Check configuration
grep -r "hardcoded" config/
```

#### Planned Fixes
- [ ] Implement automated image updates
- [ ] Add dependency vulnerability scanning
- [ ] Externalize all configurations
- [ ] Implement comprehensive backup

## 📊 Issue Tracking

### Status Legend
- 🚨 **Critical**: System-breaking issues
- ⚠️ **High**: Major functionality affected
- 🔧 **Medium**: Minor functionality affected
- 📋 **Low**: Cosmetic or minor issues
- ✅ **Fixed**: Issues that have been resolved
- 🔄 **In Progress**: Currently being worked on
- 📋 **Planned**: Scheduled for future work

### Priority Matrix
| Priority | Impact | Effort | Timeline |
|----------|--------|--------|----------|
| Critical | High | High | Immediate |
| High | High | Medium | 1-2 weeks |
| Medium | Medium | Medium | 1-2 months |
| Low | Low | Low | 3-6 months |

## 🔄 Issue Resolution Process

### 1. **Issue Identification**
- Monitor system logs and alerts
- User reports and feedback
- Automated health checks
- Performance monitoring

### 2. **Issue Assessment**
- Determine impact and scope
- Assess priority and urgency
- Estimate effort required
- Plan resolution timeline

### 3. **Issue Resolution**
- Implement fixes and workarounds
- Test solutions thoroughly
- Deploy to production
- Monitor for regression

### 4. **Issue Closure**
- Verify fix is working
- Update documentation
- Close issue ticket
- Archive resolution details

## 📚 Related Documentation

- [Network Architecture](../infrastructure/network.md)
- [Troubleshooting Guides](../patterns/troubleshooting.md)
- [Development Workflow](../patterns/development_workflow.md)
- [System Monitoring](../patterns/monitoring.md)

---

*Last Updated: 2025-01-14*
*Status: Active*
*Maintainer: AFTRS Development Team* 