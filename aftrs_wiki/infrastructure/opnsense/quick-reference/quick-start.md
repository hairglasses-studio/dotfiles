# OPNsense Quick Start Guide

## Common Tasks

### Router Configuration
1. Review [../opnsense-wiki/configurations/](../opnsense-wiki/configurations/) for backup/restore
2. Check XML configurations for router setup
3. Test configurations before deployment

### LLM Agent Integration
1. Set up [opnsense-llmagent](../opnsense-wiki/llm-agents/opnsense-llmagent.md)
2. Configure AI-driven network management
3. Test LLM agent connectivity

### Backup and Recovery
1. Configure automated backups from [../opnsense-wiki/configurations/](../opnsense-wiki/configurations/)
2. Set up multiple backup destinations
3. Test restore procedures regularly

### Network Integration
1. Coordinate with AFTRS CLI for network automation
2. Integrate with Tailscale for VPN access
3. Connect with UNRAID for server coordination

## Repository Quick Reference

- **aftrs_opnsense_config_backup:** [opnsense-wiki/configurations/aftrs_opnsense_config_backup.md](../opnsense-wiki/configurations/aftrs_opnsense_config_backup.md)
- **opnsense-llmagent:** [opnsense-wiki/llm-agents/opnsense-llmagent.md](../opnsense-wiki/llm-agents/opnsense-llmagent.md)
- **secretstudios_opnsense_router_backup:** [opnsense-wiki/configurations/secretstudios_opnsense_router_backup.md](../opnsense-wiki/configurations/secretstudios_opnsense_router_backup.md)

## Helper Scripts

### Interoperability Scripts
- `./helper-scripts/backup_all_configs.sh` - Coordinate all OPNsense backups
- `./helper-scripts/test_network_integration.sh` - Test network integration
- `./helper-scripts/coordinate_apis.sh` - API coordination across tools

### Centralization Tools  
- `./centralization-tools/setup_monitoring.sh` - Set up centralized monitoring
- `./centralization-tools/sync_configs.sh` - Synchronize configurations
- `./centralization-tools/coordinate_topology.sh` - Network topology coordination

## Troubleshooting

### Common Issues
1. **Configuration Conflicts:** Check XML config validation
2. **API Access Issues:** Verify OPNsense API credentials
3. **Network Connectivity:** Test routing and firewall rules
4. **Backup Failures:** Check storage space and permissions

### Integration Testing
```bash
# Test all integrations
./helper-scripts/test_network_integration.sh

# Test API coordination
./helper-scripts/coordinate_apis.sh

# Test backup coordination
./helper-scripts/backup_all_configs.sh
```

### Getting Help
1. Check individual repository documentation
2. Review [AFTRS Wiki](../../aftrs_wiki/) for ecosystem context
3. Use [AgentCTL](../../agentctl/) for AI assistance
4. Check [UNRAID Monolith](../../unraid-monolith/) for server integration

---
*Last updated: 2025-09-23 00:34:08*
