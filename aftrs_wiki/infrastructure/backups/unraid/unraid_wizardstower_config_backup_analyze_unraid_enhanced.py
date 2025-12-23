#!/usr/bin/env python3
"""
Enhanced Unraid Configuration Analysis Tool
Features: HTML reports, trend analysis, advanced recommendations
"""

import os
import json
import re
import subprocess
import argparse
import yaml
from pathlib import Path
from datetime import datetime, timedelta
from typing import Dict, List, Any, Optional
import statistics

class EnhancedUnraidAnalyzer:
    def __init__(self, backup_path: str, config_file: str = "config.yaml"):
        self.backup_path = Path(backup_path)
        self.config_file = config_file
        self.analysis_results = {}
        self.config = self.load_config()
        
    def load_config(self) -> Dict[str, Any]:
        """Load configuration from YAML file"""
        default_config = {
            'analysis': {
                'thresholds': {
                    'storage_warning': 80,
                    'storage_critical': 90,
                    'memory_warning': 80,
                    'memory_critical': 90,
                    'load_warning': 1.0,
                    'load_critical': 2.0
                }
            }
        }
        
        if os.path.exists(self.config_file):
            try:
                with open(self.config_file, 'r') as f:
                    config = yaml.safe_load(f)
                    return config
            except Exception as e:
                print(f"Warning: Could not load config file: {e}")
                return default_config
        return default_config
    
    def analyze_system_info(self) -> Dict[str, Any]:
        """Analyze basic system information with enhanced details"""
        system_info = {}
        
        # Read system info file
        system_file = self.backup_path / "system_info.txt"
        if system_file.exists():
            with open(system_file, 'r') as f:
                content = f.read()
                # Extract key information with better regex
                hostname_match = re.search(r'Hostname:\s*(.+)', content)
                kernel_match = re.search(r'Kernel:\s*(.+)', content)
                boot_time_match = re.search(r'Boot Time:\s*(.+)', content)
                
                if hostname_match:
                    system_info['hostname'] = hostname_match.group(1).strip()
                if kernel_match:
                    system_info['kernel'] = kernel_match.group(1).strip()
                if boot_time_match:
                    system_info['boot_time'] = boot_time_match.group(1).strip()
        
        # Read enhanced CPU info
        cpu_file = self.backup_path / "performance" / "cpu" / "cpu_info.txt"
        if cpu_file.exists():
            with open(cpu_file, 'r') as f:
                content = f.read()
                cpu_model = re.search(r'CPU Model:\s*(.+)', content)
                cpu_cores = re.search(r'CPU Cores:\s*(\d+)', content)
                cpu_freq = re.search(r'CPU Frequency:\s*(.+)', content)
                cpu_cache = re.search(r'CPU Cache:\s*(.+)', content)
                
                if cpu_model:
                    system_info['cpu_model'] = cpu_model.group(1).strip()
                if cpu_cores:
                    system_info['cpu_cores'] = int(cpu_cores.group(1))
                if cpu_freq:
                    system_info['cpu_frequency'] = cpu_freq.group(1).strip()
                if cpu_cache:
                    system_info['cpu_cache'] = cpu_cache.group(1).strip()
        
        # Read enhanced memory info
        memory_file = self.backup_path / "performance" / "memory" / "memory_info.txt"
        if memory_file.exists():
            with open(memory_file, 'r') as f:
                content = f.read()
                # Extract memory information with better parsing
                mem_match = re.search(r'Mem:\s+(\d+[GMK])\s+(\d+[GMK])\s+(\d+[GMK])', content)
                if mem_match:
                    system_info['total_memory'] = mem_match.group(1)
                    system_info['used_memory'] = mem_match.group(2)
                    system_info['free_memory'] = mem_match.group(3)
                    
                    # Calculate memory usage percentage
                    try:
                        total = self._parse_memory_size(system_info['total_memory'])
                        used = self._parse_memory_size(system_info['used_memory'])
                        if total and used:
                            system_info['memory_usage_percent'] = (used / total) * 100
                    except:
                        pass
        
        return system_info
    
    def analyze_storage(self) -> Dict[str, Any]:
        """Analyze storage configuration and performance with enhanced details"""
        storage_info = {}
        
        # Read enhanced storage info
        storage_file = self.backup_path / "performance" / "storage" / "storage_info.txt"
        if storage_file.exists():
            with open(storage_file, 'r') as f:
                content = f.read()
                
                # Parse block devices
                storage_info['block_devices'] = []
                block_section = re.search(r'Block Devices:(.*?)Filesystem Usage:', content, re.DOTALL)
                if block_section:
                    lines = block_section.group(1).strip().split('\n')[1:]  # Skip header
                    for line in lines:
                        if line.strip() and not line.startswith('NAME'):
                            parts = line.split()
                            if len(parts) >= 6:
                                device_info = {
                                    'name': parts[0],
                                    'size': parts[1],
                                    'fstype': parts[2],
                                    'mountpoint': parts[3],
                                    'label': parts[4],
                                    'uuid': parts[5]
                                }
                                storage_info['block_devices'].append(device_info)
                
                # Parse filesystem usage
                storage_info['filesystems'] = []
                fs_section = re.search(r'Filesystem Usage:(.*?)RAID Status:', content, re.DOTALL)
                if fs_section:
                    lines = fs_section.group(1).strip().split('\n')[1:]  # Skip header
                    for line in lines:
                        if line.strip() and not line.startswith('Filesystem'):
                            parts = line.split()
                            if len(parts) >= 5:
                                fs_info = {
                                    'filesystem': parts[0],
                                    'size': parts[1],
                                    'used': parts[2],
                                    'available': parts[3],
                                    'use_percent': parts[4],
                                    'mountpoint': parts[5] if len(parts) > 5 else ''
                                }
                                storage_info['filesystems'].append(fs_info)
                
                # Parse RAID status
                raid_section = re.search(r'RAID Status:(.*?)Disk I/O Statistics:', content, re.DOTALL)
                if raid_section:
                    storage_info['raid_status'] = raid_section.group(1).strip()
        
        return storage_info
    
    def analyze_docker(self) -> Dict[str, Any]:
        """Analyze Docker configuration and containers with enhanced details"""
        docker_info = {}
        
        # Read enhanced Docker containers
        containers_file = self.backup_path / "configs" / "docker" / "containers.txt"
        if containers_file.exists():
            with open(containers_file, 'r') as f:
                content = f.read()
                lines = content.strip().split('\n')[1:]  # Skip header
                docker_info['containers'] = []
                
                for line in lines:
                    if line.strip():
                        parts = line.split('\t')
                        if len(parts) >= 6:
                            container_info = {
                                'container_id': parts[0],
                                'image': parts[1],
                                'status': parts[2],
                                'ports': parts[3],
                                'size': parts[4],
                                'names': parts[5]
                            }
                            docker_info['containers'].append(container_info)
        
        # Read enhanced Docker images
        images_file = self.backup_path / "configs" / "docker" / "images.txt"
        if images_file.exists():
            with open(images_file, 'r') as f:
                content = f.read()
                lines = content.strip().split('\n')[1:]  # Skip header
                docker_info['images'] = []
                
                for line in lines:
                    if line.strip():
                        parts = line.split('\t')
                        if len(parts) >= 5:
                            image_info = {
                                'repository': parts[0],
                                'tag': parts[1],
                                'image_id': parts[2],
                                'size': parts[3],
                                'created_at': parts[4] if len(parts) > 4 else 'Unknown'
                            }
                            docker_info['images'].append(image_info)
        
        # Read Docker disk usage
        disk_usage_file = self.backup_path / "configs" / "docker" / "disk_usage.txt"
        if disk_usage_file.exists():
            with open(disk_usage_file, 'r') as f:
                content = f.read()
                docker_info['disk_usage'] = content
        
        return docker_info
    
    def analyze_network(self) -> Dict[str, Any]:
        """Analyze network configuration with enhanced details"""
        network_info = {}
        
        # Read enhanced network interfaces
        interfaces_file = self.backup_path / "configs" / "network" / "interfaces.txt"
        if interfaces_file.exists():
            with open(interfaces_file, 'r') as f:
                content = f.read()
                network_info['interfaces'] = []
                
                # Parse interface information with better regex
                current_interface = None
                for line in content.split('\n'):
                    if line.strip():
                        if re.match(r'^\d+:', line):
                            # New interface
                            if current_interface:
                                network_info['interfaces'].append(current_interface)
                            interface_name = line.split(':')[1].strip()
                            current_interface = {'name': interface_name}
                        elif current_interface and 'inet ' in line:
                            # IP address
                            ip_match = re.search(r'inet (\d+\.\d+\.\d+\.\d+)', line)
                            if ip_match:
                                current_interface['ip'] = ip_match.group(1)
                        elif current_interface and 'UP' in line:
                            current_interface['status'] = 'UP'
                        elif current_interface and 'DOWN' in line:
                            current_interface['status'] = 'DOWN'
                
                if current_interface:
                    network_info['interfaces'].append(current_interface)
        
        # Read listening ports
        ports_file = self.backup_path / "configs" / "network" / "listening_ports.txt"
        if ports_file.exists():
            with open(ports_file, 'r') as f:
                content = f.read()
                network_info['listening_ports'] = []
                lines = content.strip().split('\n')[1:]  # Skip header
                for line in lines:
                    if line.strip():
                        parts = line.split()
                        if len(parts) >= 4:
                            port_info = {
                                'protocol': parts[0],
                                'local_address': parts[3],
                                'state': parts[5] if len(parts) > 5 else 'Unknown'
                            }
                            network_info['listening_ports'].append(port_info)
        
        return network_info
    
    def analyze_performance(self) -> Dict[str, Any]:
        """Analyze system performance metrics with enhanced details"""
        performance_info = {}
        
        # Read multiple CPU usage samples
        cpu_samples = []
        for i in range(1, 4):  # Check for 3 samples
            cpu_usage_file = self.backup_path / "performance" / "cpu" / f"cpu_usage_{i}.txt"
            if cpu_usage_file.exists():
                with open(cpu_usage_file, 'r') as f:
                    content = f.read()
                    # Extract load average
                    load_match = re.search(r'load average: ([\d.]+), ([\d.]+), ([\d.]+)', content)
                    if load_match:
                        cpu_samples.append({
                            '1min': float(load_match.group(1)),
                            '5min': float(load_match.group(2)),
                            '15min': float(load_match.group(3))
                        })
        
        if cpu_samples:
            # Calculate averages
            performance_info['load_average'] = {
                '1min': statistics.mean([s['1min'] for s in cpu_samples]),
                '5min': statistics.mean([s['5min'] for s in cpu_samples]),
                '15min': statistics.mean([s['15min'] for s in cpu_samples]),
                'samples': len(cpu_samples)
            }
        
        # Read system load
        load_file = self.backup_path / "performance" / "system_load.txt"
        if load_file.exists():
            with open(load_file, 'r') as f:
                content = f.read()
                uptime_match = re.search(r'up (.+),', content)
                if uptime_match:
                    performance_info['uptime'] = uptime_match.group(1)
        
        # Read process information
        top_cpu_file = self.backup_path / "performance" / "processes" / "top_cpu_processes.txt"
        if top_cpu_file.exists():
            with open(top_cpu_file, 'r') as f:
                content = f.read()
                performance_info['top_cpu_processes'] = []
                lines = content.strip().split('\n')[1:]  # Skip header
                for line in lines[:10]:  # Top 10 processes
                    if line.strip():
                        parts = line.split()
                        if len(parts) >= 11:
                            process_info = {
                                'user': parts[0],
                                'pid': parts[1],
                                'cpu_percent': parts[2],
                                'mem_percent': parts[3],
                                'command': ' '.join(parts[10:])
                            }
                            performance_info['top_cpu_processes'].append(process_info)
        
        return performance_info
    
    def generate_enhanced_recommendations(self) -> List[Dict[str, Any]]:
        """Generate enhanced optimization recommendations with severity levels"""
        recommendations = []
        thresholds = self.config.get('analysis', {}).get('thresholds', {})
        
        # Storage recommendations
        if 'filesystems' in self.analysis_results.get('storage', {}):
            for fs in self.analysis_results['storage']['filesystems']:
                use_percent = fs.get('use_percent', '0%').rstrip('%')
                try:
                    use_percent = int(use_percent)
                    if use_percent > thresholds.get('storage_critical', 90):
                        recommendations.append({
                            'type': 'storage',
                            'severity': 'critical',
                            'message': f"Storage critical: {fs['filesystem']} is {use_percent}% full",
                            'action': 'Immediate action required - consider expanding storage or cleaning up files'
                        })
                    elif use_percent > thresholds.get('storage_warning', 80):
                        recommendations.append({
                            'type': 'storage',
                            'severity': 'warning',
                            'message': f"Storage warning: {fs['filesystem']} is {use_percent}% full",
                            'action': 'Monitor usage and consider cleanup'
                        })
                except ValueError:
                    pass
        
        # Docker recommendations
        docker_info = self.analysis_results.get('docker', {})
        containers = docker_info.get('containers', [])
        if containers:
            running_containers = [c for c in containers if 'Up' in c.get('status', '')]
            stopped_containers = [c for c in containers if 'Exited' in c.get('status', '')]
            
            if stopped_containers:
                recommendations.append({
                    'type': 'docker',
                    'severity': 'info',
                    'message': f"Found {len(stopped_containers)} stopped Docker containers",
                    'action': 'Consider cleanup with: docker container prune'
                })
            
            if len(containers) > 20:
                recommendations.append({
                    'type': 'docker',
                    'severity': 'warning',
                    'message': f"High number of Docker containers: {len(containers)}",
                    'action': 'Consider resource optimization and container consolidation'
                })
        
        # Performance recommendations
        performance = self.analysis_results.get('performance', {})
        if 'load_average' in performance:
            load_15min = performance['load_average']['15min']
            cpu_cores = self.analysis_results.get('system', {}).get('cpu_cores', 1)
            
            if load_15min > cpu_cores * thresholds.get('load_critical', 2.0):
                recommendations.append({
                    'type': 'performance',
                    'severity': 'critical',
                    'message': f"High system load: {load_15min:.2f} (15min average)",
                    'action': 'Investigate high CPU usage processes and consider resource optimization'
                })
            elif load_15min > cpu_cores * thresholds.get('load_warning', 1.0):
                recommendations.append({
                    'type': 'performance',
                    'severity': 'warning',
                    'message': f"Moderate system load: {load_15min:.2f} (15min average)",
                    'action': 'Monitor resource usage and consider optimization'
                })
        
        # Memory recommendations
        system_info = self.analysis_results.get('system', {})
        if 'memory_usage_percent' in system_info:
            memory_usage = system_info['memory_usage_percent']
            if memory_usage > thresholds.get('memory_critical', 90):
                recommendations.append({
                    'type': 'memory',
                    'severity': 'critical',
                    'message': f"High memory usage: {memory_usage:.1f}%",
                    'action': 'Consider adding RAM or optimizing memory-intensive applications'
                })
            elif memory_usage > thresholds.get('memory_warning', 80):
                recommendations.append({
                    'type': 'memory',
                    'severity': 'warning',
                    'message': f"Moderate memory usage: {memory_usage:.1f}%",
                    'action': 'Monitor memory usage patterns'
                })
        
        # Network recommendations
        network_info = self.analysis_results.get('network', {})
        interfaces = network_info.get('interfaces', [])
        if len(interfaces) < 2:
            recommendations.append({
                'type': 'network',
                'severity': 'info',
                'message': 'Single network interface detected',
                'action': 'Consider adding network redundancy for better reliability'
            })
        
        return recommendations
    
    def _parse_memory_size(self, size_str: str) -> Optional[int]:
        """Parse memory size string to bytes"""
        if not size_str:
            return None
        
        size_str = size_str.strip()
        multipliers = {'K': 1024, 'M': 1024**2, 'G': 1024**3}
        
        for unit, multiplier in multipliers.items():
            if size_str.endswith(unit):
                try:
                    return int(float(size_str[:-1]) * multiplier)
                except ValueError:
                    return None
        
        try:
            return int(size_str)
        except ValueError:
            return None
    
    def run_analysis(self) -> Dict[str, Any]:
        """Run complete enhanced analysis"""
        print("🔍 Running enhanced Unraid analysis...")
        
        self.analysis_results = {
            'system': self.analyze_system_info(),
            'storage': self.analyze_storage(),
            'docker': self.analyze_docker(),
            'network': self.analyze_network(),
            'performance': self.analyze_performance(),
            'recommendations': self.generate_enhanced_recommendations()
        }
        
        return self.analysis_results
    
    def generate_html_report(self, output_file: str):
        """Generate an HTML report with charts and visualizations"""
        html_content = f"""
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Unraid System Analysis Report</title>
    <style>
        body {{
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            line-height: 1.6;
            margin: 0;
            padding: 20px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
        }}
        .container {{
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 15px;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
            overflow: hidden;
        }}
        .header {{
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            text-align: center;
        }}
        .header h1 {{
            margin: 0;
            font-size: 2.5em;
            font-weight: 300;
        }}
        .header p {{
            margin: 10px 0 0 0;
            opacity: 0.9;
        }}
        .content {{
            padding: 30px;
        }}
        .section {{
            margin-bottom: 40px;
            padding: 25px;
            border-radius: 10px;
            background: #f8f9fa;
            border-left: 5px solid #667eea;
        }}
        .section h2 {{
            color: #333;
            margin-top: 0;
            font-size: 1.5em;
        }}
        .metric {{
            display: inline-block;
            margin: 10px;
            padding: 15px;
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            min-width: 200px;
        }}
        .metric h3 {{
            margin: 0 0 10px 0;
            color: #667eea;
            font-size: 1.1em;
        }}
        .metric p {{
            margin: 0;
            font-size: 1.2em;
            font-weight: bold;
            color: #333;
        }}
        .recommendation {{
            margin: 15px 0;
            padding: 15px;
            border-radius: 8px;
            border-left: 4px solid;
        }}
        .recommendation.critical {{
            background: #ffe6e6;
            border-left-color: #dc3545;
        }}
        .recommendation.warning {{
            background: #fff3cd;
            border-left-color: #ffc107;
        }}
        .recommendation.info {{
            background: #d1ecf1;
            border-left-color: #17a2b8;
        }}
        .recommendation h4 {{
            margin: 0 0 10px 0;
            color: #333;
        }}
        .recommendation p {{
            margin: 0;
            color: #666;
        }}
        .chart-container {{
            margin: 20px 0;
            padding: 20px;
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }}
        .progress-bar {{
            width: 100%;
            height: 20px;
            background: #e9ecef;
            border-radius: 10px;
            overflow: hidden;
            margin: 10px 0;
        }}
        .progress-fill {{
            height: 100%;
            background: linear-gradient(90deg, #28a745, #20c997);
            transition: width 0.3s ease;
        }}
        .progress-fill.warning {{
            background: linear-gradient(90deg, #ffc107, #fd7e14);
        }}
        .progress-fill.critical {{
            background: linear-gradient(90deg, #dc3545, #e83e8c);
        }}
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 Unraid System Analysis Report</h1>
            <p>Generated on {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}</p>
        </div>
        
        <div class="content">
"""

        # System Overview Section
        system = self.analysis_results.get('system', {})
        html_content += f"""
            <div class="section">
                <h2>📋 System Overview</h2>
                <div class="metric">
                    <h3>Hostname</h3>
                    <p>{system.get('hostname', 'Unknown')}</p>
                </div>
                <div class="metric">
                    <h3>CPU Model</h3>
                    <p>{system.get('cpu_model', 'Unknown')}</p>
                </div>
                <div class="metric">
                    <h3>CPU Cores</h3>
                    <p>{system.get('cpu_cores', 'Unknown')}</p>
                </div>
                <div class="metric">
                    <h3>Total Memory</h3>
                    <p>{system.get('total_memory', 'Unknown')}</p>
                </div>
            </div>
"""

        # Storage Section
        storage = self.analysis_results.get('storage', {})
        if 'filesystems' in storage:
            html_content += """
            <div class="section">
                <h2>💾 Storage Overview</h2>
"""
            for fs in storage['filesystems']:
                use_percent = fs.get('use_percent', '0%').rstrip('%')
                try:
                    use_percent_int = int(use_percent)
                    progress_class = 'critical' if use_percent_int > 90 else 'warning' if use_percent_int > 80 else ''
                    html_content += f"""
                <div class="chart-container">
                    <h3>{fs['filesystem']}</h3>
                    <p>Used: {fs['used']} / {fs['size']} ({use_percent}%)</p>
                    <div class="progress-bar">
                        <div class="progress-fill {progress_class}" style="width: {use_percent}%"></div>
                    </div>
                </div>
"""
                except ValueError:
                    pass
            html_content += "</div>"

        # Docker Section
        docker = self.analysis_results.get('docker', {})
        containers = docker.get('containers', [])
        if containers:
            running = len([c for c in containers if 'Up' in c.get('status', '')])
            stopped = len([c for c in containers if 'Exited' in c.get('status', '')])
            html_content += f"""
            <div class="section">
                <h2>🐳 Docker Overview</h2>
                <div class="metric">
                    <h3>Total Containers</h3>
                    <p>{len(containers)}</p>
                </div>
                <div class="metric">
                    <h3>Running</h3>
                    <p>{running}</p>
                </div>
                <div class="metric">
                    <h3>Stopped</h3>
                    <p>{stopped}</p>
                </div>
            </div>
"""

        # Performance Section
        performance = self.analysis_results.get('performance', {})
        if 'load_average' in performance:
            load = performance['load_average']
            html_content += f"""
            <div class="section">
                <h2>⚡ Performance Overview</h2>
                <div class="metric">
                    <h3>Load Average (1m)</h3>
                    <p>{load['1min']:.2f}</p>
                </div>
                <div class="metric">
                    <h3>Load Average (5m)</h3>
                    <p>{load['5min']:.2f}</p>
                </div>
                <div class="metric">
                    <h3>Load Average (15m)</h3>
                    <p>{load['15min']:.2f}</p>
                </div>
            </div>
"""

        # Recommendations Section
        recommendations = self.analysis_results.get('recommendations', [])
        if recommendations:
            html_content += """
            <div class="section">
                <h2>💡 Optimization Recommendations</h2>
"""
            for rec in recommendations:
                severity_class = rec.get('severity', 'info')
                html_content += f"""
                <div class="recommendation {severity_class}">
                    <h4>{rec['message']}</h4>
                    <p>{rec['action']}</p>
                </div>
"""
            html_content += "</div>"

        html_content += """
        </div>
    </div>
</body>
</html>
"""

        with open(output_file, 'w') as f:
            f.write(html_content)
        
        print(f"📄 HTML report generated: {output_file}")
    
    def print_enhanced_report(self):
        """Print formatted enhanced analysis report"""
        print("\n" + "="*80)
        print("🚀 ENHANCED UNRAID SYSTEM ANALYSIS REPORT")
        print("="*80)
        
        # System Overview
        system = self.analysis_results.get('system', {})
        print(f"\n📋 SYSTEM OVERVIEW:")
        print(f"   Hostname: {system.get('hostname', 'Unknown')}")
        print(f"   CPU: {system.get('cpu_model', 'Unknown')}")
        print(f"   Cores: {system.get('cpu_cores', 'Unknown')}")
        print(f"   Memory: {system.get('total_memory', 'Unknown')}")
        if 'memory_usage_percent' in system:
            print(f"   Memory Usage: {system['memory_usage_percent']:.1f}%")
        
        # Storage Overview
        storage = self.analysis_results.get('storage', {})
        if 'filesystems' in storage:
            print(f"\n💾 STORAGE OVERVIEW:")
            for fs in storage['filesystems']:
                print(f"   {fs['filesystem']}: {fs['use_percent']} used ({fs['used']}/{fs['size']})")
        
        # Docker Overview
        docker = self.analysis_results.get('docker', {})
        containers = docker.get('containers', [])
        if containers:
            running = len([c for c in containers if 'Up' in c.get('status', '')])
            stopped = len([c for c in containers if 'Exited' in c.get('status', '')])
            print(f"\n🐳 DOCKER OVERVIEW:")
            print(f"   Total Containers: {len(containers)}")
            print(f"   Running: {running}")
            print(f"   Stopped: {stopped}")
        
        # Network Overview
        network = self.analysis_results.get('network', {})
        interfaces = network.get('interfaces', [])
        if interfaces:
            print(f"\n🌐 NETWORK OVERVIEW:")
            for iface in interfaces:
                status = iface.get('status', 'Unknown')
                ip = iface.get('ip', 'No IP')
                print(f"   {iface['name']}: {status} ({ip})")
        
        # Performance Overview
        performance = self.analysis_results.get('performance', {})
        if 'load_average' in performance:
            load = performance['load_average']
            print(f"\n⚡ PERFORMANCE OVERVIEW:")
            print(f"   Load Average: {load['1min']:.2f} (1m), {load['5min']:.2f} (5m), {load['15min']:.2f} (15m)")
        
        if 'uptime' in performance:
            print(f"   Uptime: {performance['uptime']}")
        
        # Enhanced Recommendations
        recommendations = self.analysis_results.get('recommendations', [])
        if recommendations:
            print(f"\n💡 ENHANCED OPTIMIZATION RECOMMENDATIONS:")
            for i, rec in enumerate(recommendations, 1):
                severity_icon = {
                    'critical': '🚨',
                    'warning': '⚠️',
                    'info': 'ℹ️'
                }.get(rec.get('severity', 'info'), 'ℹ️')
                print(f"   {i}. {severity_icon} {rec['message']}")
                print(f"      Action: {rec['action']}")
        else:
            print(f"\n✅ No immediate optimization recommendations")
        
        print("\n" + "="*80)
    
    def save_enhanced_report(self, output_file: str):
        """Save enhanced analysis report to JSON file"""
        with open(output_file, 'w') as f:
            json.dump(self.analysis_results, f, indent=2, default=str)
        print(f"📄 Enhanced analysis report saved to: {output_file}")

def main():
    parser = argparse.ArgumentParser(description='Enhanced Unraid backup analysis tool')
    parser.add_argument('backup_path', help='Path to the backup directory')
    parser.add_argument('--output', '-o', help='Output JSON file for analysis results')
    parser.add_argument('--html', help='Generate HTML report file')
    parser.add_argument('--config', '-c', default='config.yaml', help='Configuration file path')
    
    args = parser.parse_args()
    
    if not os.path.exists(args.backup_path):
        print(f"❌ Error: Backup path '{args.backup_path}' does not exist")
        return 1
    
    analyzer = EnhancedUnraidAnalyzer(args.backup_path, args.config)
    analyzer.run_analysis()
    analyzer.print_enhanced_report()
    
    if args.output:
        analyzer.save_enhanced_report(args.output)
    
    if args.html:
        analyzer.generate_html_report(args.html)
    
    return 0

if __name__ == '__main__':
    exit(main()) 