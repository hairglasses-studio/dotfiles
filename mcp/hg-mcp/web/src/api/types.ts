// Tool definitions
export interface ToolParameter {
  name: string;
  type: 'string' | 'number' | 'integer' | 'boolean' | 'array' | 'object';
  description: string;
  required: boolean;
  enum?: string[];
  default?: unknown;
}

export interface ToolDefinition {
  name: string;
  description: string;
  category: string;
  subcategory: string;
  tags: string[];
  useCases: string[];
  complexity: 'simple' | 'moderate' | 'complex';
  isWrite: boolean;
  parameters: ToolParameter[];
}

export interface CategoryInfo {
  name: string;
  count: number;
  subcategories: { name: string; count: number }[];
}

export interface ToolStats {
  totalTools: number;
  moduleCount: number;
  byCategory: Record<string, number>;
  byComplexity: Record<string, number>;
  writeToolsCount: number;
  readOnlyCount: number;
}

// Dashboard types
export interface SystemStatus {
  name: string;
  category: string;
  status: 'online' | 'offline' | 'degraded' | 'unknown';
  latency: number;
  lastCheck: string;
  message?: string;
  critical: boolean;
  healthScore: number;
}

export interface DashboardSummary {
  totalSystems: number;
  onlineCount: number;
  offlineCount: number;
  degradedCount: number;
  unknownCount: number;
  overallHealth: number;
  criticalIssues: number;
  activeAlerts: number;
  lastUpdated: string;
  systems: SystemStatus[];
}

export interface Alert {
  id: string;
  system: string;
  severity: 'info' | 'warning' | 'error' | 'critical';
  message: string;
  timestamp: string;
  resolved: boolean;
}

// Execution types
export interface ToolExecutionRequest {
  toolName: string;
  parameters: Record<string, unknown>;
}

export interface ToolExecutionResult {
  success: boolean;
  data?: unknown;
  error?: string;
  duration: number;
}

export interface ExecutionRecord {
  toolName: string;
  params: Record<string, unknown>;
  result: ToolExecutionResult;
  timestamp: string;
}

// Discovery types
export interface WorkflowStep {
  tool: string;
  args?: Record<string, unknown>;
  onFail?: string;
}

export interface Workflow {
  name: string;
  description: string;
  steps: WorkflowStep[];
}

export interface Preferences {
  favorites: string[];
  aliases: Record<string, string>;
  recentTools: string[];
}

// Workflow builder types
export interface WorkflowStepResult {
  stepIndex: number;
  tool: string;
  success: boolean;
  result?: unknown;
  error?: string;
  duration: number;
}

export interface WorkflowRunResult {
  workflowName: string;
  success: boolean;
  steps: WorkflowStepResult[];
  totalDuration: number;
}

// Inventory types
export type ItemCondition = 'new' | 'like_new' | 'good' | 'fair' | 'poor' | 'for_parts';
export type ListingStatus = 'not_listed' | 'pending_review' | 'listed' | 'sold' | 'keeping';
export type PurchaseSource = 'amazon' | 'newegg' | 'ebay' | 'bestbuy' | 'microcenter' | 'bhphoto' | 'other';

export interface InventoryItem {
  sku: string;
  name: string;
  description?: string;
  category: string;
  subcategory?: string;
  brand?: string;
  model?: string;
  serial_number?: string;
  purchase_price: number;
  purchase_date: string;
  purchase_source: PurchaseSource;
  order_id?: string;
  product_url?: string;
  asin?: string;
  condition: ItemCondition;
  current_value?: number;
  notes?: string;
  location: string;
  quantity: number;
  listing_status: ListingStatus;
  ebay_listing_id?: string;
  ebay_url?: string;
  listed_price?: number;
  sold_price?: number;
  sold_date?: string;
  sold_platform?: string;
  primary_image?: string;
  images?: string[];
  warranty_expiry?: string;
  tags?: string[];
  created_at: string;
  updated_at: string;
}

export interface InventoryFilter {
  category?: string;
  subcategory?: string;
  status?: ListingStatus;
  condition?: ItemCondition;
  location?: string;
  source?: PurchaseSource;
  min_price?: number;
  max_price?: number;
  brand?: string;
  tags?: string[];
  query?: string;
  limit?: number;
  offset?: number;
}

export interface InventorySummary {
  total_items: number;
  total_value: number;
  total_cost: number;
  by_category: Record<string, number>;
  by_status: Record<ListingStatus, number>;
  by_condition: Record<ItemCondition, number>;
  by_location: Record<string, number>;
  recently_added?: InventoryItem[];
  top_value_items?: InventoryItem[];
}

export interface InventoryCategory {
  name: string;
  subcategories?: string[];
  item_count: number;
}

export interface InventoryListResponse {
  items: InventoryItem[];
  count: number;
}
