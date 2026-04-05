import type {
  ToolDefinition,
  CategoryInfo,
  ToolStats,
  DashboardSummary,
  Alert,
  ToolExecutionResult,
  Workflow,
  Preferences,
  InventoryItem,
  InventoryFilter,
  InventorySummary,
  InventoryCategory,
  InventoryListResponse,
} from './types';

const API_BASE = '/api/v1';

class ApiClient {
  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${API_BASE}${endpoint}`;
    const response = await fetch(url, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(error || `HTTP ${response.status}`);
    }

    return response.json();
  }

  // Tools API
  async getTools(): Promise<ToolDefinition[]> {
    return this.request<ToolDefinition[]>('/tools');
  }

  async getTool(name: string): Promise<ToolDefinition> {
    return this.request<ToolDefinition>(`/tools/${name}`);
  }

  async getCategories(): Promise<CategoryInfo[]> {
    return this.request<CategoryInfo[]>('/categories');
  }

  async getStats(): Promise<ToolStats> {
    return this.request<ToolStats>('/stats');
  }

  async executeTool(
    name: string,
    parameters: Record<string, unknown>
  ): Promise<ToolExecutionResult> {
    return this.request<ToolExecutionResult>(`/tools/${name}/execute`, {
      method: 'POST',
      body: JSON.stringify(parameters),
    });
  }

  async searchTools(query: string): Promise<ToolDefinition[]> {
    return this.request<ToolDefinition[]>(`/tools?search=${encodeURIComponent(query)}`);
  }

  // Dashboard API
  async getDashboardStatus(): Promise<DashboardSummary> {
    return this.request<DashboardSummary>('/dashboard/status');
  }

  async getDashboardQuick(): Promise<string> {
    return this.request<string>('/dashboard/quick');
  }

  async getAlerts(): Promise<Alert[]> {
    return this.request<Alert[]>('/dashboard/alerts');
  }

  async startMonitoring(interval?: number): Promise<void> {
    return this.request('/dashboard/monitor/start', {
      method: 'POST',
      body: JSON.stringify({ interval }),
    });
  }

  async stopMonitoring(): Promise<void> {
    return this.request('/dashboard/monitor/stop', { method: 'POST' });
  }

  async resolveAlert(alertId: string): Promise<void> {
    return this.request(`/dashboard/alerts/${alertId}/resolve`, {
      method: 'POST',
    });
  }

  // Workflows API
  async getWorkflows(): Promise<Workflow[]> {
    return this.request<Workflow[]>('/workflows');
  }

  async getWorkflow(name: string): Promise<Workflow> {
    return this.request<Workflow>(`/workflows/${name}`);
  }

  async createWorkflow(workflow: Workflow): Promise<void> {
    return this.request('/workflows', {
      method: 'POST',
      body: JSON.stringify(workflow),
    });
  }

  async updateWorkflow(name: string, workflow: Workflow): Promise<void> {
    return this.request(`/workflows/${name}`, {
      method: 'PUT',
      body: JSON.stringify(workflow),
    });
  }

  async deleteWorkflow(name: string): Promise<void> {
    return this.request(`/workflows/${name}`, { method: 'DELETE' });
  }

  async runWorkflow(name: string): Promise<ToolExecutionResult[]> {
    return this.request<ToolExecutionResult[]>(`/workflows/${name}/run`, {
      method: 'POST',
    });
  }

  // Preferences API
  async getPreferences(): Promise<Preferences> {
    return this.request<Preferences>('/preferences');
  }

  async getFavorites(): Promise<string[]> {
    return this.request<string[]>('/favorites');
  }

  async addFavorite(toolName: string): Promise<void> {
    return this.request(`/favorites/${toolName}`, { method: 'POST' });
  }

  async removeFavorite(toolName: string): Promise<void> {
    return this.request(`/favorites/${toolName}`, { method: 'DELETE' });
  }

  async getAliases(): Promise<Record<string, string>> {
    return this.request<Record<string, string>>('/aliases');
  }

  async updateAliases(aliases: Record<string, string>): Promise<void> {
    return this.request('/aliases', {
      method: 'PUT',
      body: JSON.stringify(aliases),
    });
  }

  // Inventory API
  async getInventoryItems(
    filters?: InventoryFilter
  ): Promise<InventoryListResponse> {
    const params = new URLSearchParams();
    if (filters) {
      if (filters.category) params.set('category', filters.category);
      if (filters.subcategory) params.set('subcategory', filters.subcategory);
      if (filters.status) params.set('status', filters.status);
      if (filters.condition) params.set('condition', filters.condition);
      if (filters.location) params.set('location', filters.location);
      if (filters.source) params.set('source', filters.source);
      if (filters.brand) params.set('brand', filters.brand);
      if (filters.query) params.set('query', filters.query);
      if (filters.min_price !== undefined)
        params.set('min_price', filters.min_price.toString());
      if (filters.max_price !== undefined)
        params.set('max_price', filters.max_price.toString());
      if (filters.limit !== undefined)
        params.set('limit', filters.limit.toString());
      if (filters.offset !== undefined)
        params.set('offset', filters.offset.toString());
      if (filters.tags && filters.tags.length > 0)
        params.set('tags', filters.tags.join(','));
    }
    const queryString = params.toString();
    const endpoint = queryString ? `/inventory?${queryString}` : '/inventory';
    return this.request<InventoryListResponse>(endpoint);
  }

  async getInventoryItem(sku: string): Promise<InventoryItem> {
    return this.request<InventoryItem>(`/inventory/${encodeURIComponent(sku)}`);
  }

  async createInventoryItem(
    item: Partial<InventoryItem>
  ): Promise<InventoryItem> {
    return this.request<InventoryItem>('/inventory', {
      method: 'POST',
      body: JSON.stringify(item),
    });
  }

  async updateInventoryItem(
    sku: string,
    updates: Partial<InventoryItem>
  ): Promise<InventoryItem> {
    return this.request<InventoryItem>(
      `/inventory/${encodeURIComponent(sku)}`,
      {
        method: 'PUT',
        body: JSON.stringify(updates),
      }
    );
  }

  async deleteInventoryItem(sku: string): Promise<void> {
    return this.request(`/inventory/${encodeURIComponent(sku)}`, {
      method: 'DELETE',
    });
  }

  async deleteInventoryItems(
    skus: string[]
  ): Promise<{ deleted: number; errors: number; deleted_skus: string[]; error_details: string[] }> {
    return this.request('/inventory/bulk-delete', {
      method: 'POST',
      body: JSON.stringify({ skus, confirm: true }),
    });
  }

  async getInventorySummary(): Promise<InventorySummary> {
    return this.request<InventorySummary>('/inventory/summary');
  }

  async getInventoryCategories(): Promise<InventoryCategory[]> {
    return this.request<InventoryCategory[]>('/inventory/categories');
  }

  async getInventoryLocations(): Promise<string[]> {
    return this.request<string[]>('/inventory/locations');
  }

  async getInventoryImages(
    sku: string
  ): Promise<{ sku: string; images: string[]; count: number }> {
    return this.request(`/inventory/${encodeURIComponent(sku)}/images`);
  }
}

export const api = new ApiClient();
