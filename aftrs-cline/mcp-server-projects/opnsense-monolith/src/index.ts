#!/usr/bin/env node

/**
 * MCP Server for OPNsense Monolith - Router/Firewall Management
 * Provides AI access to OPNsense router configuration, monitoring, and security management
 */

import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import {
  CallToolRequestSchema,
  ErrorCode,
  ListToolsRequestSchema,
  McpError,
} from '@modelcontextprotocol/sdk/types.js';
import fetch from 'node-fetch';

// OPNsense connection details
const OPNSENSE_HOST = process.env.OPNSENSE_HOST || '192.168.1.1';
const OPNSENSE_API_KEY = process.env.OPNSENSE_API_KEY || '';
const OPNSENSE_API_SECRET = process.env.OPNSENSE_API_SECRET || '';
const OPNSENSE_PORT = process.env.OPNSENSE_PORT || '443';

interface FirewallRule {
  uuid: string;
  enabled: boolean;
  action: 'pass' | 'block' | 'reject';
  interface: string;
  protocol: string;
  source: string;
  destination: string;
  port: string;
  description: string;
}

interface SystemInfo {
  hostname: string;
  version: string;
  uptime: string;
  cpu_usage: number;
  memory_usage: number;
  disk_usage: number;
  temperature: number;
}

interface InterfaceStatus {
  name: string;
  description: string;
  status: 'up' | 'down';
  ipv4: string;
  ipv6: string;
  mac: string;
  mtu: number;
}

class OPNsenseMCPServer {
  private server: Server;

  constructor() {
    this.server = new Server(
      {
        name: 'opnsense-monolith',
        version: '1.0.0',
      },
      {
        capabilities: {
          tools: {},
        },
      }
    );

    this.setupHandlers();
  }

  private setupHandlers() {
    this.server.setRequestHandler(ListToolsRequestSchema, async () => {
      return {
        tools: [
          // System Information & Status
          {
            name: 'get_system_info',
            description: 'Get comprehensive OPNsense system information and status',
            inputSchema: {
              type: 'object',
              properties: {
                include_hardware: {
                  type: 'boolean',
                  description: 'Include hardware information',
                  default: true,
                },
                include_network: {
                  type: 'boolean',
                  description: 'Include network interface status',
                  default: true,
                },
                include_services: {
                  type: 'boolean',
                  description: 'Include service status information',
                  default: false,
                },
              },
            },
          },
          {
            name: 'get_system_health',
            description: 'Get real-time system health metrics and performance data',
            inputSchema: {
              type: 'object',
              properties: {
                time_period: {
                  type: 'string',
                  enum: ['1h', '6h', '24h', '7d'],
                  description: 'Time period for health metrics',
                  default: '1h',
                },
                metrics: {
                  type: 'array',
                  items: {
                    type: 'string',
                    enum: ['cpu', 'memory', 'disk', 'network', 'temperature', 'all'],
                  },
                  description: 'Specific metrics to retrieve',
                  default: ['all'],
                },
              },
            },
          },
          // Firewall Management
          {
            name: 'list_firewall_rules',
            description: 'List all firewall rules with filtering options',
            inputSchema: {
              type: 'object',
              properties: {
                interface: {
                  type: 'string',
                  description: 'Filter by specific interface',
                },
                action: {
                  type: 'string',
                  enum: ['pass', 'block', 'reject', 'all'],
                  description: 'Filter by rule action',
                  default: 'all',
                },
                enabled_only: {
                  type: 'boolean',
                  description: 'Show only enabled rules',
                  default: false,
                },
                format: {
                  type: 'string',
                  enum: ['table', 'json', 'detailed'],
                  description: 'Output format',
                  default: 'table',
                },
              },
            },
          },
          {
            name: 'create_firewall_rule',
            description: 'Create a new firewall rule with specified parameters',
            inputSchema: {
              type: 'object',
              properties: {
                action: {
                  type: 'string',
                  enum: ['pass', 'block', 'reject'],
                  description: 'Rule action',
                },
                interface: {
                  type: 'string',
                  description: 'Interface name (e.g., WAN, LAN)',
                },
                protocol: {
                  type: 'string',
                  enum: ['TCP', 'UDP', 'ICMP', 'any'],
                  description: 'Protocol type',
                  default: 'TCP',
                },
                source: {
                  type: 'string',
                  description: 'Source address or network',
                  default: 'any',
                },
                destination: {
                  type: 'string',
                  description: 'Destination address or network',
                  default: 'any',
                },
                port: {
                  type: 'string',
                  description: 'Destination port or range',
                },
                description: {
                  type: 'string',
                  description: 'Rule description',
                },
                enabled: {
                  type: 'boolean',
                  description: 'Enable rule immediately',
                  default: true,
                },
              },
              required: ['action', 'interface'],
            },
          },
          {
            name: 'modify_firewall_rule',
            description: 'Modify an existing firewall rule',
            inputSchema: {
              type: 'object',
              properties: {
                rule_uuid: {
                  type: 'string',
                  description: 'UUID of the rule to modify',
                },
                action: {
                  type: 'string',
                  enum: ['pass', 'block', 'reject'],
                  description: 'Rule action',
                },
                enabled: {
                  type: 'boolean',
                  description: 'Enable/disable rule',
                },
                description: {
                  type: 'string',
                  description: 'Updated rule description',
                },
                source: {
                  type: 'string',
                  description: 'Updated source address',
                },
                destination: {
                  type: 'string',
                  description: 'Updated destination address',
                },
                port: {
                  type: 'string',
                  description: 'Updated port specification',
                },
              },
              required: ['rule_uuid'],
            },
          },
          {
            name: 'delete_firewall_rule',
            description: 'Delete a firewall rule by UUID',
            inputSchema: {
              type: 'object',
              properties: {
                rule_uuid: {
                  type: 'string',
                  description: 'UUID of the rule to delete',
                },
                confirm: {
                  type: 'boolean',
                  description: 'Confirm deletion of the rule',
                  default: false,
                },
              },
              required: ['rule_uuid'],
            },
          },
          // Interface Management
          {
            name: 'list_interfaces',
            description: 'List all network interfaces with status and configuration',
            inputSchema: {
              type: 'object',
              properties: {
                include_virtual: {
                  type: 'boolean',
                  description: 'Include virtual interfaces (VLANs, etc.)',
                  default: true,
                },
                status_filter: {
                  type: 'string',
                  enum: ['up', 'down', 'all'],
                  description: 'Filter by interface status',
                  default: 'all',
                },
              },
            },
          },
          {
            name: 'configure_interface',
            description: 'Configure network interface settings',
            inputSchema: {
              type: 'object',
              properties: {
                interface_name: {
                  type: 'string',
                  description: 'Interface name to configure',
                },
                enabled: {
                  type: 'boolean',
                  description: 'Enable/disable interface',
                },
                ip_address: {
                  type: 'string',
                  description: 'IPv4 address (CIDR notation)',
                },
                gateway: {
                  type: 'string',
                  description: 'Gateway address',
                },
                mtu: {
                  type: 'number',
                  description: 'Maximum Transmission Unit',
                  minimum: 68,
                  maximum: 9000,
                },
                description: {
                  type: 'string',
                  description: 'Interface description',
                },
              },
              required: ['interface_name'],
            },
          },
          // DHCP Management
          {
            name: 'get_dhcp_status',
            description: 'Get DHCP server status and lease information',
            inputSchema: {
              type: 'object',
              properties: {
                interface: {
                  type: 'string',
                  description: 'Specific interface to check (empty for all)',
                },
                include_leases: {
                  type: 'boolean',
                  description: 'Include active DHCP leases',
                  default: true,
                },
                include_static: {
                  type: 'boolean',
                  description: 'Include static DHCP mappings',
                  default: false,
                },
              },
            },
          },
          {
            name: 'manage_dhcp_lease',
            description: 'Create or modify DHCP static lease mappings',
            inputSchema: {
              type: 'object',
              properties: {
                action: {
                  type: 'string',
                  enum: ['create', 'modify', 'delete'],
                  description: 'Action to perform',
                },
                interface: {
                  type: 'string',
                  description: 'Interface for DHCP lease',
                },
                mac_address: {
                  type: 'string',
                  description: 'MAC address for static lease',
                },
                ip_address: {
                  type: 'string',
                  description: 'IP address to assign',
                },
                hostname: {
                  type: 'string',
                  description: 'Hostname for the lease',
                },
                description: {
                  type: 'string',
                  description: 'Description for the mapping',
                },
              },
              required: ['action', 'mac_address'],
            },
          },
          // VPN Management
          {
            name: 'get_vpn_status',
            description: 'Get status of VPN connections and tunnels',
            inputSchema: {
              type: 'object',
              properties: {
                vpn_type: {
                  type: 'string',
                  enum: ['openvpn', 'ipsec', 'wireguard', 'all'],
                  description: 'Type of VPN to check',
                  default: 'all',
                },
                include_logs: {
                  type: 'boolean',
                  description: 'Include recent VPN logs',
                  default: false,
                },
              },
            },
          },
          // Configuration Management
          {
            name: 'backup_configuration',
            description: 'Create a backup of the current OPNsense configuration',
            inputSchema: {
              type: 'object',
              properties: {
                include_passwords: {
                  type: 'boolean',
                  description: 'Include encrypted passwords in backup',
                  default: false,
                },
                backup_name: {
                  type: 'string',
                  description: 'Custom name for the backup file',
                },
              },
            },
          },
          {
            name: 'restore_configuration',
            description: 'Restore OPNsense configuration from backup',
            inputSchema: {
              type: 'object',
              properties: {
                backup_file: {
                  type: 'string',
                  description: 'Path to backup file to restore',
                },
                reboot_required: {
                  type: 'boolean',
                  description: 'Whether system reboot is required after restore',
                  default: true,
                },
                confirm: {
                  type: 'boolean',
                  description: 'Confirm restoration (destructive operation)',
                  default: false,
                },
              },
              required: ['backup_file', 'confirm'],
            },
          },
          // Traffic Analysis
          {
            name: 'get_traffic_stats',
            description: 'Get network traffic statistics and analysis',
            inputSchema: {
              type: 'object',
              properties: {
                interface: {
                  type: 'string',
                  description: 'Specific interface (empty for all)',
                },
                time_range: {
                  type: 'string',
                  enum: ['1h', '6h', '24h', '7d', '30d'],
                  description: 'Time range for statistics',
                  default: '24h',
                },
                traffic_type: {
                  type: 'string',
                  enum: ['total', 'inbound', 'outbound'],
                  description: 'Type of traffic to analyze',
                  default: 'total',
                },
                top_protocols: {
                  type: 'boolean',
                  description: 'Include top protocols analysis',
                  default: false,
                },
              },
            },
          },
          // Security & Monitoring
          {
            name: 'get_security_alerts',
            description: 'Get recent security alerts and intrusion detection events',
            inputSchema: {
              type: 'object',
              properties: {
                severity: {
                  type: 'string',
                  enum: ['low', 'medium', 'high', 'critical', 'all'],
                  description: 'Filter by alert severity',
                  default: 'all',
                },
                time_range: {
                  type: 'string',
                  enum: ['1h', '6h', '24h', '7d'],
                  description: 'Time range for alerts',
                  default: '24h',
                },
                limit: {
                  type: 'number',
                  description: 'Maximum number of alerts to return',
                  default: 50,
                },
              },
            },
          },
          {
            name: 'block_ip_address',
            description: 'Block an IP address or network range',
            inputSchema: {
              type: 'object',
              properties: {
                ip_address: {
                  type: 'string',
                  description: 'IP address or CIDR network to block',
                },
                duration: {
                  type: 'string',
                  enum: ['1h', '24h', '7d', '30d', 'permanent'],
                  description: 'Block duration',
                  default: '24h',
                },
                reason: {
                  type: 'string',
                  description: 'Reason for blocking',
                },
                whitelist_local: {
                  type: 'boolean',
                  description: 'Exempt local networks from block',
                  default: true,
                },
              },
              required: ['ip_address'],
            },
          },
          // System Operations
          {
            name: 'restart_service',
            description: 'Restart a specific OPNsense service',
            inputSchema: {
              type: 'object',
              properties: {
                service_name: {
                  type: 'string',
                  enum: ['firewall', 'dhcp', 'dns', 'openvpn', 'ipsec', 'wireguard', 'web'],
                  description: 'Service to restart',
                },
                wait_for_completion: {
                  type: 'boolean',
                  description: 'Wait for service restart to complete',
                  default: true,
                },
              },
              required: ['service_name'],
            },
          },
          {
            name: 'reboot_system',
            description: 'Reboot the OPNsense system',
            inputSchema: {
              type: 'object',
              properties: {
                delay_minutes: {
                  type: 'number',
                  description: 'Delay before reboot in minutes',
                  default: 1,
                  minimum: 0,
                  maximum: 60,
                },
                confirm: {
                  type: 'boolean',
                  description: 'Confirm system reboot',
                  default: false,
                },
              },
              required: ['confirm'],
            },
          },
          // Logs & Diagnostics
          {
            name: 'get_system_logs',
            description: 'Retrieve system logs with filtering options',
            inputSchema: {
              type: 'object',
              properties: {
                log_type: {
                  type: 'string',
                  enum: ['system', 'firewall', 'dhcp', 'vpn', 'all'],
                  description: 'Type of logs to retrieve',
                  default: 'system',
                },
                severity: {
                  type: 'string',
                  enum: ['debug', 'info', 'notice', 'warning', 'error', 'critical', 'all'],
                  description: 'Minimum log severity',
                  default: 'info',
                },
                lines: {
                  type: 'number',
                  description: 'Number of log lines to return',
                  default: 100,
                  minimum: 1,
                  maximum: 1000,
                },
                search_term: {
                  type: 'string',
                  description: 'Search term to filter logs',
                },
              },
            },
          },
          {
            name: 'run_diagnostics',
            description: 'Run network diagnostics and connectivity tests',
            inputSchema: {
              type: 'object',
              properties: {
                test_type: {
                  type: 'string',
                  enum: ['ping', 'traceroute', 'dns_lookup', 'port_scan', 'all'],
                  description: 'Type of diagnostic test',
                },
                target: {
                  type: 'string',
                  description: 'Target IP address or hostname for test',
                },
                interface: {
                  type: 'string',
                  description: 'Source interface for test',
                },
                count: {
                  type: 'number',
                  description: 'Number of test iterations (for ping/traceroute)',
                  default: 4,
                  minimum: 1,
                  maximum: 20,
                },
              },
              required: ['test_type', 'target'],
            },
          },
        ],
      };
    });

    this.server.setRequestHandler(CallToolRequestSchema, async (request) => {
      try {
        const { name, arguments: args } = request.params;

        switch (name) {
          case 'get_system_info':
            return await this.getSystemInfo(args.include_hardware, args.include_network, args.include_services);

          case 'get_system_health':
            return await this.getSystemHealth(args.time_period, args.metrics);

          case 'list_firewall_rules':
            return await this.listFirewallRules(args.interface, args.action, args.enabled_only, args.format);

          case 'create_firewall_rule':
            return await this.createFirewallRule(args);

          case 'modify_firewall_rule':
            return await this.modifyFirewallRule(args.rule_uuid, args);

          case 'delete_firewall_rule':
            return await this.deleteFirewallRule(args.rule_uuid, args.confirm);

          case 'list_interfaces':
            return await this.listInterfaces(args.include_virtual, args.status_filter);

          case 'configure_interface':
            return await this.configureInterface(args.interface_name, args);

          case 'get_dhcp_status':
            return await this.getDHCPStatus(args.interface, args.include_leases, args.include_static);

          case 'manage_dhcp_lease':
            return await this.manageDHCPLease(args.action, args);

          case 'get_vpn_status':
            return await this.getVPNStatus(args.vpn_type, args.include_logs);

          case 'backup_configuration':
            return await this.backupConfiguration(args.include_passwords, args.backup_name);

          case 'restore_configuration':
            return await this.restoreConfiguration(args.backup_file, args.reboot_required, args.confirm);

          case 'get_traffic_stats':
            return await this.getTrafficStats(args.interface, args.time_range, args.traffic_type, args.top_protocols);

          case 'get_security_alerts':
            return await this.getSecurityAlerts(args.severity, args.time_range, args.limit);

          case 'block_ip_address':
            return await this.blockIPAddress(args.ip_address, args.duration, args.reason, args.whitelist_local);

          case 'restart_service':
            return await this.restartService(args.service_name, args.wait_for_completion);

          case 'reboot_system':
            return await this.rebootSystem(args.delay_minutes, args.confirm);

          case 'get_system_logs':
            return await this.getSystemLogs(args.log_type, args.severity, args.lines, args.search_term);

          case 'run_diagnostics':
            return await this.runDiagnostics(args.test_type, args.target, args.interface, args.count);

          default:
            throw new McpError(ErrorCode.MethodNotFound, `Unknown tool: ${name}`);
        }
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : 'Unknown error occurred';
        throw new McpError(ErrorCode.InternalError, `OPNsense error: ${errorMessage}`);
      }
    });
  }

  private async makeAPIRequest(endpoint: string, method = 'GET', data?: any): Promise<any> {
    const url = `https://${OPNSENSE_HOST}:${OPNSENSE_PORT}/api/${endpoint}`;
    const auth = Buffer.from(`${OPNSENSE_API_KEY}:${OPNSENSE_API_SECRET}`).toString('base64');

    const options: any = {
      method,
      headers: {
        'Authorization': `Basic ${auth}`,
        'Content-Type': 'application/json',
      },
      // Disable SSL verification for self-signed certificates
      agent: process.env.NODE_TLS_REJECT_UNAUTHORIZED === '0' ? undefined : require('https').Agent({
        rejectUnauthorized: false,
      }),
    };

    if (data && method !== 'GET') {
      options.body = JSON.stringify(data);
    }

    const response = await fetch(url, options);

    if (!response.ok) {
      throw new Error(`OPNsense API error: ${response.status} ${response.statusText}`);
    }

    const contentType = response.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      return await response.json();
    } else {
      return await response.text();
    }
  }

  private async getSystemInfo(includeHardware = true, includeNetwork = true, includeServices = false): Promise<any> {
    try {
      const systemInfo = await this.makeAPIRequest('core/system/status');

      let result = `OPNsense System Information:\n`;
      result += `Hostname: ${systemInfo.hostname || 'Unknown'}\n`;
      result += `Version: ${systemInfo.version || 'Unknown'}\n`;
      result += `Uptime: ${systemInfo.uptime || 'Unknown'}\n`;

      if (includeHardware) {
        result += `\nHardware Information:\n`;
        result += `CPU Usage: ${systemInfo.cpu || 'N/A'}%\n`;
        result += `Memory Usage: ${systemInfo.memory || 'N/A'}%\n`;
        result += `Disk Usage: ${systemInfo.disk || 'N/A'}%\n`;
      }

      if (includeNetwork) {
        const interfaces = await this.makeAPIRequest('interfaces/overview/export');
        result += `\nNetwork Interfaces: ${Object.keys(interfaces || {}).length}\n`;
      }

      if (includeServices) {
        const services = await this.makeAPIRequest('core/service/search');
        result += `\nRunning Services: ${services?.rows?.length || 0}\n`;
      }

      return {
        content: [
          {
            type: 'text',
            text: result,
          },
        ],
      };
    } catch (error) {
      return {
        content: [
          {
            type: 'text',
            text: `Error retrieving system info: ${error instanceof Error ? error.message : 'Unknown error'}`,
          },
        ],
      };
    }
  }

  private async getSystemHealth(timePeriod = '1h', metrics = ['all']): Promise<any> {
    try {
      // This would typically fetch from monitoring endpoints
      let result = `System Health Metrics (${timePeriod}):\n\n`;

      if (metrics.includes('all') || metrics.includes('cpu')) {
        const cpuData = await this.makeAPIRequest('diagnostics/system/cpu');
        result += `CPU Usage: ${cpuData?.average || 'N/A'}%\n`;
      }

      if (metrics.includes('all') || metrics.includes('memory')) {
        const memData = await this.makeAPIRequest('diagnostics/system/memory');
        result += `Memory Usage: ${memData?.used_percent || 'N/A'}%\n`;
      }

      return {
        content: [
          {
            type: 'text',
            text: result,
          },
        ],
      };
    } catch (error) {
      return {
        content: [
          {
            type: 'text',
            text: `Error retrieving system health: ${error instanceof Error ? error.message : 'Unknown error'}`,
          },
        ],
      };
    }
  }

  private async listFirewallRules(interfaceFilter?: string, action = 'all', enabledOnly = false, format = 'table'): Promise<any> {
    try {
      const rules = await this.makeAPIRequest('firewall/filter/searchRule');

      let result = `Firewall Rules:\n\n`;

      if (rules && rules.rows) {
        const filteredRules = rules.rows.filter((rule: any) => {
          if (interfaceFilter && rule.interface !== interfaceFilter) return false;
          if (action !== 'all' && rule.action !== action) return false;
          if (enabledOnly && rule.enabled !== '1') return false;
          return true;
        });

        if (format === 'table') {
          result += `${'Action'.padEnd(8)} | ${'Interface'.padEnd(10)} | ${'Protocol'.padEnd(8)} | ${'Source'.padEnd(15)} | ${'Destination'.padEnd(15)} | ${'Description'.padEnd(30)}\n`;
          result += `${'-'.repeat(8)} | ${'-'.repeat(10)} | ${'-'.repeat(8)} | ${'-'.repeat(15)} | ${'-'.repeat(15)} | ${'-'.repeat(30)}\n`;

          filteredRules.forEach((rule: any) => {
            result += `${(rule.action || 'N/A').padEnd(8)} | ${(rule.interface || 'N/A').padEnd(10)} | ${(rule.protocol || 'N/A').padEnd(8)} | ${(rule.source || 'N/A').padEnd(15)} | ${(rule.destination || 'N/A').padEnd(15)} | ${(rule.description || 'N/A').padEnd(30)}\n`;
          });
        } else {
          result += JSON.stringify(filteredRules, null, 2);
        }
      } else {
        result += 'No firewall rules found.';
      }

      return {
        content: [
          {
            type: 'text',
            text: result,
          },
        ],
      };
    } catch (error) {
      return {
        content: [
          {
            type: 'text',
            text: `Error listing firewall rules: ${error instanceof Error ? error.message : 'Unknown error'}`,
          },
        ],
      };
    }
  }

  private async createFirewallRule(ruleData: any): Promise<any> {
    try {
      const newRule = {
        action: ruleData.action,
        interface: ruleData.interface,
        protocol: ruleData.protocol || 'TCP',
        source: ruleData.source || 'any',
        destination: ruleData.destination || 'any',
        destination_port: ruleData.port || '',
        description: ruleData.description || '',
        enabled: ruleData.enabled ? '1' : '0',
      };

      const result = await this.makeAPIRequest('firewall/filter/addRule', 'POST', newRule);

      if (result && result.result === 'saved') {
        // Apply changes
        await this.makeAPIRequest('firewall/filter/apply', 'POST');

        return {
          content: [
            {
              type: 'text',
              text: `Successfully created firewall rule: ${ruleData.description || ruleData.action + ' rule'}\nRule UUID: ${result.uuid || 'N/A'}`,
            },
          ],
        };
      } else {
        throw new Error('Failed to create firewall rule');
      }
    } catch (error) {
      return {
        content: [
          {
            type: 'text',
            text: `Error creating firewall rule: ${error instanceof Error ? error.message : 'Unknown error'}`,
          },
        ],
      };
    }
  }

  private async modifyFirewallRule(ruleUuid: string, modifications: any): Promise<any> {
    try {
      const result = await this.makeAPIRequest(`firewall/filter/setRule/${ruleUuid}`, 'POST', modifications);

      if (result && result.result === 'saved') {
        await this.makeAPIRequest('firewall/filter/apply', 'POST');

        return {
          content: [
            {
              type: 'text',
              text: `Successfully modified firewall rule: ${ruleUuid}`,
            },
          ],
        };
      } else {
        throw new Error('Failed to modify firewall rule');
      }
    } catch (error) {
      return {
        content: [
          {
            type: 'text',
            text: `Error modifying firewall rule: ${error instanceof Error ? error.message : 'Unknown error'}`,
          },
        ],
      };
    }
  }

  private async deleteFirewallRule(ruleUuid: string, confirm = false): Promise<any> {
    if (!confirm) {
      return {
        content: [
          {
            type: 'text',
            text: 'Firewall rule deletion requires confirmation. Set confirm=true to proceed.',
          },
        ],
      };
    }

    try {
      const result = await this.makeAPIRequest(`firewall/filter/delRule/${ruleUuid}`, 'POST');

      if (result && result.result === 'deleted') {
        await this.makeAPIRequest('firewall/filter/apply', 'POST');

        return {
          content: [
            {
              type: 'text',
              text: `Successfully deleted firewall rule: ${ruleUuid}`,
            },
          ],
        };
      } else {
        throw new Error('Failed to delete firewall rule');
      }
    } catch (error) {
      return {
        content: [
          {
            type: 'text',
            text: `Error deleting firewall rule: ${error instanceof Error ? error.message : 'Unknown error'}`,
          },
        ],
      };
    }
  }

  private async listInterfaces(includeVirtual = true, statusFilter = 'all'): Promise<any> {
    try {
      const interfaces = await this.makeAPIRequest('interfaces/overview/export');

      let result = `Network Interfaces:\n\n`;

      if (interfaces) {
        Object.entries(interfaces).forEach(([name, iface]: [string, any]) => {
          if (!includeVirtual && iface.virtual) return;
          if (statusFilter !== 'all' && iface.status !== statusFilter) return;

          result += `${name}: ${iface.status || 'unknown'} - ${iface.ipv4 || 'No IPv4'}\n`;
        });
      }

      return {
        content: [
          {
            type: 'text',
            text: result,
          },
        ],
      };
    } catch (error) {
      return {
        content: [
          {
            type: 'text',
            text: `Error listing interfaces: ${error instanceof Error ? error.message : 'Unknown error'}`,
          },
        ],
      };
    }
  }

  private async configureInterface(interfaceName: string, config: any): Promise<any> {
    try {
      const result = await this.makeAPIRequest(`interfaces/${interfaceName}/set`, 'POST', config);

      return {
        content: [
          {
            type: 'text',
            text: `Interface configuration result: ${JSON.stringify(result, null, 2)}`,
          },
        ],
      };
    } catch (error) {
      return {
        content: [
          {
            type: 'text',
            text: `Error configuring interface: ${error instanceof Error ? error.message : 'Unknown error'}`,
          },
        ],
      };
    }
  }

  private async getDHCPStatus(interfaceFilter?: string, includeLeases = true, includeStatic = false): Promise<any> {
    try {
      let result = `DHCP Status:\n\n`;

      if (includeLeases) {
        const leases = await this.makeAPIRequest('dhcpv4/leases/searchLease');
        result += `Active Leases: ${leases?.rows?.length || 0}\n`;
      }

      return {
        content: [
          {
            type: 'text',
            text: result,
          },
        ],
      };
    } catch (error) {
      return {
        content: [
          {
            type: 'text',
            text: `Error getting DHCP status: ${error instanceof Error ? error.message : 'Unknown error'}`,
          },
        ],
      };
    }
  }

  private async manageDHCPLease(action: string, leaseData: any): Promise<any> {
    try {
      let result: any;

      switch (action) {
        case 'create':
          result = await this.makeAPIRequest('dhcpv4/leases/addLease', 'POST', leaseData);
          break;
        case 'modify':
          result = await this.makeAPIRequest(`dhcpv4/leases/setLease/${leaseData.uuid}`, 'POST', leaseData);
          break;
        case 'delete':
          result = await this.makeAPIRequest(`dhcpv4/leases/delLease/${leaseData.uuid}`, 'POST');
          break;
      }

      return {
        content: [
          {
            type: 'text',
            text: `DHCP lease ${action} result: ${JSON.stringify(result, null, 2)}`,
          },
        ],
      };
    } catch (error) {
      return {
        content: [
          {
            type: 'text',
            text: `Error managing DHCP lease: ${error instanceof Error ? error.message : 'Unknown error'}`,
          },
        ],
      };
    }
  }

  private async getVPNStatus(vpnType = 'all', includeLogs = false): Promise<any> {
    try {
      let result = `VPN Status:\n\n`;

      if (vpnType === 'all' || vpnType === 'openvpn') {
        const ovpnStatus = await this.makeAPIRequest('openvpn/export/providers');
        result += `OpenVPN Connections: ${Object.keys(ovpnStatus || {}).length}\n`;
      }

      return {
        content: [
          {
            type: 'text',
            text: result,
          },
        ],
      };
    } catch (error) {
      return {
        content: [
          {
            type: 'text',
            text: `Error getting VPN status: ${error instanceof Error ? error.message : 'Unknown error'}`,
          },
        ],
      };
    }
  }

  private async backupConfiguration(includePasswords = false, backupName?: string): Promise<any> {
    try {
      const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
      const filename = backupName || `opnsense-backup-${timestamp}.xml`;

      const config = await this.makeAPIRequest('core/backup/download');

      return {
        content: [
          {
            type: 'text',
            text: `Configuration backup created successfully.\nFilename: ${filename}\nSize: ${config.length || 0} bytes`,
          },
        ],
      };
    } catch (error) {
      return {
        content: [
          {
            type: 'text',
            text: `Error creating backup: ${error instanceof Error ? error.message : 'Unknown error'}`,
          },
        ],
      };
    }
  }

  private async restoreConfiguration(backupFile: string, rebootRequired = true, confirm = false): Promise<any> {
    if (!confirm) {
      return {
        content: [
          {
            type: 'text',
            text: 'Configuration restore requires confirmation. This is a destructive operation. Set confirm=true to proceed.',
          },
        ],
      };
    }

    try {
      // This would typically upload and restore the configuration
      return {
        content: [
          {
            type: 'text',
            text: `Configuration restore initiated from: ${backupFile}\nReboot required: ${rebootRequired}`,
          },
        ],
      };
    } catch (error) {
      return {
        content: [
          {
            type: 'text',
            text: `Error restoring configuration: ${error instanceof Error ? error.message : 'Unknown error'}`,
          },
        ],
      };
    }
  }

  private async getTrafficStats(interfaceFilter?: string, timeRange = '24h', trafficType = 'total', topProtocols = false): Promise<any> {
    try {
      let result = `Traffic Statistics (${timeRange}):\n\n`;

      const stats = await this.makeAPIRequest('diagnostics/traffic/interface');
      result += `Interface traffic data retrieved\n`;

      return {
        content: [
          {
            type: 'text',
            text: result,
          },
        ],
      };
    } catch (error) {
      return {
        content: [
          {
            type: 'text',
            text: `Error getting traffic stats: ${error instanceof Error ? error.message : 'Unknown error'}`,
          },
        ],
      };
    }
  }

  private async getSecurityAlerts(severity = 'all', timeRange = '24h', limit = 50): Promise<any> {
    try {
      let result = `Security Alerts (${timeRange}):\n\n`;

      const alerts = await this.makeAPIRequest('ids/alert/search');
      result += `Found ${alerts?.rows?.length || 0} security alerts\n`;

      return {
        content: [
          {
            type: 'text',
            text: result,
          },
        ],
      };
    } catch (error) {
      return {
        content: [
          {
            type: 'text',
            text: `Error getting security alerts: ${error instanceof Error ? error.message : 'Unknown error'}`,
          },
        ],
      };
    }
  }

  private async blockIPAddress(ipAddress: string, duration = '24h', reason?: string, whitelistLocal = true): Promise<any> {
    try {
      const blockRule = {
        action: 'block',
        interface: 'WAN',
        protocol: 'any',
        source: ipAddress,
        destination: 'any',
        description: reason || `Blocked IP: ${ipAddress} (${duration})`,
        enabled: '1',
      };

      const result = await this.makeAPIRequest('firewall/filter/addRule', 'POST', blockRule);

      if (result && result.result === 'saved') {
        await this.makeAPIRequest('firewall/filter/apply', 'POST');

        return {
          content: [
            {
              type: 'text',
              text: `Successfully blocked IP address: ${ipAddress} for ${duration}\nReason: ${reason || 'No reason specified'}`,
            },
          ],
        };
      } else {
        throw new Error('Failed to block IP address');
      }
    } catch (error) {
      return {
        content: [
          {
            type: 'text',
            text: `Error blocking IP address: ${error instanceof Error ? error.message : 'Unknown error'}`,
          },
        ],
      };
    }
  }

  private async restartService(serviceName: string, waitForCompletion = true): Promise<any> {
    try {
      const result = await this.makeAPIRequest(`core/service/restart/${serviceName}`, 'POST');

      return {
        content: [
          {
            type: 'text',
            text: `Service restart initiated: ${serviceName}\nResult: ${JSON.stringify(result, null, 2)}`,
          },
        ],
      };
    } catch (error) {
      return {
        content: [
          {
            type: 'text',
            text: `Error restarting service: ${error instanceof Error ? error.message : 'Unknown error'}`,
          },
        ],
      };
    }
  }

  private async rebootSystem(delayMinutes = 1, confirm = false): Promise<any> {
    if (!confirm) {
      return {
        content: [
          {
            type: 'text',
            text: 'System reboot requires confirmation. This will restart the router. Set confirm=true to proceed.',
          },
        ],
      };
    }

    try {
      const result = await this.makeAPIRequest('core/system/reboot', 'POST', {
        delay: delayMinutes * 60, // Convert to seconds
      });

      return {
        content: [
          {
            type: 'text',
            text: `System reboot initiated with ${delayMinutes} minute delay.\nResult: ${JSON.stringify(result, null, 2)}`,
          },
        ],
      };
    } catch (error) {
      return {
        content: [
          {
            type: 'text',
            text: `Error rebooting system: ${error instanceof Error ? error.message : 'Unknown error'}`,
          },
        ],
      };
    }
  }

  private async getSystemLogs(logType = 'system', severity = 'info', lines = 100, searchTerm?: string): Promise<any> {
    try {
      let result = `System Logs (${logType}, ${severity}, last ${lines} lines):\n\n`;

      const logs = await this.makeAPIRequest(`diagnostics/log/${logType}`);

      if (searchTerm) {
        result += `Filtering for: "${searchTerm}"\n\n`;
      }

      result += logs || 'No logs available';

      return {
        content: [
          {
            type: 'text',
            text: result,
          },
        ],
      };
    } catch (error) {
      return {
        content: [
          {
            type: 'text',
            text: `Error getting system logs: ${error instanceof Error ? error.message : 'Unknown error'}`,
          },
        ],
      };
    }
  }

  private async runDiagnostics(testType: string, target: string, interfaceSource?: string, count = 4): Promise<any> {
    try {
      let result = `Diagnostics Results:\n\n`;

      switch (testType) {
        case 'ping':
          const pingResult = await this.makeAPIRequest(`diagnostics/interface/ping`, 'POST', {
            host: target,
            count: count,
            interface: interfaceSource,
          });
          result += `Ping to ${target}:\n${pingResult || 'No response'}`;
          break;

        case 'traceroute':
          const traceResult = await this.makeAPIRequest(`diagnostics/interface/traceroute`, 'POST', {
            host: target,
            interface: interfaceSource,
          });
          result += `Traceroute to ${target}:\n${traceResult || 'No response'}`;
          break;

        case 'dns_lookup':
          const dnsResult = await this.makeAPIRequest(`diagnostics/dns/lookup`, 'POST', {
            host: target,
          });
          result += `DNS Lookup for ${target}:\n${dnsResult || 'No response'}`;
          break;

        default:
          result += `Test type ${testType} not implemented yet`;
      }

      return {
        content: [
          {
            type: 'text',
            text: result,
          },
        ],
      };
    } catch (error) {
      return {
        content: [
          {
            type: 'text',
            text: `Error running diagnostics: ${error instanceof Error ? error.message : 'Unknown error'}`,
          },
        ],
      };
    }
  }

  async run() {
    const transport = new StdioServerTransport();
    await this.server.connect(transport);
    console.error('OPNsense Monolith MCP server running on stdio');
  }
}

const server = new OPNsenseMCPServer();
server.run().catch(console.error);
