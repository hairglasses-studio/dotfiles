import { describe, it, expect } from 'vitest';
import { cn, formatLatency, getStatusColor, getStatusBgColor } from './utils';

describe('cn utility', () => {
  it('should merge class names', () => {
    expect(cn('foo', 'bar')).toBe('foo bar');
  });

  it('should handle conditional classes', () => {
    expect(cn('base', true && 'active', false && 'inactive')).toBe('base active');
  });

  it('should handle undefined values', () => {
    expect(cn('base', undefined, 'end')).toBe('base end');
  });

  it('should handle empty strings', () => {
    expect(cn('base', '', 'end')).toBe('base end');
  });
});

describe('formatLatency', () => {
  it('should format milliseconds correctly', () => {
    expect(formatLatency(50)).toBe('50ms');
    expect(formatLatency(100)).toBe('100ms');
  });

  it('should format seconds correctly', () => {
    expect(formatLatency(1000)).toBe('1.0s');
    expect(formatLatency(1500)).toBe('1.5s');
    expect(formatLatency(2500)).toBe('2.5s');
  });

  it('should handle zero and small values', () => {
    expect(formatLatency(0)).toBe('<1ms');
    expect(formatLatency(0.5)).toBe('<1ms');
  });

  it('should handle large values', () => {
    expect(formatLatency(60000)).toBe('60.0s');
  });

  it('should handle negative values', () => {
    expect(formatLatency(-1)).toBe('-');
  });
});

describe('getStatusColor', () => {
  it('should return correct color for online status', () => {
    expect(getStatusColor('online')).toBe('text-status-online');
  });

  it('should return correct color for offline status', () => {
    expect(getStatusColor('offline')).toBe('text-status-offline');
  });

  it('should return correct color for degraded status', () => {
    expect(getStatusColor('degraded')).toBe('text-status-degraded');
  });

  it('should return correct color for unknown status', () => {
    expect(getStatusColor('unknown')).toBe('text-status-unknown');
  });
});

describe('getStatusBgColor', () => {
  it('should return correct background color for online status', () => {
    expect(getStatusBgColor('online')).toBe('bg-status-online/20');
  });

  it('should return correct background color for offline status', () => {
    expect(getStatusBgColor('offline')).toBe('bg-status-offline/20');
  });

  it('should return correct background color for degraded status', () => {
    expect(getStatusBgColor('degraded')).toBe('bg-status-degraded/20');
  });

  it('should return correct background color for unknown status', () => {
    expect(getStatusBgColor('unknown')).toBe('bg-status-unknown/20');
  });
});
