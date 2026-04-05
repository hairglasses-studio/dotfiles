import { describe, it, expect, beforeEach } from 'vitest';
import { useDashboardStore } from './dashboardStore';

describe('dashboardStore', () => {
  beforeEach(() => {
    useDashboardStore.setState({
      systems: [],
      overallHealth: 100,
      onlineCount: 0,
      offlineCount: 0,
      degradedCount: 0,
      totalSystems: 0,
      criticalIssues: 0,
      alerts: [],
      lastUpdated: null,
      isLoading: false,
      error: null,
    });
  });

  describe('updateFromSummary', () => {
    it('should update state from dashboard summary', () => {
      const mockSummary = {
        totalSystems: 10,
        onlineCount: 8,
        offlineCount: 1,
        degradedCount: 1,
        unknownCount: 0,
        overallHealth: 80,
        criticalIssues: 1,
        activeAlerts: 2,
        lastUpdated: '2026-01-04T12:00:00Z',
        systems: [
          {
            name: 'resolume',
            category: 'video',
            status: 'online' as const,
            latency: 10,
            lastCheck: '2026-01-04T12:00:00Z',
            critical: true,
            healthScore: 100,
          },
        ],
      };

      useDashboardStore.getState().updateFromSummary(mockSummary);
      const state = useDashboardStore.getState();

      expect(state.overallHealth).toBe(80);
      expect(state.onlineCount).toBe(8);
      expect(state.offlineCount).toBe(1);
      expect(state.degradedCount).toBe(1);
      expect(state.criticalIssues).toBe(1);
      expect(state.systems).toHaveLength(1);
      expect(state.systems[0].name).toBe('resolume');
    });
  });

  describe('setAlerts', () => {
    it('should set alerts array', () => {
      const mockAlerts = [
        {
          id: '1',
          system: 'resolume',
          severity: 'error' as const,
          message: 'Connection lost',
          timestamp: '2026-01-04T12:00:00Z',
          resolved: false,
        },
        {
          id: '2',
          system: 'obs',
          severity: 'warning' as const,
          message: 'High latency',
          timestamp: '2026-01-04T12:00:00Z',
          resolved: false,
        },
      ];

      useDashboardStore.getState().setAlerts(mockAlerts);
      expect(useDashboardStore.getState().alerts).toEqual(mockAlerts);
    });
  });

  describe('setLoading', () => {
    it('should set loading state', () => {
      useDashboardStore.getState().setLoading(true);
      expect(useDashboardStore.getState().isLoading).toBe(true);

      useDashboardStore.getState().setLoading(false);
      expect(useDashboardStore.getState().isLoading).toBe(false);
    });
  });

  describe('setError', () => {
    it('should set error message', () => {
      useDashboardStore.getState().setError('Network error');
      expect(useDashboardStore.getState().error).toBe('Network error');
    });

    it('should clear error with null', () => {
      useDashboardStore.setState({ error: 'Previous error' });
      useDashboardStore.getState().setError(null);
      expect(useDashboardStore.getState().error).toBeNull();
    });
  });

  describe('status filtering', () => {
    it('should track systems by status', () => {
      const mockSummary = {
        totalSystems: 5,
        onlineCount: 2,
        offlineCount: 2,
        degradedCount: 1,
        unknownCount: 0,
        overallHealth: 60,
        criticalIssues: 2,
        activeAlerts: 0,
        lastUpdated: '2026-01-04T12:00:00Z',
        systems: [
          { name: 'sys1', category: 'video', status: 'online' as const, latency: 10, lastCheck: '', critical: false, healthScore: 100 },
          { name: 'sys2', category: 'video', status: 'online' as const, latency: 15, lastCheck: '', critical: false, healthScore: 95 },
          { name: 'sys3', category: 'audio', status: 'offline' as const, latency: 0, lastCheck: '', critical: true, healthScore: 0 },
          { name: 'sys4', category: 'audio', status: 'offline' as const, latency: 0, lastCheck: '', critical: true, healthScore: 0 },
          { name: 'sys5', category: 'lighting', status: 'degraded' as const, latency: 500, lastCheck: '', critical: false, healthScore: 50 },
        ],
      };

      useDashboardStore.getState().updateFromSummary(mockSummary);
      const state = useDashboardStore.getState();

      expect(state.onlineCount).toBe(2);
      expect(state.offlineCount).toBe(2);
      expect(state.degradedCount).toBe(1);
    });
  });
});
