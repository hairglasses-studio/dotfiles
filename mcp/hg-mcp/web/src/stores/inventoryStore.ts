import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type {
  InventoryItem,
  InventoryFilter,
  InventorySummary,
  InventoryCategory,
} from '../api/types';

type ViewMode = 'table' | 'grid' | 'compact';
type SortDirection = 'asc' | 'desc';

interface InventoryState {
  // Data
  items: InventoryItem[];
  summary: InventorySummary | null;
  categories: InventoryCategory[];
  locations: string[];
  isLoading: boolean;
  error: string | null;

  // Filters
  filters: InventoryFilter;

  // Selection
  selectedSKUs: Set<string>;

  // View preferences (persisted)
  viewMode: ViewMode;
  sortColumn: string;
  sortDirection: SortDirection;
  pageSize: number;
  currentPage: number;

  // Actions - Data
  setItems: (items: InventoryItem[]) => void;
  setSummary: (summary: InventorySummary | null) => void;
  setCategories: (categories: InventoryCategory[]) => void;
  setLocations: (locations: string[]) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;

  // Actions - Filters
  setFilter: <K extends keyof InventoryFilter>(
    key: K,
    value: InventoryFilter[K]
  ) => void;
  setFilters: (filters: Partial<InventoryFilter>) => void;
  clearFilters: () => void;

  // Actions - Selection
  toggleSelect: (sku: string) => void;
  selectAll: () => void;
  clearSelection: () => void;
  isSelected: (sku: string) => boolean;

  // Actions - View
  setViewMode: (mode: ViewMode) => void;
  setSort: (column: string, direction?: SortDirection) => void;
  setPageSize: (size: number) => void;
  setCurrentPage: (page: number) => void;

  // Computed
  getSortedItems: () => InventoryItem[];
  getPagedItems: () => InventoryItem[];
  getTotalPages: () => number;
}

const defaultFilters: InventoryFilter = {
  category: undefined,
  subcategory: undefined,
  status: undefined,
  condition: undefined,
  location: undefined,
  source: undefined,
  min_price: undefined,
  max_price: undefined,
  brand: undefined,
  tags: undefined,
  query: undefined,
  limit: 50,
  offset: 0,
};

export const useInventoryStore = create<InventoryState>()(
  persist(
    (set, get) => ({
      // Initial state
      items: [],
      summary: null,
      categories: [],
      locations: [],
      isLoading: false,
      error: null,

      filters: { ...defaultFilters },
      selectedSKUs: new Set(),

      viewMode: 'table',
      sortColumn: 'updated_at',
      sortDirection: 'desc',
      pageSize: 25,
      currentPage: 1,

      // Data actions
      setItems: (items) => set({ items }),
      setSummary: (summary) => set({ summary }),
      setCategories: (categories) => set({ categories }),
      setLocations: (locations) => set({ locations }),
      setLoading: (isLoading) => set({ isLoading }),
      setError: (error) => set({ error }),

      // Filter actions
      setFilter: (key, value) => {
        const { filters } = get();
        set({
          filters: { ...filters, [key]: value },
          currentPage: 1, // Reset to first page on filter change
        });
      },

      setFilters: (newFilters) => {
        const { filters } = get();
        set({
          filters: { ...filters, ...newFilters },
          currentPage: 1,
        });
      },

      clearFilters: () =>
        set({
          filters: { ...defaultFilters },
          currentPage: 1,
        }),

      // Selection actions
      toggleSelect: (sku) => {
        const { selectedSKUs } = get();
        const newSelection = new Set(selectedSKUs);
        if (newSelection.has(sku)) {
          newSelection.delete(sku);
        } else {
          newSelection.add(sku);
        }
        set({ selectedSKUs: newSelection });
      },

      selectAll: () => {
        const { items } = get();
        set({ selectedSKUs: new Set(items.map((item) => item.sku)) });
      },

      clearSelection: () => set({ selectedSKUs: new Set() }),

      isSelected: (sku) => get().selectedSKUs.has(sku),

      // View actions
      setViewMode: (viewMode) => set({ viewMode }),

      setSort: (column, direction) => {
        const { sortColumn, sortDirection } = get();
        if (column === sortColumn && !direction) {
          // Toggle direction if clicking same column
          set({ sortDirection: sortDirection === 'asc' ? 'desc' : 'asc' });
        } else {
          set({
            sortColumn: column,
            sortDirection: direction || 'asc',
          });
        }
      },

      setPageSize: (pageSize) => set({ pageSize, currentPage: 1 }),

      setCurrentPage: (currentPage) => set({ currentPage }),

      // Computed
      getSortedItems: () => {
        const { items, sortColumn, sortDirection } = get();
        const sorted = [...items].sort((a, b) => {
          const aVal = a[sortColumn as keyof InventoryItem];
          const bVal = b[sortColumn as keyof InventoryItem];

          if (aVal === undefined || aVal === null) return 1;
          if (bVal === undefined || bVal === null) return -1;

          let comparison = 0;
          if (typeof aVal === 'string' && typeof bVal === 'string') {
            comparison = aVal.localeCompare(bVal);
          } else if (typeof aVal === 'number' && typeof bVal === 'number') {
            comparison = aVal - bVal;
          } else {
            comparison = String(aVal).localeCompare(String(bVal));
          }

          return sortDirection === 'asc' ? comparison : -comparison;
        });
        return sorted;
      },

      getPagedItems: () => {
        const { pageSize, currentPage } = get();
        const sorted = get().getSortedItems();
        const start = (currentPage - 1) * pageSize;
        return sorted.slice(start, start + pageSize);
      },

      getTotalPages: () => {
        const { items, pageSize } = get();
        return Math.ceil(items.length / pageSize);
      },
    }),
    {
      name: 'aftrs-inventory-preferences',
      partialize: (state) => ({
        viewMode: state.viewMode,
        sortColumn: state.sortColumn,
        sortDirection: state.sortDirection,
        pageSize: state.pageSize,
      }),
    }
  )
);
