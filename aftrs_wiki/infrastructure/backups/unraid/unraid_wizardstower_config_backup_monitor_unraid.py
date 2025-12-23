#!/usr/bin/env python3
"""
Unraid System Monitoring Tool
Continuous health monitoring and alerting
"""

import os
import json
import time
import subprocess
import argparse
import yaml
from pathlib import Path
from datetime import datetime, timedelta
from typing import Dict, List, Any, Optional
import smtplib
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart
import re

class UnraidMonitor:
    def __init__(self, config_file: str = "config.yaml"):
        self.config_file = config_file
        self.config = self.load_config()
        self.alert_history = {}
        self.monitoring_data = {}
        
    def load_config(self) -> Dict[str, Any]:
        """Load monitoring configuration"""
        default_config = {
            'monitoring': {
                'check_interval': 300,  # 5 minutes
                'alert_on_failure': True,
                'retention_days': 7
            },
            'thresholds': {
                'cpu_warning': 80,
                'cpu_critical': 95,
                'memory_warning': 80,
                'memory_critical': 95,
                'storage_warning': 80,
                'storage_critical': 90,
                'load_warning': 1.0,
                'load_critical': 2.0
            },
            'alerts': {
                'email_enabled': False,
                'webhook_enabled': False,
                'log_enabled': True
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
    
    def get_system_metrics(self) -> Dict[str, Any]:
        """Collect current system metrics"""
        metrics = {}
        
        # CPU usage
        try:
            cpu_output = subprocess.check_output(['top', '-bn1'], text=True)
            cpu_match = re.search(r'%Cpu\(s\):\s+(\d+\.\d+)', cpu_output)
            if cpu_match:
                metrics['cpu_usage'] = float(cpu_match.group(1))
        except:
            metrics['cpu_usage'] = 0
        
        # Memory usage
        try:
            mem_output = subprocess.check_output(['free'], text=True)
            mem_lines = mem_output.strip().split('\n')
            if len(mem_lines) > 1:
                mem_parts = mem_lines[1].split()
                if len(mem_parts) >= 3:
                    total = int(mem_parts[1])
                    used = int(mem_parts[2])
                    metrics['memory_usage'] = (used / total) * 100
        except:
            metrics['memory_usage'] = 0
        
        # Load average
        try:
            load_output = subprocess.check_output(['cat', '/proc/loadavg'], text=True)
            load_parts = load_output.strip().split()
            if len(load_parts) >= 3:
                metrics['load_1min'] = float(load_parts[0])
                metrics['load_5min'] = float(load_parts[1])
                metrics['load_15min'] = float(load_parts[2])
        except:
            metrics['load_1min'] = 0
            metrics['load_5min'] = 0
            metrics['load_15min'] = 0
        
        # Storage usage
        try:
            df_output = subprocess.check_output(['df', '/'], text=True)
            df_lines = df_output.strip().split('\n')
            if len(df_lines) > 1:
                df_parts = df_lines[1].split()
                if len(df_parts) >= 5:
                    usage_str = df_parts[4].rstrip('%')
                    metrics['storage_usage'] = float(usage_str)
        except:
            metrics['storage_usage'] = 0
        
        # Docker status
        try:
            docker_output = subprocess.check_output(['docker', 'ps', '--format', '{{.Status}}'], text=True)
            running_containers = len([line for line in docker_output.strip().split('\n') if line and 'Up' in line])
            metrics['docker_running'] = running_containers
        except:
            metrics['docker_running'] = 0
        
        # Network status
        try:
            network_output = subprocess.check_output(['ip', 'link', 'show'], text=True)
            up_interfaces = len([line for line in network_output.strip().split('\n') if 'UP' in line])
            metrics['network_interfaces_up'] = up_interfaces
        except:
            metrics['network_interfaces_up'] = 0
        
        metrics['timestamp'] = datetime.now().isoformat()
        return metrics
    
    def check_thresholds(self, metrics: Dict[str, Any]) -> List[Dict[str, Any]]:
        """Check metrics against thresholds and generate alerts"""
        alerts = []
        thresholds = self.config.get('thresholds', {})
        
        # CPU checks
        if 'cpu_usage' in metrics:
            cpu_usage = metrics['cpu_usage']
            if cpu_usage > thresholds.get('cpu_critical', 95):
                alerts.append({
                    'type': 'cpu',
                    'severity': 'critical',
                    'message': f'CPU usage critical: {cpu_usage:.1f}%',
                    'value': cpu_usage,
                    'threshold': thresholds.get('cpu_critical', 95)
                })
            elif cpu_usage > thresholds.get('cpu_warning', 80):
                alerts.append({
                    'type': 'cpu',
                    'severity': 'warning',
                    'message': f'CPU usage high: {cpu_usage:.1f}%',
                    'value': cpu_usage,
                    'threshold': thresholds.get('cpu_warning', 80)
                })
        
        # Memory checks
        if 'memory_usage' in metrics:
            memory_usage = metrics['memory_usage']
            if memory_usage > thresholds.get('memory_critical', 95):
                alerts.append({
                    'type': 'memory',
                    'severity': 'critical',
                    'message': f'Memory usage critical: {memory_usage:.1f}%',
                    'value': memory_usage,
                    'threshold': thresholds.get('memory_critical', 95)
                })
            elif memory_usage > thresholds.get('memory_warning', 80):
                alerts.append({
                    'type': 'memory',
                    'severity': 'warning',
                    'message': f'Memory usage high: {memory_usage:.1f}%',
                    'value': memory_usage,
                    'threshold': thresholds.get('memory_warning', 80)
                })
        
        # Storage checks
        if 'storage_usage' in metrics:
            storage_usage = metrics['storage_usage']
            if storage_usage > thresholds.get('storage_critical', 90):
                alerts.append({
                    'type': 'storage',
                    'severity': 'critical',
                    'message': f'Storage usage critical: {storage_usage:.1f}%',
                    'value': storage_usage,
                    'threshold': thresholds.get('storage_critical', 90)
                })
            elif storage_usage > thresholds.get('storage_warning', 80):
                alerts.append({
                    'type': 'storage',
                    'severity': 'warning',
                    'message': f'Storage usage high: {storage_usage:.1f}%',
                    'value': storage_usage,
                    'threshold': thresholds.get('storage_warning', 80)
                })
        
        # Load average checks
        if 'load_15min' in metrics:
            load_15min = metrics['load_15min']
            cpu_cores = os.cpu_count() or 1
            
            if load_15min > cpu_cores * thresholds.get('load_critical', 2.0):
                alerts.append({
                    'type': 'load',
                    'severity': 'critical',
                    'message': f'System load critical: {load_15min:.2f}',
                    'value': load_15min,
                    'threshold': cpu_cores * thresholds.get('load_critical', 2.0)
                })
            elif load_15min > cpu_cores * thresholds.get('load_warning', 1.0):
                alerts.append({
                    'type': 'load',
                    'severity': 'warning',
                    'message': f'System load high: {load_15min:.2f}',
                    'value': load_15min,
                    'threshold': cpu_cores * thresholds.get('load_warning', 1.0)
                })
        
        return alerts
    
    def should_send_alert(self, alert: Dict[str, Any]) -> bool:
        """Check if alert should be sent (avoid spam)"""
        alert_key = f"{alert['type']}_{alert['severity']}"
        now = datetime.now()
        
        # Check if we've sent this alert recently
        if alert_key in self.alert_history:
            last_alert = self.alert_history[alert_key]
            time_diff = now - last_alert
            
            # Don't send same alert more than once per hour
            if time_diff.total_seconds() < 3600:
                return False
        
        # Update alert history
        self.alert_history[alert_key] = now
        return True
    
    def send_email_alert(self, alerts: List[Dict[str, Any]]):
        """Send email alerts"""
        email_config = self.config.get('output', {}).get('email', {})
        
        if not email_config.get('enabled', False):
            return
        
        try:
            msg = MIMEMultipart()
            msg['From'] = email_config.get('username', '')
            msg['To'] = ', '.join(email_config.get('recipients', []))
            msg['Subject'] = f"Unraid System Alert - {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}"
            
            body = "Unraid System Monitoring Alert\n\n"
            for alert in alerts:
                body += f"🚨 {alert['message']}\n"
                body += f"   Severity: {alert['severity'].upper()}\n"
                body += f"   Value: {alert['value']}\n"
                body += f"   Threshold: {alert['threshold']}\n\n"
            
            msg.attach(MIMEText(body, 'plain'))
            
            server = smtplib.SMTP(email_config.get('smtp_server', ''), email_config.get('smtp_port', 587))
            server.starttls()
            server.login(email_config.get('username', ''), email_config.get('password', ''))
            server.send_message(msg)
            server.quit()
            
            print(f"📧 Email alert sent to {len(email_config.get('recipients', []))} recipients")
        except Exception as e:
            print(f"❌ Failed to send email alert: {e}")
    
    def log_alert(self, alerts: List[Dict[str, Any]]):
        """Log alerts to file"""
        if not self.config.get('alerts', {}).get('log_enabled', True):
            return
        
        log_file = "unraid_monitor.log"
        timestamp = datetime.now().strftime('%Y-%m-%d %H:%M:%S')
        
        with open(log_file, 'a') as f:
            f.write(f"\n[{timestamp}] System Alerts:\n")
            for alert in alerts:
                f.write(f"  {alert['severity'].upper()}: {alert['message']}\n")
    
    def save_monitoring_data(self, metrics: Dict[str, Any], alerts: List[Dict[str, Any]]):
        """Save monitoring data for trend analysis"""
        data = {
            'timestamp': metrics['timestamp'],
            'metrics': metrics,
            'alerts': alerts
        }
        
        # Save to JSON file for trend analysis
        monitoring_file = "monitoring_data.json"
        try:
            if os.path.exists(monitoring_file):
                with open(monitoring_file, 'r') as f:
                    existing_data = json.load(f)
            else:
                existing_data = []
            
            existing_data.append(data)
            
            # Keep only recent data (last 7 days by default)
            retention_days = self.config.get('monitoring', {}).get('retention_days', 7)
            cutoff_time = datetime.now() - timedelta(days=retention_days)
            
            filtered_data = [
                entry for entry in existing_data
                if datetime.fromisoformat(entry['timestamp']) > cutoff_time
            ]
            
            with open(monitoring_file, 'w') as f:
                json.dump(filtered_data, f, indent=2)
                
        except Exception as e:
            print(f"❌ Failed to save monitoring data: {e}")
    
    def print_status(self, metrics: Dict[str, Any], alerts: List[Dict[str, Any]]):
        """Print current system status"""
        print(f"\n{'='*60}")
        print(f"🖥️  UNRAID SYSTEM STATUS - {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        print(f"{'='*60}")
        
        print(f"\n📊 CURRENT METRICS:")
        print(f"   CPU Usage: {metrics.get('cpu_usage', 0):.1f}%")
        print(f"   Memory Usage: {metrics.get('memory_usage', 0):.1f}%")
        print(f"   Storage Usage: {metrics.get('storage_usage', 0):.1f}%")
        print(f"   Load Average: {metrics.get('load_15min', 0):.2f} (15min)")
        print(f"   Docker Containers: {metrics.get('docker_running', 0)} running")
        print(f"   Network Interfaces: {metrics.get('network_interfaces_up', 0)} up")
        
        if alerts:
            print(f"\n🚨 ACTIVE ALERTS:")
            for alert in alerts:
                severity_icon = {
                    'critical': '🚨',
                    'warning': '⚠️',
                    'info': 'ℹ️'
                }.get(alert['severity'], 'ℹ️')
                print(f"   {severity_icon} {alert['message']}")
        else:
            print(f"\n✅ No active alerts - system healthy")
        
        print(f"\n{'='*60}")
    
    def run_monitoring_cycle(self):
        """Run one monitoring cycle"""
        try:
            # Collect metrics
            metrics = self.get_system_metrics()
            
            # Check thresholds
            alerts = self.check_thresholds(metrics)
            
            # Filter alerts to avoid spam
            new_alerts = [alert for alert in alerts if self.should_send_alert(alert)]
            
            # Print status
            self.print_status(metrics, new_alerts)
            
            # Send alerts if any
            if new_alerts:
                self.send_email_alert(new_alerts)
                self.log_alert(new_alerts)
            
            # Save monitoring data
            self.save_monitoring_data(metrics, alerts)
            
            return len(new_alerts)
            
        except Exception as e:
            print(f"❌ Monitoring cycle failed: {e}")
            return 0
    
    def start_monitoring(self, continuous: bool = True):
        """Start continuous monitoring"""
        print("🚀 Starting Unraid system monitoring...")
        print(f"📋 Configuration: {self.config_file}")
        print(f"⏰ Check interval: {self.config.get('monitoring', {}).get('check_interval', 300)} seconds")
        
        if not continuous:
            return self.run_monitoring_cycle()
        
        try:
            while True:
                self.run_monitoring_cycle()
                
                # Wait for next check
                interval = self.config.get('monitoring', {}).get('check_interval', 300)
                time.sleep(interval)
                
        except KeyboardInterrupt:
            print("\n🛑 Monitoring stopped by user")
        except Exception as e:
            print(f"❌ Monitoring failed: {e}")

def main():
    parser = argparse.ArgumentParser(description='Unraid System Monitor')
    parser.add_argument('--config', '-c', default='config.yaml', help='Configuration file path')
    parser.add_argument('--once', action='store_true', help='Run monitoring once instead of continuously')
    parser.add_argument('--interval', '-i', type=int, help='Override check interval in seconds')
    
    args = parser.parse_args()
    
    monitor = UnraidMonitor(args.config)
    
    # Override interval if specified
    if args.interval:
        monitor.config['monitoring']['check_interval'] = args.interval
    
    if args.once:
        monitor.run_monitoring_cycle()
    else:
        monitor.start_monitoring()

if __name__ == '__main__':
    main() 