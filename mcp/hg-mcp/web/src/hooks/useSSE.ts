import { useEffect, useRef, useCallback, useState } from 'react';
import { useDashboardStore } from '../stores/dashboardStore';
import type { DashboardSummary, Alert } from '../api/types';

interface SSEOptions {
  url: string;
  onOpen?: () => void;
  onError?: (error: Event) => void;
  reconnectInterval?: number;
  maxRetries?: number;
}

interface SSEState {
  isConnected: boolean;
  isConnecting: boolean;
  error: string | null;
  retryCount: number;
}

export function useSSE(options: SSEOptions) {
  const {
    url,
    onOpen,
    onError,
    reconnectInterval = 5000,
    maxRetries = 10,
  } = options;

  const eventSourceRef = useRef<EventSource | null>(null);
  const retryCountRef = useRef(0);
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const [state, setState] = useState<SSEState>({
    isConnected: false,
    isConnecting: false,
    error: null,
    retryCount: 0,
  });

  const { updateFromSummary, setAlerts, addAlert } = useDashboardStore();

  const connect = useCallback(() => {
    if (eventSourceRef.current?.readyState === EventSource.OPEN) {
      return;
    }

    setState(prev => ({ ...prev, isConnecting: true, error: null }));

    try {
      const eventSource = new EventSource(url);
      eventSourceRef.current = eventSource;

      eventSource.onopen = () => {
        setState({
          isConnected: true,
          isConnecting: false,
          error: null,
          retryCount: 0,
        });
        retryCountRef.current = 0;
        onOpen?.();
      };

      eventSource.onerror = (event) => {
        setState(prev => ({
          ...prev,
          isConnected: false,
          isConnecting: false,
          error: 'Connection lost',
        }));
        onError?.(event);
        eventSource.close();

        // Attempt reconnection
        if (retryCountRef.current < maxRetries) {
          retryCountRef.current++;
          setState(prev => ({ ...prev, retryCount: retryCountRef.current }));
          reconnectTimeoutRef.current = setTimeout(() => {
            connect();
          }, reconnectInterval);
        } else {
          setState(prev => ({
            ...prev,
            error: 'Max reconnection attempts reached',
          }));
        }
      };

      // Handle status updates
      eventSource.addEventListener('status', (event) => {
        try {
          const data = JSON.parse(event.data) as DashboardSummary;
          updateFromSummary(data);
        } catch (e) {
          console.error('Failed to parse status event:', e);
        }
      });

      // Handle alerts
      eventSource.addEventListener('alert', (event) => {
        try {
          const alert = JSON.parse(event.data) as Alert;
          addAlert(alert);
        } catch (e) {
          console.error('Failed to parse alert event:', e);
        }
      });

      // Handle alerts batch update
      eventSource.addEventListener('alerts', (event) => {
        try {
          const alerts = JSON.parse(event.data) as Alert[];
          setAlerts(alerts);
        } catch (e) {
          console.error('Failed to parse alerts event:', e);
        }
      });

      // Handle tool execution results
      eventSource.addEventListener('execution', (event) => {
        try {
          const result = JSON.parse(event.data);
          // Dispatch custom event for components to listen to
          window.dispatchEvent(new CustomEvent('tool-execution', { detail: result }));
        } catch (e) {
          console.error('Failed to parse execution event:', e);
        }
      });

    } catch (error) {
      setState(prev => ({
        ...prev,
        isConnecting: false,
        error: error instanceof Error ? error.message : 'Failed to connect',
      }));
    }
  }, [url, onOpen, onError, reconnectInterval, maxRetries, updateFromSummary, setAlerts, addAlert]);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }
    setState({
      isConnected: false,
      isConnecting: false,
      error: null,
      retryCount: 0,
    });
  }, []);

  useEffect(() => {
    connect();
    return () => {
      disconnect();
    };
  }, [connect, disconnect]);

  return {
    ...state,
    connect,
    disconnect,
    reconnect: () => {
      disconnect();
      retryCountRef.current = 0;
      connect();
    },
  };
}

// Hook for using SSE in the dashboard
export function useDashboardSSE() {
  const apiBase = import.meta.env.VITE_API_URL || '';
  return useSSE({
    url: `${apiBase}/api/v1/sse`,
    onOpen: () => console.log('Dashboard SSE connected'),
    onError: () => console.warn('Dashboard SSE connection error'),
  });
}
