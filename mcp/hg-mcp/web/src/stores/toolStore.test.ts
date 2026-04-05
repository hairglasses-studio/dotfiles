import { describe, it, expect, beforeEach } from 'vitest';
import { useToolStore } from './toolStore';

describe('toolStore', () => {
  beforeEach(() => {
    // Reset store state before each test
    useToolStore.setState({
      tools: [],
      favorites: [],
      recentTools: [],
      aliases: {},
      executionHistory: [],
      isLoading: false,
      error: null,
    });
  });

  describe('setTools', () => {
    it('should set tools array', () => {
      const mockTools = [
        {
          name: 'test_tool',
          description: 'Test description',
          category: 'test',
          subcategory: 'unit',
          tags: ['test'],
          useCases: ['testing'],
          complexity: 'simple' as const,
          isWrite: false,
          parameters: [],
        },
      ];

      useToolStore.getState().setTools(mockTools);
      expect(useToolStore.getState().tools).toEqual(mockTools);
    });
  });

  describe('toggleFavorite', () => {
    it('should add tool to favorites', () => {
      useToolStore.getState().toggleFavorite('test_tool');
      expect(useToolStore.getState().favorites).toContain('test_tool');
    });

    it('should remove tool from favorites if already present', () => {
      useToolStore.setState({ favorites: ['test_tool'] });
      useToolStore.getState().toggleFavorite('test_tool');
      expect(useToolStore.getState().favorites).not.toContain('test_tool');
    });
  });

  describe('addRecentTool', () => {
    it('should add tool to recent tools', () => {
      useToolStore.getState().addRecentTool('test_tool');
      expect(useToolStore.getState().recentTools).toContain('test_tool');
    });

    it('should limit recent tools to 20', () => {
      // Add 25 tools
      for (let i = 0; i < 25; i++) {
        useToolStore.getState().addRecentTool(`tool_${i}`);
      }
      expect(useToolStore.getState().recentTools.length).toBeLessThanOrEqual(20);
    });

    it('should not add duplicate consecutive tools', () => {
      useToolStore.getState().addRecentTool('test_tool');
      useToolStore.getState().addRecentTool('test_tool');
      const toolCount = useToolStore.getState().recentTools.filter(t => t === 'test_tool').length;
      expect(toolCount).toBe(1);
    });
  });

  describe('setAliases', () => {
    it('should set aliases', () => {
      const aliases = { st: 'status_tool', tr: 'trigger_tool' };
      useToolStore.getState().setAliases(aliases);
      expect(useToolStore.getState().aliases).toEqual(aliases);
    });
  });

  describe('clearFavorites', () => {
    it('should clear all favorites', () => {
      useToolStore.setState({ favorites: ['tool1', 'tool2', 'tool3'] });
      useToolStore.getState().clearFavorites();
      expect(useToolStore.getState().favorites).toEqual([]);
    });
  });

  describe('clearRecentTools', () => {
    it('should clear all recent tools', () => {
      useToolStore.setState({ recentTools: ['tool1', 'tool2', 'tool3'] });
      useToolStore.getState().clearRecentTools();
      expect(useToolStore.getState().recentTools).toEqual([]);
    });
  });

  describe('addExecution', () => {
    it('should add execution record', () => {
      const record = {
        toolName: 'test_tool',
        params: { arg: 'value' },
        result: { success: true, data: 'result', duration: 100 },
        timestamp: new Date().toISOString(),
      };
      useToolStore.getState().addExecution(record);
      expect(useToolStore.getState().executionHistory).toContainEqual(record);
    });

    it('should limit execution history to 100 records', () => {
      for (let i = 0; i < 120; i++) {
        useToolStore.getState().addExecution({
          toolName: `tool_${i}`,
          params: {},
          result: { success: true, data: null, duration: 0 },
          timestamp: new Date().toISOString(),
        });
      }
      expect(useToolStore.getState().executionHistory.length).toBeLessThanOrEqual(100);
    });
  });

  describe('setLoading', () => {
    it('should set loading state', () => {
      useToolStore.getState().setLoading(true);
      expect(useToolStore.getState().isLoading).toBe(true);
    });
  });

  describe('setError', () => {
    it('should set error message', () => {
      useToolStore.getState().setError('Test error');
      expect(useToolStore.getState().error).toBe('Test error');
    });

    it('should clear error with null', () => {
      useToolStore.setState({ error: 'Previous error' });
      useToolStore.getState().setError(null);
      expect(useToolStore.getState().error).toBeNull();
    });
  });
});
