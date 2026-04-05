import { useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import {
  Package,
  DollarSign,
  Tag,
  MapPin,
  TrendingUp,
  AlertCircle,
  Clock,
  ChevronRight,
  RefreshCw,
  BarChart3,
  ShoppingCart,
  Eye,
  CheckCircle,
  XCircle,
} from 'lucide-react';
import { api } from '../api/client';
import { useInventoryStore } from '../stores/inventoryStore';
import { cn } from '../lib/utils';
import type { ListingStatus } from '../api/types';

export function InventoryDashboard() {
  const { setSummary, setCategories, setLocations, setLoading, setError } =
    useInventoryStore();

  const {
    data: summary,
    isLoading: summaryLoading,
    error: summaryError,
    refetch: refetchSummary,
  } = useQuery({
    queryKey: ['inventory-summary'],
    queryFn: () => api.getInventorySummary(),
    refetchInterval: 60000,
  });

  const {
    data: categories,
    isLoading: categoriesLoading,
    error: categoriesError,
  } = useQuery({
    queryKey: ['inventory-categories'],
    queryFn: () => api.getInventoryCategories(),
  });

  const { data: locations, error: locationsError } = useQuery({
    queryKey: ['inventory-locations'],
    queryFn: () => api.getInventoryLocations(),
  });

  const { data: recentItems, error: recentError } = useQuery({
    queryKey: ['inventory-recent'],
    queryFn: () => api.getInventoryItems({ limit: 5 }),
  });

  useEffect(() => {
    if (summary) setSummary(summary);
  }, [summary, setSummary]);

  useEffect(() => {
    if (categories) setCategories(categories);
  }, [categories, setCategories]);

  useEffect(() => {
    if (locations) setLocations(locations);
  }, [locations, setLocations]);

  useEffect(() => {
    setLoading(summaryLoading || categoriesLoading);
    if (summaryError) {
      setError(
        summaryError instanceof Error ? summaryError.message : 'Unknown error'
      );
    } else {
      setError(null);
    }
  }, [summaryLoading, categoriesLoading, summaryError, setLoading, setError]);

  const formatCurrency = (value: number) =>
    new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(value);

  const getStatusIcon = (status: ListingStatus) => {
    switch (status) {
      case 'listed':
        return <ShoppingCart className="w-3 h-3 text-green-400" />;
      case 'sold':
        return <CheckCircle className="w-3 h-3 text-blue-400" />;
      case 'pending_review':
        return <Eye className="w-3 h-3 text-yellow-400" />;
      case 'keeping':
        return <Tag className="w-3 h-3 text-purple-400" />;
      default:
        return <XCircle className="w-3 h-3 text-muted-foreground" />;
    }
  };

  const profitMargin =
    summary && summary.total_cost > 0
      ? ((summary.total_value - summary.total_cost) / summary.total_cost) * 100
      : 0;

  return (
    <div className="space-y-4">
      {/* Header Row */}
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-4">
          <h1 className="text-xl font-bold flex items-center gap-2">
            <Package className="w-6 h-6" />
            Inventory Dashboard
          </h1>
        </div>
        <div className="flex items-center gap-2">
          <Link
            to="/inventory/list"
            className="flex items-center gap-1 px-3 py-1.5 rounded-md bg-secondary hover:bg-secondary/80 text-sm"
          >
            View All Items
            <ChevronRight className="w-4 h-4" />
          </Link>
          <button
            onClick={() => refetchSummary()}
            className="flex items-center gap-1 px-3 py-1.5 rounded-md bg-primary text-primary-foreground hover:bg-primary/90 text-sm"
          >
            <RefreshCw className="w-3 h-3" />
            Refresh
          </button>
        </div>
      </div>

      {/* Error Banner */}
      {(summaryError || categoriesError || locationsError || recentError) && (
        <div className="flex items-center gap-2 px-4 py-3 rounded-lg bg-destructive/10 text-destructive border border-destructive/20">
          <AlertCircle className="w-4 h-4 shrink-0" />
          <span className="text-sm">
            Failed to load{' '}
            {[
              summaryError && 'summary',
              categoriesError && 'categories',
              locationsError && 'locations',
              recentError && 'recent items',
            ]
              .filter(Boolean)
              .join(', ')}
          </span>
          <button
            onClick={() => refetchSummary()}
            className="ml-auto text-xs underline"
          >
            Retry
          </button>
        </div>
      )}

      {/* Main Stats Bar */}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
        <StatCard
          label="Total Items"
          value={summary?.total_items ?? 0}
          icon={<Package className="w-4 h-4" />}
          color="blue"
        />
        <StatCard
          label="Total Value"
          value={formatCurrency(summary?.total_value ?? 0)}
          icon={<DollarSign className="w-4 h-4" />}
          color="green"
        />
        <StatCard
          label="Total Cost"
          value={formatCurrency(summary?.total_cost ?? 0)}
          icon={<TrendingUp className="w-4 h-4" />}
          color="purple"
        />
        <StatCard
          label="Profit Margin"
          value={`${profitMargin.toFixed(1)}%`}
          icon={<BarChart3 className="w-4 h-4" />}
          color={profitMargin > 0 ? 'green' : 'red'}
        />
        <StatCard
          label="Categories"
          value={categories?.length ?? 0}
          icon={<Tag className="w-4 h-4" />}
          color="orange"
        />
        <StatCard
          label="Locations"
          value={locations?.length ?? 0}
          icon={<MapPin className="w-4 h-4" />}
          color="cyan"
        />
      </div>

      {/* Status Overview */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {/* Status Breakdown */}
        <div className="rounded-lg border border-border bg-card p-4">
          <h2 className="text-sm font-semibold mb-3 flex items-center gap-2">
            <BarChart3 className="w-4 h-4" />
            Status Breakdown
          </h2>
          <div className="space-y-2">
            {summary?.by_status &&
              Object.entries(summary.by_status).map(([status, count]) => (
                <StatusBar
                  key={status}
                  status={status as ListingStatus}
                  count={count}
                  total={summary.total_items}
                />
              ))}
          </div>
        </div>

        {/* Condition Breakdown */}
        <div className="rounded-lg border border-border bg-card p-4">
          <h2 className="text-sm font-semibold mb-3 flex items-center gap-2">
            <Tag className="w-4 h-4" />
            Condition Overview
          </h2>
          <div className="grid grid-cols-2 sm:grid-cols-3 gap-2">
            {summary?.by_condition &&
              Object.entries(summary.by_condition).map(([condition, count]) => (
                <div
                  key={condition}
                  className="px-3 py-2 rounded-md bg-secondary"
                >
                  <div className="text-xs text-muted-foreground capitalize">
                    {condition.replace('_', ' ')}
                  </div>
                  <div className="text-lg font-semibold">{count}</div>
                </div>
              ))}
          </div>
        </div>
      </div>

      {/* Categories and Recent Items */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        {/* Categories */}
        <div className="lg:col-span-2 rounded-lg border border-border bg-card p-4">
          <div className="flex items-center justify-between mb-3">
            <h2 className="text-sm font-semibold flex items-center gap-2">
              <Tag className="w-4 h-4" />
              Categories
            </h2>
            <span className="text-xs text-muted-foreground">
              {categories?.length ?? 0} categories
            </span>
          </div>
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-2">
            {categories?.slice(0, 12).map((cat) => (
              <Link
                key={cat.name}
                to={`/inventory/list?category=${encodeURIComponent(cat.name)}`}
                className="group flex items-center justify-between px-3 py-2 rounded-md bg-secondary hover:bg-secondary/80 transition-colors"
              >
                <span className="text-sm truncate">{cat.name}</span>
                <div className="flex items-center gap-1">
                  <span className="text-xs text-muted-foreground">
                    {cat.item_count}
                  </span>
                  <ChevronRight className="w-3 h-3 text-muted-foreground group-hover:text-foreground" />
                </div>
              </Link>
            ))}
          </div>
          {categories && categories.length > 12 && (
            <Link
              to="/inventory/list"
              className="block text-center text-xs text-muted-foreground hover:text-foreground mt-3"
            >
              +{categories.length - 12} more categories
            </Link>
          )}
        </div>

        {/* Recent Items */}
        <div className="rounded-lg border border-border bg-card p-4">
          <div className="flex items-center justify-between mb-3">
            <h2 className="text-sm font-semibold flex items-center gap-2">
              <Clock className="w-4 h-4" />
              Recently Added
            </h2>
          </div>
          <div className="space-y-2">
            {recentItems?.items.slice(0, 5).map((item) => (
              <Link
                key={item.sku}
                to={`/inventory/${encodeURIComponent(item.sku)}`}
                className="group flex items-center gap-2 px-2 py-1.5 rounded-md hover:bg-secondary transition-colors"
              >
                {item.primary_image ? (
                  <img
                    src={item.primary_image}
                    alt={item.name}
                    className="w-8 h-8 rounded object-cover bg-secondary"
                  />
                ) : (
                  <div className="w-8 h-8 rounded bg-secondary flex items-center justify-center">
                    <Package className="w-4 h-4 text-muted-foreground" />
                  </div>
                )}
                <div className="flex-1 min-w-0">
                  <div className="text-sm font-medium truncate group-hover:text-primary">
                    {item.name}
                  </div>
                  <div className="flex items-center gap-2 text-xs text-muted-foreground">
                    {getStatusIcon(item.listing_status)}
                    <span className="capitalize">
                      {item.listing_status.replace('_', ' ')}
                    </span>
                  </div>
                </div>
                <div className="text-sm font-medium text-green-400">
                  {formatCurrency(item.purchase_price)}
                </div>
              </Link>
            ))}
            {(!recentItems || recentItems.items.length === 0) && (
              <div className="text-sm text-muted-foreground text-center py-4">
                No items found
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Locations Overview */}
      <div className="rounded-lg border border-border bg-card p-4">
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-sm font-semibold flex items-center gap-2">
            <MapPin className="w-4 h-4" />
            Items by Location
          </h2>
        </div>
        <div className="flex flex-wrap gap-2">
          {summary?.by_location &&
            Object.entries(summary.by_location).map(([location, count]) => (
              <Link
                key={location}
                to={`/inventory/list?location=${encodeURIComponent(location)}`}
                className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-secondary hover:bg-secondary/80 text-sm transition-colors"
              >
                <MapPin className="w-3 h-3" />
                <span>{location}</span>
                <span className="text-xs text-muted-foreground bg-background px-1.5 rounded-full">
                  {count}
                </span>
              </Link>
            ))}
        </div>
      </div>

      {/* Actions Needed */}
      {summary?.by_status?.pending_review &&
        summary.by_status.pending_review > 0 && (
          <div className="rounded-md border border-yellow-500/30 bg-yellow-500/5 p-3">
            <div className="flex items-center gap-2 text-sm text-yellow-400 font-medium mb-2">
              <AlertCircle className="w-4 h-4" />
              Actions Needed
            </div>
            <div className="flex flex-wrap gap-3">
              <Link
                to="/inventory/list?status=pending_review"
                className="flex items-center gap-2 px-3 py-1.5 rounded-md bg-yellow-500/20 text-yellow-400 hover:bg-yellow-500/30 text-sm transition-colors"
              >
                <Eye className="w-4 h-4" />
                {summary.by_status.pending_review} items pending review
                <ChevronRight className="w-4 h-4" />
              </Link>
            </div>
          </div>
        )}
    </div>
  );
}

function StatCard({
  label,
  value,
  icon,
  color,
}: {
  label: string;
  value: string | number;
  icon: React.ReactNode;
  color: string;
}) {
  const colorClasses: Record<string, string> = {
    green: 'text-green-400',
    red: 'text-red-400',
    yellow: 'text-yellow-400',
    orange: 'text-orange-400',
    blue: 'text-blue-400',
    purple: 'text-purple-400',
    cyan: 'text-cyan-400',
    muted: 'text-muted-foreground',
  };

  const bgClasses: Record<string, string> = {
    green: 'bg-green-500/10',
    red: 'bg-red-500/10',
    yellow: 'bg-yellow-500/10',
    orange: 'bg-orange-500/10',
    blue: 'bg-blue-500/10',
    purple: 'bg-purple-500/10',
    cyan: 'bg-cyan-500/10',
    muted: 'bg-muted/10',
  };

  return (
    <div className="flex items-center gap-3 px-4 py-3 rounded-lg border border-border bg-card">
      <div className={cn('p-2 rounded-md', bgClasses[color])}>
        <div className={colorClasses[color]}>{icon}</div>
      </div>
      <div>
        <div className="text-xs text-muted-foreground">{label}</div>
        <div className={cn('text-lg font-semibold', colorClasses[color])}>
          {value}
        </div>
      </div>
    </div>
  );
}

function StatusBar({
  status,
  count,
  total,
}: {
  status: ListingStatus;
  count: number;
  total: number;
}) {
  const percentage = total > 0 ? (count / total) * 100 : 0;

  const statusColors: Record<ListingStatus, string> = {
    not_listed: 'bg-gray-500',
    pending_review: 'bg-yellow-500',
    listed: 'bg-green-500',
    sold: 'bg-blue-500',
    keeping: 'bg-purple-500',
  };

  const statusLabels: Record<ListingStatus, string> = {
    not_listed: 'Not Listed',
    pending_review: 'Pending Review',
    listed: 'Listed',
    sold: 'Sold',
    keeping: 'Keeping',
  };

  return (
    <div className="flex items-center gap-3">
      <div className="w-24 text-xs capitalize truncate">
        {statusLabels[status]}
      </div>
      <div className="flex-1 h-2 rounded-full bg-secondary overflow-hidden">
        <div
          className={cn('h-full rounded-full', statusColors[status])}
          style={{ width: `${percentage}%` }}
        />
      </div>
      <div className="w-12 text-xs text-right text-muted-foreground">
        {count}
      </div>
    </div>
  );
}
