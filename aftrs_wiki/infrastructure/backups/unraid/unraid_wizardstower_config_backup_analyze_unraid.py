#!/usr/bin/env python3
"""
Unraid Configuration Analysis Tool
Analyzes backup data and provides optimization recommendations
"""

import os
import json
import re
import subprocess
import argparse
from pathlib import Path
from datetime import datetime
from typing import Dict, List, Any, Optional

class UnraidAnalyzer:
    def __init__(self, backup_path: str):
        self.backup_path = Path(backup_path)
        self.analysis_results = {}
        
    def analyze_system_info(self) -> Dict[str, Any]:
        """Analyze basic system information"""
        system_info = {}
        
        # Read system info file
        system_file = self.backup_path / "system_info.txt"
        if system_file.exists():
            with open(system_file, 'r') as f:
                content = f.read()
                # Extract key information
                system_info['hostname'] = re.search(r'Hostname: (.+)', content)
                system_info['kernel'] = re.search(r'Kernel: (.+)', content)
        
        # Read CPU info
        cpu_file = self.backup_path / "performance" / "cpu" / "cpu_info.txt"
        if cpu_file.exists():
            with open(cpu_file, 'r') as f:
                content = f.read()
                cpu_model = re.search(r'CPU Model: (.+)', content)
                cpu_cores = re.search(r'CPU Cores: (\d+)', content)
                if cpu_model:
                    system_info['cpu_model'] = cpu_model.group(1).strip()
                if cpu_cores:
                    system_info['cpu_cores'] = int(cpu_cores.group(1))
        
        # Read memory info
        memory_file = self.backup_path / "performance" / "memory" / "memory_info.txt"
        if memory_file.exists():
            with open(memory_file, 'r') as f:
                content = f.read()
                # Extract total memory
                mem_match = re.search(r'Mem:\s+(\d+[GMK])\s+(\d+[GMK])\s+(\d+[GMK])', content)
                if mem_match:
                    system_info['total_memory'] = mem_match.group(1)
                    system_info['used_memory'] = mem_match.group(2)
                    system_info['free_memory'] = mem_match.group(3)
        
        return system_info
    
    def analyze_storage(self) -> Dict[str, Any]:
        """Analyze storage configuration and performance"""
        storage_info = {}
        
        # Read storage info
        storage_file = self.backup_path / "performance" / "storage" / "storage_info.txt"
        if storage_file.exists():
            with open(storage_file, 'r') as f:
                content = f.read()
                
                # Parse df output
                df_lines = [line for line in content.split('\n') if line.strip() and not line.startswith('Filesystem')]
                storage_info['filesystems'] = []
                
                for line in df_lines:
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
        
        return storage_info
    
    def analyze_docker(self) -> Dict[str, Any]:
        """Analyze Docker configuration and containers"""
        docker_info = {}
        
        # Read Docker containers
        containers_file = self.backup_path / "configs" / "docker" / "containers.txt"
        if containers_file.exists():
            with open(containers_file, 'r') as f:
                content = f.read()
                lines = content.strip().split('\n')[1:]  # Skip header
                docker_info['containers'] = []
                
                for line in lines:
                    if line.strip():
                        parts = line.split()
                        if len(parts) >= 2:
                            container_info = {
                                'container_id': parts[0],
                                'image': parts[1],
                                'status': ' '.join(parts[2:]) if len(parts) > 2 else 'Unknown'
                            }
                            docker_info['containers'].append(container_info)
        
        # Read Docker images
        images_file = self.backup_path / "configs" / "docker" / "images.txt"
        if images_file.exists():
            with open(images_file, 'r') as f:
                content = f.read()
                lines = content.strip().split('\n')[1:]  # Skip header
                docker_info['images'] = []
                
                for line in lines:
                    if line.strip():
                        parts = line.split()
                        if len(parts) >= 3:
                            image_info = {
                                'repository': parts[0],
                                'tag': parts[1],
                                'image_id': parts[2],
                                'size': parts[3] if len(parts) > 3 else 'Unknown'
                            }
                            docker_info['images'].append(image_info)
        
        return docker_info
    
    def analyze_network(self) -> Dict[str, Any]:
        """Analyze network configuration"""
        network_info = {}
        
        # Read network interfaces
        interfaces_file = self.backup_path / "configs" / "network" / "interfaces.txt"
        if interfaces_file.exists():
            with open(interfaces_file, 'r') as f:
                content = f.read()
                network_info['interfaces'] = []
                
                # Parse interface information
                current_interface = None
                for line in content.split('\n'):
                    if line.strip():
                        if re.match(r'^\d+:', line):
                            # New interface
                            if current_interface:
                                network_info['interfaces'].append(current_interface)
                            current_interface = {'name': line.split(':')[1].strip()}
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
        
        return network_info
    
    def analyze_performance(self) -> Dict[str, Any]:
        """Analyze system performance metrics"""
        performance_info = {}
        
        # Read CPU usage
        cpu_usage_file = self.backup_path / "performance" / "cpu" / "cpu_usage.txt"
        if cpu_usage_file.exists():
            with open(cpu_usage_file, 'r') as f:
                content = f.read()
                # Extract load average
                load_match = re.search(r'load average: ([\d.]+), ([\d.]+), ([\d.]+)', content)
                if load_match:
                    performance_info['load_average'] = {
                        '1min': float(load_match.group(1)),
                        '5min': float(load_match.group(2)),
                        '15min': float(load_match.group(3))
                    }
        
        # Read system load
        load_file = self.backup_path / "performance" / "system_load.txt"
        if load_file.exists():
            with open(load_file, 'r') as f:
                content = f.read()
                uptime_match = re.search(r'up (.+),', content)
                if uptime_match:
                    performance_info['uptime'] = uptime_match.group(1)
        
        return performance_info
    
    def generate_recommendations(self) -> List[str]:
        """Generate optimization recommendations based on analysis"""
        recommendations = []
        
        # Storage recommendations
        if 'filesystems' in self.analysis_results.get('storage', {}):
            for fs in self.analysis_results['storage']['filesystems']:
                use_percent = fs.get('use_percent', '0%').rstrip('%')
                try:
                    use_percent = int(use_percent)
                    if use_percent > 90:
                        recommendations.append(f"⚠️  Storage warning: {fs['filesystem']} is {use_percent}% full")
                    elif use_percent > 80:
                        recommendations.append(f"📊 Storage alert: {fs['filesystem']} is {use_percent}% full")
                except ValueError:
                    pass
        
        # Docker recommendations
        docker_info = self.analysis_results.get('docker', {})
        containers = docker_info.get('containers', [])
        if containers:
            running_containers = [c for c in containers if 'Up' in c.get('status', '')]
            stopped_containers = [c for c in containers if 'Exited' in c.get('status', '')]
            
            if stopped_containers:
                recommendations.append(f"🐳 Found {len(stopped_containers)} stopped Docker containers - consider cleanup")
            
            if len(containers) > 20:
                recommendations.append("🐳 High number of Docker containers - consider resource optimization")
        
        # Performance recommendations
        performance = self.analysis_results.get('performance', {})
        if 'load_average' in performance:
            load_15min = performance['load_average']['15min']
            cpu_cores = self.analysis_results.get('system', {}).get('cpu_cores', 1)
            
            if load_15min > cpu_cores * 2:
                recommendations.append("⚡ High system load detected - consider resource optimization")
            elif load_15min > cpu_cores:
                recommendations.append("📈 Moderate system load - monitor resource usage")
        
        # Memory recommendations
        system_info = self.analysis_results.get('system', {})
        if 'used_memory' in system_info and 'total_memory' in system_info:
            # Parse memory values (simplified)
            try:
                used = self._parse_memory_size(system_info['used_memory'])
                total = self._parse_memory_size(system_info['total_memory'])
                if used and total:
                    memory_usage = (used / total) * 100
                    if memory_usage > 90:
                        recommendations.append("💾 High memory usage detected - consider adding RAM or optimizing applications")
                    elif memory_usage > 80:
                        recommendations.append("📊 Moderate memory usage - monitor for potential optimization")
            except:
                pass
        
        # Network recommendations
        network_info = self.analysis_results.get('network', {})
        interfaces = network_info.get('interfaces', [])
        if len(interfaces) < 2:
            recommendations.append("🌐 Consider adding network redundancy for better reliability")
        
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
        """Run complete analysis"""
        print("🔍 Analyzing Unraid backup data...")
        
        self.analysis_results = {
            'system': self.analyze_system_info(),
            'storage': self.analyze_storage(),
            'docker': self.analyze_docker(),
            'network': self.analyze_network(),
            'performance': self.analyze_performance(),
            'recommendations': self.generate_recommendations()
        }
        
        return self.analysis_results
    
    def print_report(self):
        """Print formatted analysis report"""
        print("\n" + "="*60)
        print("🚀 UNRAID SYSTEM ANALYSIS REPORT")
        print("="*60)
        
        # System Overview
        system = self.analysis_results.get('system', {})
        print(f"\n📋 SYSTEM OVERVIEW:")
        print(f"   Hostname: {system.get('hostname', 'Unknown')}")
        print(f"   CPU: {system.get('cpu_model', 'Unknown')}")
        print(f"   Cores: {system.get('cpu_cores', 'Unknown')}")
        print(f"   Memory: {system.get('total_memory', 'Unknown')}")
        
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
        
        # Recommendations
        recommendations = self.analysis_results.get('recommendations', [])
        if recommendations:
            print(f"\n💡 OPTIMIZATION RECOMMENDATIONS:")
            for i, rec in enumerate(recommendations, 1):
                print(f"   {i}. {rec}")
        else:
            print(f"\n✅ No immediate optimization recommendations")
        
        print("\n" + "="*60)
    
    def save_report(self, output_file: str):
        """Save analysis report to JSON file"""
        with open(output_file, 'w') as f:
            json.dump(self.analysis_results, f, indent=2, default=str)
        print(f"📄 Analysis report saved to: {output_file}")

def main():
    parser = argparse.ArgumentParser(description='Analyze Unraid backup data')
    parser.add_argument('backup_path', help='Path to the backup directory')
    parser.add_argument('--output', '-o', help='Output JSON file for analysis results')
    
    args = parser.parse_args()
    
    if not os.path.exists(args.backup_path):
        print(f"❌ Error: Backup path '{args.backup_path}' does not exist")
        return 1
    
    analyzer = UnraidAnalyzer(args.backup_path)
    analyzer.run_analysis()
    analyzer.print_report()
    
    if args.output:
        analyzer.save_report(args.output)
    
    return 0

if __name__ == '__main__':
    exit(main()) 