import { useEffect, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  Activity,
  AlertTriangle,
  CheckCircle,
  XCircle,
  Clock,
  Wifi,
  WifiOff,
  ChevronDown,
  ChevronRight,
  Server,
  Zap,
  RefreshCw,
  ExternalLink,
  Circle,
  TrendingUp,
  TrendingDown,
  Minus,
  BarChart3,
} from 'lucide-react';
import { api } from '../api/client';
import { useDashboardStore } from '../stores/dashboardStore';
import { useDashboardSSE } from '../hooks/useSSE';
import { cn, formatLatency, getStatusColor, getStatusBgColor } from '../lib/utils';
import { HealthSparkline, LatencySparkline, StatusTimeline, MiniGauge } from '../components/charts/Sparkline';
import type { SystemStatus } from '../api/types';

export function Dashboard() {
  const {
    systems,
    overallHealth,
    onlineCount,
    offlineCount,
    degradedCount,
    criticalIssues,
    alerts,
    lastUpdated,
    updateFromSummary,
    setAlerts,
    setLoading,
    setError,
    getSystemHistory,
    getHealthHistory,
  } = useDashboardStore();

  const [expandedCategories, setExpandedCategories] = useState<Set<string>>(new Set(['video', 'audio', 'lighting']));
  const [selectedSystem, setSelectedSystem] = useState<SystemStatus | null>(null);
  const [viewMode, setViewMode] = useState<'grid' | 'list' | 'compact'>('compact');

  // SSE for real-time updates
  const { isConnected, isConnecting, error: sseError, reconnect } = useDashboardSSE();

  // Fallback polling when SSE is not connected
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['dashboard'],
    queryFn: () => api.getDashboardStatus(),
    refetchInterval: isConnected ? false : 30000,
    enabled: !isConnected,
  });

  const { data: alertsData } = useQuery({
    queryKey: ['alerts'],
    queryFn: () => api.getAlerts(),
    refetchInterval: isConnected ? false : 10000,
    enabled: !isConnected,
  });

  useEffect(() => {
    if (data && !isConnected) {
      updateFromSummary(data);
    }
  }, [data, updateFromSummary, isConnected]);

  useEffect(() => {
    if (alertsData && !isConnected) {
      setAlerts(alertsData);
    }
  }, [alertsData, setAlerts, isConnected]);

  useEffect(() => {
    setLoading(isLoading || isConnecting);
    if (error && !isConnected) {
      setError(error instanceof Error ? error.message : 'Unknown error');
    } else if (sseError && !isConnected) {
      setError(sseError);
    } else {
      setError(null);
    }
  }, [isLoading, isConnecting, error, sseError, isConnected, setLoading, setError]);

  const activeAlerts = alerts.filter((a) => !a.resolved);

  // Group systems by category
  const systemsByCategory = systems.reduce((acc, system) => {
    if (!acc[system.category]) {
      acc[system.category] = [];
    }
    acc[system.category].push(system);
    return acc;
  }, {} as Record<string, typeof systems>);

  const categoryOrder = ['video', 'audio', 'lighting', 'network', 'automation', 'ai'];
  const sortedCategories = Object.keys(systemsByCategory).sort((a, b) => {
    const aIdx = categoryOrder.indexOf(a);
    const bIdx = categoryOrder.indexOf(b);
    if (aIdx === -1 && bIdx === -1) return a.localeCompare(b);
    if (aIdx === -1) return 1;
    if (bIdx === -1) return -1;
    return aIdx - bIdx;
  });

  const toggleCategory = (category: string) => {
    const newExpanded = new Set(expandedCategories);
    if (newExpanded.has(category)) {
      newExpanded.delete(category);
    } else {
      newExpanded.add(category);
    }
    setExpandedCategories(newExpanded);
  };

  const getCategoryStats = (category: string) => {
    const catSystems = systemsByCategory[category] || [];
    const online = catSystems.filter(s => s.status === 'online').length;
    const critical = catSystems.filter(s => s.critical && s.status !== 'online').length;
    return { total: catSystems.length, online, critical };
  };

  return (
    <div className="space-y-4">
      {/* Header Row - Compact */}
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-4">
          <h1 className="text-xl font-bold">System Monitor</h1>
          <div
            className={cn(
              'flex items-center gap-1.5 px-2 py-1 rounded-full text-xs font-medium',
              isConnected
                ? 'bg-green-500/20 text-green-400'
                : isConnecting
                ? 'bg-yellow-500/20 text-yellow-400'
                : 'bg-red-500/20 text-red-400'
            )}
          >
            {isConnected ? <Wifi className="w-3 h-3" /> : <WifiOff className="w-3 h-3" />}
            {isConnected ? 'Live' : isConnecting ? 'Connecting' : 'Offline'}
          </div>
        </div>
        <div className="flex items-center gap-2">
          {/* View Mode Toggle */}
          <div className="flex rounded-md border border-border overflow-hidden">
            {(['compact', 'grid', 'list'] as const).map((mode) => (
              <button
                key={mode}
                onClick={() => setViewMode(mode)}
                className={cn(
                  'px-2 py-1 text-xs capitalize',
                  viewMode === mode ? 'bg-primary text-primary-foreground' : 'bg-card hover:bg-accent'
                )}
              >
                {mode}
              </button>
            ))}
          </div>
          <button
            onClick={() => isConnected ? refetch() : reconnect()}
            className="flex items-center gap-1 px-3 py-1.5 rounded-md bg-primary text-primary-foreground hover:bg-primary/90 text-sm"
          >
            <RefreshCw className="w-3 h-3" />
            Refresh
          </button>
        </div>
      </div>

      {/* Stats Bar - Information Dense */}
      <div className="grid grid-cols-3 sm:grid-cols-5 lg:grid-cols-10 gap-2">
        {/* Health with Sparkline */}
        <div className="col-span-2 flex items-center gap-2 px-3 py-2 rounded-md bg-card border border-border">
          <MiniGauge value={overallHealth} size={36} strokeWidth={4} />
          <div className="flex-1 min-w-0">
            <div className="text-xs text-muted-foreground">Health</div>
            <HealthSparkline data={getHealthHistory().map(h => h.health)} width={60} height={16} />
          </div>
        </div>
        <StatCard label="Online" value={onlineCount} color="green" />
        <StatCard label="Offline" value={offlineCount} color="red" />
        <StatCard label="Degraded" value={degradedCount} color="yellow" />
        <StatCard label="Critical" value={criticalIssues} color={criticalIssues > 0 ? 'red' : 'muted'} />
        <StatCard label="Alerts" value={activeAlerts.length} color={activeAlerts.length > 0 ? 'orange' : 'muted'} />
        <StatCard label="Systems" value={systems.length} color="blue" />
        <div className="flex items-center gap-1 px-2 py-1 rounded-md bg-card border border-border text-xs">
          <Clock className="w-3 h-3 text-muted-foreground" />
          <span className="text-muted-foreground">{lastUpdated?.toLocaleTimeString() || '--:--'}</span>
        </div>
      </div>

      {/* Active Alerts - Compact */}
      {activeAlerts.length > 0 && (
        <div className="rounded-md border border-orange-500/30 bg-orange-500/5 p-2">
          <div className="flex items-center gap-2 text-sm text-orange-400 font-medium mb-2">
            <AlertTriangle className="w-4 h-4" />
            Active Alerts ({activeAlerts.length})
          </div>
          <div className="flex flex-wrap gap-2">
            {activeAlerts.slice(0, 6).map((alert) => (
              <div
                key={alert.id}
                className={cn(
                  'flex items-center gap-1.5 px-2 py-1 rounded text-xs',
                  alert.severity === 'critical' && 'bg-red-500/20 text-red-400',
                  alert.severity === 'error' && 'bg-orange-500/20 text-orange-400',
                  alert.severity === 'warning' && 'bg-yellow-500/20 text-yellow-400'
                )}
              >
                <Circle className="w-1.5 h-1.5 fill-current" />
                <span className="font-medium">{alert.system}</span>
                <span className="text-muted-foreground truncate max-w-32">{alert.message}</span>
              </div>
            ))}
            {activeAlerts.length > 6 && (
              <span className="text-xs text-muted-foreground">+{activeAlerts.length - 6} more</span>
            )}
          </div>
        </div>
      )}

      {/* Main Content Area */}
      <div className="flex gap-4">
        {/* Systems Panel */}
        <div className="flex-1 space-y-2">
          {sortedCategories.map((category) => {
            const stats = getCategoryStats(category);
            const isExpanded = expandedCategories.has(category);
            const catSystems = systemsByCategory[category];

            return (
              <div key={category} className="rounded-lg border border-border bg-card overflow-hidden">
                {/* Category Header */}
                <button
                  onClick={() => toggleCategory(category)}
                  className="w-full flex items-center justify-between px-3 py-2 hover:bg-accent/50 transition-colors"
                >
                  <div className="flex items-center gap-2">
                    {isExpanded ? <ChevronDown className="w-4 h-4" /> : <ChevronRight className="w-4 h-4" />}
                    <span className="font-semibold capitalize">{category}</span>
                    <span className="text-xs text-muted-foreground">({stats.total})</span>
                  </div>
                  <div className="flex items-center gap-2">
                    {stats.critical > 0 && (
                      <span className="flex items-center gap-1 text-xs text-red-400">
                        <AlertTriangle className="w-3 h-3" />
                        {stats.critical}
                      </span>
                    )}
                    <div className="flex items-center gap-1">
                      <span className="text-xs text-green-400">{stats.online}</span>
                      <span className="text-xs text-muted-foreground">/</span>
                      <span className="text-xs text-muted-foreground">{stats.total}</span>
                    </div>
                    <CategoryHealthBar systems={catSystems} />
                  </div>
                </button>

                {/* Category Systems */}
                {isExpanded && (
                  <div className={cn(
                    'border-t border-border',
                    viewMode === 'grid' && 'grid grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-2 p-2',
                    viewMode === 'list' && 'divide-y divide-border',
                    viewMode === 'compact' && 'grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-1 p-1'
                  )}>
                    {catSystems.map((system) => (
                      viewMode === 'list' ? (
                        <SystemListItem
                          key={system.name}
                          system={system}
                          isSelected={selectedSystem?.name === system.name}
                          onClick={() => setSelectedSystem(selectedSystem?.name === system.name ? null : system)}
                        />
                      ) : viewMode === 'grid' ? (
                        <SystemGridCard
                          key={system.name}
                          system={system}
                          isSelected={selectedSystem?.name === system.name}
                          onClick={() => setSelectedSystem(selectedSystem?.name === system.name ? null : system)}
                        />
                      ) : (
                        <SystemCompactCard
                          key={system.name}
                          system={system}
                          isSelected={selectedSystem?.name === system.name}
                          onClick={() => setSelectedSystem(selectedSystem?.name === system.name ? null : system)}
                        />
                      )
                    ))}
                  </div>
                )}
              </div>
            );
          })}
        </div>

        {/* System Detail Panel */}
        {selectedSystem && (
          <div className="w-80 flex-shrink-0 hidden lg:block">
            <SystemDetailPanel system={selectedSystem} history={getSystemHistory(selectedSystem.name)} onClose={() => setSelectedSystem(null)} />
          </div>
        )}
      </div>

      {/* Mobile System Detail */}
      {selectedSystem && (
        <div className="lg:hidden fixed inset-0 z-50 bg-black/50" onClick={() => setSelectedSystem(null)}>
          <div
            className="absolute bottom-0 left-0 right-0 max-h-[70vh] bg-card border-t border-border rounded-t-xl overflow-auto"
            onClick={(e) => e.stopPropagation()}
          >
            <SystemDetailPanel system={selectedSystem} history={getSystemHistory(selectedSystem.name)} onClose={() => setSelectedSystem(null)} />
          </div>
        </div>
      )}
    </div>
  );
}

function StatCard({ label, value, color, trend }: { label: string; value: string | number; color: string; trend?: 'up' | 'down' | 'neutral' }) {
  const colorClasses: Record<string, string> = {
    green: 'text-green-400',
    red: 'text-red-400',
    yellow: 'text-yellow-400',
    orange: 'text-orange-400',
    blue: 'text-blue-400',
    purple: 'text-purple-400',
    muted: 'text-muted-foreground',
  };

  return (
    <div className="flex items-center justify-between px-2 py-1.5 rounded-md bg-card border border-border">
      <span className="text-xs text-muted-foreground">{label}</span>
      <div className="flex items-center gap-1">
        {trend && (
          trend === 'up' ? <TrendingUp className="w-3 h-3 text-green-400" /> :
          trend === 'down' ? <TrendingDown className="w-3 h-3 text-red-400" /> :
          <Minus className="w-3 h-3 text-muted-foreground" />
        )}
        <span className={cn('text-sm font-semibold', colorClasses[color])}>{value}</span>
      </div>
    </div>
  );
}

function CategoryHealthBar({ systems }: { systems: SystemStatus[] }) {
  const total = systems.length;
  if (total === 0) return null;

  const online = systems.filter(s => s.status === 'online').length;
  const degraded = systems.filter(s => s.status === 'degraded').length;
  const offline = systems.filter(s => s.status === 'offline').length;

  return (
    <div className="w-16 h-2 rounded-full bg-secondary overflow-hidden flex">
      <div className="h-full bg-green-500" style={{ width: `${(online / total) * 100}%` }} />
      <div className="h-full bg-yellow-500" style={{ width: `${(degraded / total) * 100}%` }} />
      <div className="h-full bg-red-500" style={{ width: `${(offline / total) * 100}%` }} />
    </div>
  );
}

function SystemCompactCard({ system, isSelected, onClick }: { system: SystemStatus; isSelected: boolean; onClick: () => void }) {
  const StatusIcon = system.status === 'online' ? CheckCircle : system.status === 'offline' ? XCircle : AlertTriangle;

  return (
    <button
      onClick={onClick}
      className={cn(
        'flex items-center gap-2 px-2 py-1.5 rounded text-left transition-all w-full',
        isSelected ? 'bg-primary/20 border border-primary' : 'bg-secondary/50 hover:bg-secondary border border-transparent',
        system.status === 'offline' && system.critical && 'border-red-500/50'
      )}
    >
      <StatusIcon className={cn('w-3.5 h-3.5 flex-shrink-0', getStatusColor(system.status))} />
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-1">
          <span className="text-xs font-medium truncate">{system.name}</span>
          {system.critical && <Zap className="w-2.5 h-2.5 text-primary" />}
        </div>
      </div>
      <div className="flex items-center gap-1.5 flex-shrink-0">
        <LatencyBar latency={system.latency} />
        <span className="text-[10px] text-muted-foreground w-10 text-right">{formatLatency(system.latency)}</span>
      </div>
    </button>
  );
}

function SystemGridCard({ system, isSelected, onClick }: { system: SystemStatus; isSelected: boolean; onClick: () => void }) {
  const StatusIcon = system.status === 'online' ? CheckCircle : system.status === 'offline' ? XCircle : AlertTriangle;

  return (
    <button
      onClick={onClick}
      className={cn(
        'flex flex-col p-3 rounded-lg text-left transition-all',
        isSelected ? 'bg-primary/20 ring-2 ring-primary' : 'bg-secondary/30 hover:bg-secondary/50',
        getStatusBgColor(system.status)
      )}
    >
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-2">
          <StatusIcon className={cn('w-4 h-4', getStatusColor(system.status))} />
          <span className="font-medium text-sm">{system.name}</span>
        </div>
        {system.critical && <Zap className="w-3.5 h-3.5 text-primary" />}
      </div>
      <div className="flex items-center justify-between text-xs">
        <span className={cn('capitalize', getStatusColor(system.status))}>{system.status}</span>
        <span className="text-muted-foreground">{formatLatency(system.latency)}</span>
      </div>
      <div className="mt-2">
        <div className="flex items-center gap-1 text-[10px] text-muted-foreground">
          <span>Health:</span>
          <div className="flex-1 h-1 rounded-full bg-secondary overflow-hidden">
            <div
              className={cn(
                'h-full rounded-full',
                system.healthScore >= 80 ? 'bg-green-500' : system.healthScore >= 50 ? 'bg-yellow-500' : 'bg-red-500'
              )}
              style={{ width: `${system.healthScore}%` }}
            />
          </div>
          <span>{system.healthScore}%</span>
        </div>
      </div>
    </button>
  );
}

function SystemListItem({ system, isSelected, onClick }: { system: SystemStatus; isSelected: boolean; onClick: () => void }) {
  const StatusIcon = system.status === 'online' ? CheckCircle : system.status === 'offline' ? XCircle : AlertTriangle;

  return (
    <button
      onClick={onClick}
      className={cn(
        'flex items-center gap-4 px-4 py-3 text-left w-full hover:bg-accent/50 transition-colors',
        isSelected && 'bg-primary/10'
      )}
    >
      <StatusIcon className={cn('w-5 h-5 flex-shrink-0', getStatusColor(system.status))} />
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="font-medium">{system.name}</span>
          {system.critical && <span className="text-[10px] px-1 py-0.5 rounded bg-primary/20 text-primary">Critical</span>}
        </div>
        {system.message && (
          <p className="text-xs text-muted-foreground truncate mt-0.5">{system.message}</p>
        )}
      </div>
      <div className="flex items-center gap-4 flex-shrink-0">
        <div className="text-right">
          <div className="text-sm font-medium">{system.healthScore}%</div>
          <div className="text-xs text-muted-foreground">Health</div>
        </div>
        <div className="text-right">
          <div className="text-sm font-medium">{formatLatency(system.latency)}</div>
          <div className="text-xs text-muted-foreground">Latency</div>
        </div>
        <ChevronRight className="w-4 h-4 text-muted-foreground" />
      </div>
    </button>
  );
}

function LatencyBar({ latency }: { latency: number }) {
  const ms = latency / 1000000; // Convert nanoseconds to ms
  const width = Math.min(100, Math.max(5, ms < 10 ? 100 : ms < 50 ? 80 : ms < 100 ? 60 : ms < 500 ? 40 : 20));
  const color = ms < 50 ? 'bg-green-500' : ms < 200 ? 'bg-yellow-500' : 'bg-red-500';

  return (
    <div className="w-8 h-1.5 rounded-full bg-secondary overflow-hidden">
      <div className={cn('h-full rounded-full', color)} style={{ width: `${width}%` }} />
    </div>
  );
}

interface HistoryEntry {
  timestamp: number;
  latency: number;
  healthScore: number;
  status: 'online' | 'offline' | 'degraded' | 'unknown';
}

function SystemDetailPanel({ system, history, onClose }: { system: SystemStatus; history: HistoryEntry[]; onClose: () => void }) {
  const StatusIcon = system.status === 'online' ? CheckCircle : system.status === 'offline' ? XCircle : AlertTriangle;

  // Calculate uptime from history
  const uptimePercent = history.length > 0
    ? Math.round((history.filter(h => h.status === 'online').length / history.length) * 100)
    : 0;

  return (
    <div className="rounded-lg border border-border bg-card overflow-hidden">
      {/* Header */}
      <div className={cn('p-4 border-b border-border', getStatusBgColor(system.status))}>
        <div className="flex items-center justify-between mb-2">
          <div className="flex items-center gap-2">
            <StatusIcon className={cn('w-5 h-5', getStatusColor(system.status))} />
            <h3 className="font-semibold text-lg">{system.name}</h3>
          </div>
          <button onClick={onClose} className="p-1 rounded hover:bg-accent lg:hidden">
            <XCircle className="w-4 h-4" />
          </button>
        </div>
        <div className="flex items-center gap-2">
          <span className={cn('px-2 py-0.5 rounded text-xs capitalize', getStatusBgColor(system.status), getStatusColor(system.status))}>
            {system.status}
          </span>
          {system.critical && (
            <span className="px-2 py-0.5 rounded text-xs bg-primary/20 text-primary">Critical System</span>
          )}
          <span className="text-xs text-muted-foreground capitalize">{system.category}</span>
        </div>
      </div>

      {/* Metrics */}
      <div className="p-4 space-y-4">
        {/* Status Timeline */}
        {history.length > 0 && (
          <div>
            <div className="flex items-center justify-between text-sm mb-2">
              <span className="text-muted-foreground flex items-center gap-1">
                <BarChart3 className="w-3.5 h-3.5" />
                Status History
              </span>
              <span className="text-xs text-muted-foreground">{uptimePercent}% uptime</span>
            </div>
            <StatusTimeline data={history} width={240} height={10} />
          </div>
        )}

        {/* Health Score with Chart */}
        <div>
          <div className="flex items-center justify-between text-sm mb-1">
            <span className="text-muted-foreground">Health Score</span>
            <span className="font-semibold">{system.healthScore}%</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="flex-1 h-2 rounded-full bg-secondary overflow-hidden">
              <div
                className={cn(
                  'h-full rounded-full transition-all',
                  system.healthScore >= 80 ? 'bg-green-500' : system.healthScore >= 50 ? 'bg-yellow-500' : 'bg-red-500'
                )}
                style={{ width: `${system.healthScore}%` }}
              />
            </div>
            {history.length > 1 && (
              <HealthSparkline data={history.map(h => h.healthScore)} width={60} height={20} />
            )}
          </div>
        </div>

        {/* Latency with Chart */}
        <div>
          <div className="flex items-center justify-between text-sm mb-1">
            <span className="text-muted-foreground">Response Time</span>
            <span className="font-semibold">{formatLatency(system.latency)}</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="flex-1 grid grid-cols-5 gap-1 h-6">
              {[10, 50, 100, 200, 500].map((threshold, i) => {
                const ms = system.latency / 1000000;
                const isActive = ms <= threshold || i === 4;
                const isPast = ms > (i === 0 ? 0 : [10, 50, 100, 200][i - 1]);
                return (
                  <div key={threshold} className="flex flex-col items-center">
                    <div className={cn(
                      'flex-1 w-full rounded',
                      isActive && isPast ? 'bg-green-500' : isActive ? 'bg-green-500/30' : 'bg-secondary'
                    )} />
                    <span className="text-[8px] text-muted-foreground mt-0.5">{threshold}</span>
                  </div>
                );
              })}
            </div>
            {history.length > 1 && (
              <LatencySparkline data={history.map(h => h.latency)} width={60} height={20} />
            )}
          </div>
        </div>

        {/* Connection Info */}
        <div className="space-y-2">
          <h4 className="text-sm font-medium flex items-center gap-2">
            <Server className="w-4 h-4" />
            Connection
          </h4>
          <div className="grid grid-cols-2 gap-2 text-xs">
            <div className="px-2 py-1.5 rounded bg-secondary">
              <span className="text-muted-foreground block">Last Check</span>
              <span className="font-medium">{new Date(system.lastCheck).toLocaleTimeString()}</span>
            </div>
            <div className="px-2 py-1.5 rounded bg-secondary">
              <span className="text-muted-foreground block">Category</span>
              <span className="font-medium capitalize">{system.category}</span>
            </div>
            <div className="px-2 py-1.5 rounded bg-secondary">
              <span className="text-muted-foreground block">Data Points</span>
              <span className="font-medium">{history.length}</span>
            </div>
            <div className="px-2 py-1.5 rounded bg-secondary">
              <span className="text-muted-foreground block">Uptime</span>
              <span className={cn('font-medium', uptimePercent >= 90 ? 'text-green-400' : uptimePercent >= 50 ? 'text-yellow-400' : 'text-red-400')}>
                {uptimePercent}%
              </span>
            </div>
          </div>
        </div>

        {/* Error Message */}
        {system.message && (
          <div className="space-y-2">
            <h4 className="text-sm font-medium flex items-center gap-2 text-red-400">
              <AlertTriangle className="w-4 h-4" />
              Error Details
            </h4>
            <div className="p-2 rounded bg-red-500/10 border border-red-500/20 text-xs text-red-400 font-mono break-all">
              {system.message}
            </div>
          </div>
        )}

        {/* Quick Actions */}
        <div className="space-y-2">
          <h4 className="text-sm font-medium">Quick Actions</h4>
          <div className="flex flex-wrap gap-2">
            <button className="flex items-center gap-1 px-2 py-1 rounded text-xs bg-secondary hover:bg-secondary/80">
              <RefreshCw className="w-3 h-3" />
              Refresh
            </button>
            <button className="flex items-center gap-1 px-2 py-1 rounded text-xs bg-secondary hover:bg-secondary/80">
              <Activity className="w-3 h-3" />
              View Logs
            </button>
            <button className="flex items-center gap-1 px-2 py-1 rounded text-xs bg-secondary hover:bg-secondary/80">
              <ExternalLink className="w-3 h-3" />
              Open UI
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
