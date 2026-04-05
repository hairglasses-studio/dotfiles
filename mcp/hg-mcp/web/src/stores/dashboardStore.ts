import { create } from 'zustand';
import type { SystemStatus, Alert, DashboardSummary } from '../api/types';

// History entry for tracking metrics over time
interface HistoryEntry {
  timestamp: number;
  latency: number;
  healthScore: number;
  status: 'online' | 'offline' | 'degraded' | 'unknown';
}

// Maximum history entries per system (last 60 data points = ~30 min at 30s intervals)
const MAX_HISTORY = 60;

interface DashboardState {
  systems: SystemStatus[];
  alerts: Alert[];
  overallHealth: number;
  totalSystems: number;
  onlineCount: number;
  offlineCount: number;
  degradedCount: number;
  criticalIssues: number;
  isMonitoring: boolean;
  lastUpdated: Date | null;
  isLoading: boolean;
  error: string | null;

  // History tracking
  systemHistory: Record<string, HistoryEntry[]>;
  healthHistory: { timestamp: number; health: number }[];

  // Actions
  updateFromSummary: (summary: DashboardSummary) => void;
  setAlerts: (alerts: Alert[]) => void;
  addAlert: (alert: Alert) => void;
  resolveAlert: (alertId: string) => void;
  setMonitoring: (isMonitoring: boolean) => void;
  setLoading: (isLoading: boolean) => void;
  setError: (error: string | null) => void;

  // Computed
  getSystemsByCategory: () => Record<string, SystemStatus[]>;
  getCriticalSystems: () => SystemStatus[];
  getSystemHistory: (systemName: string) => HistoryEntry[];
  getHealthHistory: () => { timestamp: number; health: number }[];
}

export const useDashboardStore = create<DashboardState>((set, get) => ({
  systems: [],
  alerts: [],
  overallHealth: 0,
  totalSystems: 0,
  onlineCount: 0,
  offlineCount: 0,
  degradedCount: 0,
  criticalIssues: 0,
  isMonitoring: false,
  lastUpdated: null,
  isLoading: false,
  error: null,
  systemHistory: {},
  healthHistory: [],

  updateFromSummary: (summary) => {
    const { systemHistory, healthHistory } = get();
    const now = Date.now();

    // Update system history
    const newSystemHistory = { ...systemHistory };
    for (const system of summary.systems) {
      const history = newSystemHistory[system.name] || [];
      const newEntry: HistoryEntry = {
        timestamp: now,
        latency: system.latency,
        healthScore: system.healthScore,
        status: system.status,
      };
      newSystemHistory[system.name] = [...history.slice(-(MAX_HISTORY - 1)), newEntry];
    }

    // Update overall health history
    const newHealthHistory = [
      ...healthHistory.slice(-(MAX_HISTORY - 1)),
      { timestamp: now, health: summary.overallHealth },
    ];

    set({
      systems: summary.systems,
      overallHealth: summary.overallHealth,
      totalSystems: summary.totalSystems,
      onlineCount: summary.onlineCount,
      offlineCount: summary.offlineCount,
      degradedCount: summary.degradedCount,
      criticalIssues: summary.criticalIssues,
      lastUpdated: new Date(summary.lastUpdated),
      systemHistory: newSystemHistory,
      healthHistory: newHealthHistory,
    });
  },

  setAlerts: (alerts) => set({ alerts }),

  addAlert: (alert) => {
    const { alerts } = get();
    set({ alerts: [alert, ...alerts] });
  },

  resolveAlert: (alertId) => {
    const { alerts } = get();
    set({
      alerts: alerts.map((a) =>
        a.id === alertId ? { ...a, resolved: true } : a
      ),
    });
  },

  setMonitoring: (isMonitoring) => set({ isMonitoring }),
  setLoading: (isLoading) => set({ isLoading }),
  setError: (error) => set({ error }),

  getSystemsByCategory: () => {
    const { systems } = get();
    return systems.reduce((acc, system) => {
      if (!acc[system.category]) {
        acc[system.category] = [];
      }
      acc[system.category].push(system);
      return acc;
    }, {} as Record<string, SystemStatus[]>);
  },

  getCriticalSystems: () => {
    const { systems } = get();
    return systems.filter((s) => s.critical && s.status !== 'online');
  },

  getSystemHistory: (systemName: string) => {
    const { systemHistory } = get();
    return systemHistory[systemName] || [];
  },

  getHealthHistory: () => {
    const { healthHistory } = get();
    return healthHistory;
  },
}));
