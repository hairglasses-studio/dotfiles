import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatLatency(ms: number): string {
  if (ms < 0) return '-';
  if (ms < 1) return '<1ms';
  if (ms < 1000) return `${Math.round(ms)}ms`;
  return `${(ms / 1000).toFixed(1)}s`;
}

export function formatTimeAgo(date: Date | string): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  const now = new Date();
  const diffMs = now.getTime() - d.getTime();
  const diffSecs = Math.floor(diffMs / 1000);
  const diffMins = Math.floor(diffSecs / 60);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffSecs < 60) return 'just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  return `${diffDays}d ago`;
}

export function getStatusColor(status: string): string {
  switch (status) {
    case 'online':
      return 'text-status-online';
    case 'offline':
      return 'text-status-offline';
    case 'degraded':
      return 'text-status-degraded';
    default:
      return 'text-status-unknown';
  }
}

export function getStatusBgColor(status: string): string {
  switch (status) {
    case 'online':
      return 'bg-status-online/20';
    case 'offline':
      return 'bg-status-offline/20';
    case 'degraded':
      return 'bg-status-degraded/20';
    default:
      return 'bg-status-unknown/20';
  }
}

export function getSeverityColor(severity: string): string {
  switch (severity) {
    case 'critical':
      return 'text-red-500';
    case 'error':
      return 'text-orange-500';
    case 'warning':
      return 'text-yellow-500';
    case 'info':
      return 'text-blue-500';
    default:
      return 'text-gray-500';
  }
}

export function getComplexityColor(complexity: string): string {
  switch (complexity) {
    case 'simple':
      return 'text-green-400 bg-green-400/10';
    case 'moderate':
      return 'text-yellow-400 bg-yellow-400/10';
    case 'complex':
      return 'text-red-400 bg-red-400/10';
    default:
      return 'text-gray-400 bg-gray-400/10';
  }
}
