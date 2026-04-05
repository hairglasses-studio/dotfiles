import { useEffect, useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link, useSearchParams } from 'react-router-dom';
import {
  Package,
  Search,
  Filter,
  ChevronDown,
  ChevronUp,
  ChevronLeft,
  ChevronRight,
  Trash2,
  RefreshCw,
  LayoutGrid,
  LayoutList,
  Table2,
  X,
  Check,
  AlertCircle,
  ExternalLink,
} from 'lucide-react';
import { api } from '../api/client';
import { useInventoryStore } from '../stores/inventoryStore';
import { cn } from '../lib/utils';
import type {
  InventoryFilter,
  ListingStatus,
  ItemCondition,
} from '../api/types';

export function InventoryList() {
  const queryClient = useQueryClient();
  const [searchParams] = useSearchParams();

  const {
    items,
    setItems,
    categories,
    setCategories,
    locations,
    setLocations,
    filters,
    setFilter,
    setFilters,
    clearFilters,
    selectedSKUs,
    toggleSelect,
    selectAll,
    clearSelection,
    isSelected,
    viewMode,
    setViewMode,
    sortColumn,
    sortDirection,
    setSort,
    pageSize,
    setPageSize,
    currentPage,
    setCurrentPage,
    getPagedItems,
    getTotalPages,
  } = useInventoryStore();

  const [showFilters, setShowFilters] = useState(false);
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null);

  // Sync URL params to filters on mount
  useEffect(() => {
    const urlFilters: Partial<InventoryFilter> = {};
    const category = searchParams.get('category');
    const status = searchParams.get('status');
    const location = searchParams.get('location');
    const query = searchParams.get('query');

    if (category) urlFilters.category = category;
    if (status) urlFilters.status = status as ListingStatus;
    if (location) urlFilters.location = location;
    if (query) urlFilters.query = query;

    if (Object.keys(urlFilters).length > 0) {
      setFilters(urlFilters);
    }
  }, []);

  // Fetch inventory items
  const {
    data,
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: ['inventory-items', filters],
    queryFn: () => api.getInventoryItems(filters),
  });

  // Fetch categories and locations for filters
  const { data: categoriesData } = useQuery({
    queryKey: ['inventory-categories'],
    queryFn: () => api.getInventoryCategories(),
  });

  const { data: locationsData } = useQuery({
    queryKey: ['inventory-locations'],
    queryFn: () => api.getInventoryLocations(),
  });

  useEffect(() => {
    if (data) setItems(data.items);
  }, [data, setItems]);

  useEffect(() => {
    if (categoriesData) setCategories(categoriesData);
  }, [categoriesData, setCategories]);

  useEffect(() => {
    if (locationsData) setLocations(locationsData);
  }, [locationsData, setLocations]);

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: (sku: string) => api.deleteInventoryItem(sku),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['inventory-items'] });
      queryClient.invalidateQueries({ queryKey: ['inventory-summary'] });
      setDeleteConfirm(null);
    },
  });

  // Bulk delete mutation
  const bulkDeleteMutation = useMutation({
    mutationFn: (skus: string[]) => api.deleteInventoryItems(skus),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['inventory-items'] });
      queryClient.invalidateQueries({ queryKey: ['inventory-summary'] });
      clearSelection();
    },
  });

  const formatCurrency = (value: number) =>
    new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(value);

  const formatDate = (dateStr: string) =>
    new Date(dateStr).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });

  const handleSort = (column: string) => {
    setSort(column);
  };

  const handleFilterChange = (key: keyof InventoryFilter, value: string) => {
    if (value === '') {
      setFilter(key, undefined);
    } else {
      setFilter(key, value as any);
    }
  };

  const handleSearch = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    const query = formData.get('search') as string;
    setFilter('query', query || undefined);
  };

  const handleBulkDelete = () => {
    if (selectedSKUs.size === 0) return;
    bulkDeleteMutation.mutate(Array.from(selectedSKUs));
  };

  const pagedItems = getPagedItems();
  const totalPages = getTotalPages();
  const hasActiveFilters =
    filters.category ||
    filters.status ||
    filters.condition ||
    filters.location ||
    filters.query;

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-4">
          <Link
            to="/inventory"
            className="flex items-center gap-1 text-muted-foreground hover:text-foreground"
          >
            <ChevronLeft className="w-4 h-4" />
            Dashboard
          </Link>
          <h1 className="text-xl font-bold flex items-center gap-2">
            <Package className="w-6 h-6" />
            Inventory Items
          </h1>
          <span className="text-sm text-muted-foreground">
            {data?.count ?? 0} items
          </span>
        </div>
        <div className="flex items-center gap-2">
          {/* View Mode Toggle */}
          <div className="flex rounded-md border border-border overflow-hidden">
            <button
              onClick={() => setViewMode('table')}
              className={cn(
                'p-1.5',
                viewMode === 'table'
                  ? 'bg-primary text-primary-foreground'
                  : 'bg-card hover:bg-accent'
              )}
              title="Table view"
            >
              <Table2 className="w-4 h-4" />
            </button>
            <button
              onClick={() => setViewMode('grid')}
              className={cn(
                'p-1.5',
                viewMode === 'grid'
                  ? 'bg-primary text-primary-foreground'
                  : 'bg-card hover:bg-accent'
              )}
              title="Grid view"
            >
              <LayoutGrid className="w-4 h-4" />
            </button>
            <button
              onClick={() => setViewMode('compact')}
              className={cn(
                'p-1.5',
                viewMode === 'compact'
                  ? 'bg-primary text-primary-foreground'
                  : 'bg-card hover:bg-accent'
              )}
              title="Compact view"
            >
              <LayoutList className="w-4 h-4" />
            </button>
          </div>
          <button
            onClick={() => refetch()}
            className="flex items-center gap-1 px-3 py-1.5 rounded-md bg-primary text-primary-foreground hover:bg-primary/90 text-sm"
          >
            <RefreshCw className="w-3 h-3" />
            Refresh
          </button>
        </div>
      </div>

      {/* Search and Filters Bar */}
      <div className="flex flex-wrap items-center gap-2">
        {/* Search */}
        <form onSubmit={handleSearch} className="flex-1 min-w-[200px]">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            <input
              type="text"
              name="search"
              placeholder="Search items..."
              defaultValue={filters.query || ''}
              className="w-full pl-9 pr-4 py-2 rounded-md border border-border bg-card text-sm focus:outline-none focus:ring-2 focus:ring-primary"
            />
          </div>
        </form>

        {/* Filter Toggle */}
        <button
          onClick={() => setShowFilters(!showFilters)}
          className={cn(
            'flex items-center gap-1 px-3 py-2 rounded-md border text-sm',
            hasActiveFilters
              ? 'border-primary bg-primary/10 text-primary'
              : 'border-border bg-card hover:bg-accent'
          )}
        >
          <Filter className="w-4 h-4" />
          Filters
          {hasActiveFilters && (
            <span className="ml-1 px-1.5 rounded-full bg-primary text-primary-foreground text-xs">
              {
                [
                  filters.category,
                  filters.status,
                  filters.condition,
                  filters.location,
                ].filter(Boolean).length
              }
            </span>
          )}
        </button>

        {/* Clear Filters */}
        {hasActiveFilters && (
          <button
            onClick={clearFilters}
            className="flex items-center gap-1 px-3 py-2 rounded-md border border-border bg-card hover:bg-accent text-sm"
          >
            <X className="w-4 h-4" />
            Clear
          </button>
        )}

        {/* Bulk Actions */}
        {selectedSKUs.size > 0 && (
          <div className="flex items-center gap-2 ml-auto">
            <span className="text-sm text-muted-foreground">
              {selectedSKUs.size} selected
            </span>
            <button
              onClick={() => clearSelection()}
              className="px-2 py-1 text-xs rounded-md border border-border hover:bg-accent"
            >
              Deselect
            </button>
            <button
              onClick={handleBulkDelete}
              disabled={bulkDeleteMutation.isPending}
              className="flex items-center gap-1 px-2 py-1 text-xs rounded-md bg-red-500/20 text-red-400 hover:bg-red-500/30"
            >
              <Trash2 className="w-3 h-3" />
              Delete Selected
            </button>
          </div>
        )}
      </div>

      {/* Filter Panel */}
      {showFilters && (
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 p-3 rounded-lg border border-border bg-card">
          {/* Category */}
          <div>
            <label className="text-xs text-muted-foreground mb-1 block">
              Category
            </label>
            <select
              value={filters.category || ''}
              onChange={(e) => handleFilterChange('category', e.target.value)}
              className="w-full px-2 py-1.5 rounded-md border border-border bg-background text-sm"
            >
              <option value="">All Categories</option>
              {categories.map((cat) => (
                <option key={cat.name} value={cat.name}>
                  {cat.name} ({cat.item_count})
                </option>
              ))}
            </select>
          </div>

          {/* Status */}
          <div>
            <label className="text-xs text-muted-foreground mb-1 block">
              Status
            </label>
            <select
              value={filters.status || ''}
              onChange={(e) => handleFilterChange('status', e.target.value)}
              className="w-full px-2 py-1.5 rounded-md border border-border bg-background text-sm"
            >
              <option value="">All Statuses</option>
              <option value="not_listed">Not Listed</option>
              <option value="pending_review">Pending Review</option>
              <option value="listed">Listed</option>
              <option value="sold">Sold</option>
              <option value="keeping">Keeping</option>
            </select>
          </div>

          {/* Condition */}
          <div>
            <label className="text-xs text-muted-foreground mb-1 block">
              Condition
            </label>
            <select
              value={filters.condition || ''}
              onChange={(e) => handleFilterChange('condition', e.target.value)}
              className="w-full px-2 py-1.5 rounded-md border border-border bg-background text-sm"
            >
              <option value="">All Conditions</option>
              <option value="new">New</option>
              <option value="like_new">Like New</option>
              <option value="good">Good</option>
              <option value="fair">Fair</option>
              <option value="poor">Poor</option>
              <option value="for_parts">For Parts</option>
            </select>
          </div>

          {/* Location */}
          <div>
            <label className="text-xs text-muted-foreground mb-1 block">
              Location
            </label>
            <select
              value={filters.location || ''}
              onChange={(e) => handleFilterChange('location', e.target.value)}
              className="w-full px-2 py-1.5 rounded-md border border-border bg-background text-sm"
            >
              <option value="">All Locations</option>
              {locations.map((loc) => (
                <option key={loc} value={loc}>
                  {loc}
                </option>
              ))}
            </select>
          </div>
        </div>
      )}

      {/* Loading/Error States */}
      {isLoading && (
        <div className="flex items-center justify-center py-12">
          <RefreshCw className="w-6 h-6 animate-spin text-muted-foreground" />
        </div>
      )}

      {error && (
        <div className="flex items-center gap-2 p-4 rounded-lg border border-red-500/30 bg-red-500/5 text-red-400">
          <AlertCircle className="w-5 h-5" />
          <span>
            {error instanceof Error ? error.message : 'Failed to load items'}
          </span>
        </div>
      )}

      {/* Table View */}
      {!isLoading && !error && viewMode === 'table' && (
        <div className="rounded-lg border border-border bg-card overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-secondary/50">
                <tr>
                  <th className="w-8 px-3 py-2">
                    <input
                      type="checkbox"
                      checked={
                        selectedSKUs.size === items.length && items.length > 0
                      }
                      onChange={() =>
                        selectedSKUs.size === items.length
                          ? clearSelection()
                          : selectAll()
                      }
                      className="rounded border-border"
                    />
                  </th>
                  <SortHeader
                    label="SKU"
                    column="sku"
                    current={sortColumn}
                    direction={sortDirection}
                    onSort={handleSort}
                  />
                  <SortHeader
                    label="Name"
                    column="name"
                    current={sortColumn}
                    direction={sortDirection}
                    onSort={handleSort}
                  />
                  <SortHeader
                    label="Category"
                    column="category"
                    current={sortColumn}
                    direction={sortDirection}
                    onSort={handleSort}
                  />
                  <SortHeader
                    label="Price"
                    column="purchase_price"
                    current={sortColumn}
                    direction={sortDirection}
                    onSort={handleSort}
                  />
                  <SortHeader
                    label="Condition"
                    column="condition"
                    current={sortColumn}
                    direction={sortDirection}
                    onSort={handleSort}
                  />
                  <SortHeader
                    label="Status"
                    column="listing_status"
                    current={sortColumn}
                    direction={sortDirection}
                    onSort={handleSort}
                  />
                  <SortHeader
                    label="Location"
                    column="location"
                    current={sortColumn}
                    direction={sortDirection}
                    onSort={handleSort}
                  />
                  <SortHeader
                    label="Updated"
                    column="updated_at"
                    current={sortColumn}
                    direction={sortDirection}
                    onSort={handleSort}
                  />
                  <th className="w-16 px-3 py-2 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {pagedItems.map((item) => (
                  <tr
                    key={item.sku}
                    className={cn(
                      'hover:bg-accent/50 transition-colors',
                      isSelected(item.sku) && 'bg-primary/10'
                    )}
                  >
                    <td className="px-3 py-2">
                      <input
                        type="checkbox"
                        checked={isSelected(item.sku)}
                        onChange={() => toggleSelect(item.sku)}
                        className="rounded border-border"
                      />
                    </td>
                    <td className="px-3 py-2 font-mono text-xs">{item.sku}</td>
                    <td className="px-3 py-2">
                      <Link
                        to={`/inventory/${encodeURIComponent(item.sku)}`}
                        className="font-medium hover:text-primary"
                      >
                        {item.name}
                      </Link>
                    </td>
                    <td className="px-3 py-2 text-muted-foreground">
                      {item.category}
                    </td>
                    <td className="px-3 py-2 font-medium">
                      {formatCurrency(item.purchase_price)}
                    </td>
                    <td className="px-3 py-2">
                      <ConditionBadge condition={item.condition} />
                    </td>
                    <td className="px-3 py-2">
                      <StatusBadge status={item.listing_status} />
                    </td>
                    <td className="px-3 py-2 text-muted-foreground">
                      {item.location}
                    </td>
                    <td className="px-3 py-2 text-muted-foreground text-xs">
                      {formatDate(item.updated_at)}
                    </td>
                    <td className="px-3 py-2">
                      <div className="flex items-center justify-end gap-1">
                        <Link
                          to={`/inventory/${encodeURIComponent(item.sku)}`}
                          className="p-1 rounded hover:bg-secondary"
                          title="View details"
                        >
                          <ExternalLink className="w-4 h-4" />
                        </Link>
                        {deleteConfirm === item.sku ? (
                          <>
                            <button
                              onClick={() => deleteMutation.mutate(item.sku)}
                              className="p-1 rounded bg-red-500/20 text-red-400 hover:bg-red-500/30"
                              title="Confirm delete"
                            >
                              <Check className="w-4 h-4" />
                            </button>
                            <button
                              onClick={() => setDeleteConfirm(null)}
                              className="p-1 rounded hover:bg-secondary"
                              title="Cancel"
                            >
                              <X className="w-4 h-4" />
                            </button>
                          </>
                        ) : (
                          <button
                            onClick={() => setDeleteConfirm(item.sku)}
                            className="p-1 rounded hover:bg-secondary text-muted-foreground hover:text-red-400"
                            title="Delete"
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Grid View */}
      {!isLoading && !error && viewMode === 'grid' && (
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-3">
          {pagedItems.map((item) => (
            <Link
              key={item.sku}
              to={`/inventory/${encodeURIComponent(item.sku)}`}
              className="group rounded-lg border border-border bg-card overflow-hidden hover:border-primary transition-colors"
            >
              <div className="aspect-square bg-secondary relative">
                {item.primary_image ? (
                  <img
                    src={item.primary_image}
                    alt={item.name}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <div className="w-full h-full flex items-center justify-center">
                    <Package className="w-12 h-12 text-muted-foreground" />
                  </div>
                )}
                <div className="absolute top-2 right-2">
                  <StatusBadge status={item.listing_status} />
                </div>
              </div>
              <div className="p-3">
                <div className="text-sm font-medium truncate group-hover:text-primary">
                  {item.name}
                </div>
                <div className="text-xs text-muted-foreground truncate">
                  {item.category}
                </div>
                <div className="flex items-center justify-between mt-2">
                  <span className="text-sm font-semibold text-green-400">
                    {formatCurrency(item.purchase_price)}
                  </span>
                  <ConditionBadge condition={item.condition} />
                </div>
              </div>
            </Link>
          ))}
        </div>
      )}

      {/* Compact View */}
      {!isLoading && !error && viewMode === 'compact' && (
        <div className="space-y-1">
          {pagedItems.map((item) => (
            <div
              key={item.sku}
              className={cn(
                'flex items-center gap-3 px-3 py-2 rounded-md border border-border bg-card hover:border-primary transition-colors',
                isSelected(item.sku) && 'bg-primary/10 border-primary'
              )}
            >
              <input
                type="checkbox"
                checked={isSelected(item.sku)}
                onChange={() => toggleSelect(item.sku)}
                className="rounded border-border"
              />
              {item.primary_image ? (
                <img
                  src={item.primary_image}
                  alt={item.name}
                  className="w-10 h-10 rounded object-cover bg-secondary"
                />
              ) : (
                <div className="w-10 h-10 rounded bg-secondary flex items-center justify-center">
                  <Package className="w-5 h-5 text-muted-foreground" />
                </div>
              )}
              <div className="flex-1 min-w-0">
                <Link
                  to={`/inventory/${encodeURIComponent(item.sku)}`}
                  className="text-sm font-medium hover:text-primary truncate block"
                >
                  {item.name}
                </Link>
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  <span>{item.category}</span>
                  <span>•</span>
                  <span>{item.location}</span>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <ConditionBadge condition={item.condition} />
                <StatusBadge status={item.listing_status} />
                <span className="text-sm font-medium w-20 text-right">
                  {formatCurrency(item.purchase_price)}
                </span>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Empty State */}
      {!isLoading && !error && items.length === 0 && (
        <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
          <Package className="w-12 h-12 mb-4" />
          <p>No items found</p>
          {hasActiveFilters && (
            <button
              onClick={clearFilters}
              className="mt-2 text-sm text-primary hover:underline"
            >
              Clear filters
            </button>
          )}
        </div>
      )}

      {/* Pagination */}
      {!isLoading && !error && items.length > 0 && (
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <span>Show</span>
            <select
              value={pageSize}
              onChange={(e) => setPageSize(Number(e.target.value))}
              className="px-2 py-1 rounded-md border border-border bg-card"
            >
              <option value={10}>10</option>
              <option value={25}>25</option>
              <option value={50}>50</option>
              <option value={100}>100</option>
            </select>
            <span>per page</span>
          </div>

          <div className="flex items-center gap-1">
            <button
              onClick={() => setCurrentPage(1)}
              disabled={currentPage === 1}
              className="p-1.5 rounded-md border border-border bg-card hover:bg-accent disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <ChevronLeft className="w-4 h-4" />
              <ChevronLeft className="w-4 h-4 -ml-3" />
            </button>
            <button
              onClick={() => setCurrentPage(currentPage - 1)}
              disabled={currentPage === 1}
              className="p-1.5 rounded-md border border-border bg-card hover:bg-accent disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <ChevronLeft className="w-4 h-4" />
            </button>

            <span className="px-3 text-sm">
              Page {currentPage} of {totalPages}
            </span>

            <button
              onClick={() => setCurrentPage(currentPage + 1)}
              disabled={currentPage === totalPages}
              className="p-1.5 rounded-md border border-border bg-card hover:bg-accent disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <ChevronRight className="w-4 h-4" />
            </button>
            <button
              onClick={() => setCurrentPage(totalPages)}
              disabled={currentPage === totalPages}
              className="p-1.5 rounded-md border border-border bg-card hover:bg-accent disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <ChevronRight className="w-4 h-4" />
              <ChevronRight className="w-4 h-4 -ml-3" />
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

function SortHeader({
  label,
  column,
  current,
  direction,
  onSort,
}: {
  label: string;
  column: string;
  current: string;
  direction: 'asc' | 'desc';
  onSort: (column: string) => void;
}) {
  const isActive = current === column;

  return (
    <th className="px-3 py-2 text-left">
      <button
        onClick={() => onSort(column)}
        className={cn(
          'flex items-center gap-1 hover:text-foreground',
          isActive ? 'text-foreground' : 'text-muted-foreground'
        )}
      >
        {label}
        {isActive &&
          (direction === 'asc' ? (
            <ChevronUp className="w-3 h-3" />
          ) : (
            <ChevronDown className="w-3 h-3" />
          ))}
      </button>
    </th>
  );
}

function StatusBadge({ status }: { status: ListingStatus }) {
  const colors: Record<ListingStatus, string> = {
    not_listed: 'bg-gray-500/20 text-gray-400',
    pending_review: 'bg-yellow-500/20 text-yellow-400',
    listed: 'bg-green-500/20 text-green-400',
    sold: 'bg-blue-500/20 text-blue-400',
    keeping: 'bg-purple-500/20 text-purple-400',
  };

  const labels: Record<ListingStatus, string> = {
    not_listed: 'Not Listed',
    pending_review: 'Review',
    listed: 'Listed',
    sold: 'Sold',
    keeping: 'Keeping',
  };

  return (
    <span className={cn('px-1.5 py-0.5 rounded text-xs', colors[status])}>
      {labels[status]}
    </span>
  );
}

function ConditionBadge({ condition }: { condition: ItemCondition }) {
  const colors: Record<ItemCondition, string> = {
    new: 'bg-green-500/20 text-green-400',
    like_new: 'bg-cyan-500/20 text-cyan-400',
    good: 'bg-blue-500/20 text-blue-400',
    fair: 'bg-yellow-500/20 text-yellow-400',
    poor: 'bg-orange-500/20 text-orange-400',
    for_parts: 'bg-red-500/20 text-red-400',
  };

  const labels: Record<ItemCondition, string> = {
    new: 'New',
    like_new: 'Like New',
    good: 'Good',
    fair: 'Fair',
    poor: 'Poor',
    for_parts: 'Parts',
  };

  return (
    <span className={cn('px-1.5 py-0.5 rounded text-xs', colors[condition])}>
      {labels[condition]}
    </span>
  );
}
