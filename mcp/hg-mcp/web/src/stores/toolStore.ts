import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { ToolDefinition, CategoryInfo, ExecutionRecord } from '../api/types';

interface ToolState {
  tools: ToolDefinition[];
  categories: CategoryInfo[];
  isLoading: boolean;
  error: string | null;

  // Filters
  searchQuery: string;
  selectedCategory: string | null;
  selectedSubcategory: string | null;
  showWriteOnly: boolean;
  showReadOnly: boolean;
  selectedComplexity: string | null;

  // User preferences (persisted)
  favorites: string[];
  aliases: Record<string, string>;
  recentTools: string[];
  executionHistory: ExecutionRecord[];

  // Actions
  setTools: (tools: ToolDefinition[]) => void;
  setCategories: (categories: CategoryInfo[]) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;

  setSearchQuery: (query: string) => void;
  setSelectedCategory: (category: string | null) => void;
  setSelectedSubcategory: (subcategory: string | null) => void;
  setShowWriteOnly: (show: boolean) => void;
  setShowReadOnly: (show: boolean) => void;
  setSelectedComplexity: (complexity: string | null) => void;

  toggleFavorite: (toolName: string) => void;
  isFavorite: (toolName: string) => boolean;
  clearFavorites: () => void;
  setAlias: (alias: string, toolName: string) => void;
  setAliases: (aliases: Record<string, string>) => void;
  removeAlias: (alias: string) => void;
  addRecentTool: (toolName: string) => void;
  clearRecentTools: () => void;
  addExecution: (record: ExecutionRecord) => void;

  // Computed
  getFilteredTools: () => ToolDefinition[];
}

export const useToolStore = create<ToolState>()(
  persist(
    (set, get) => ({
      tools: [],
      categories: [],
      isLoading: false,
      error: null,

      searchQuery: '',
      selectedCategory: null,
      selectedSubcategory: null,
      showWriteOnly: false,
      showReadOnly: false,
      selectedComplexity: null,

      favorites: [],
      aliases: {},
      recentTools: [],
      executionHistory: [],

      setTools: (tools) => set({ tools }),
      setCategories: (categories) => set({ categories }),
      setLoading: (isLoading) => set({ isLoading }),
      setError: (error) => set({ error }),

      setSearchQuery: (searchQuery) => set({ searchQuery }),
      setSelectedCategory: (selectedCategory) =>
        set({ selectedCategory, selectedSubcategory: null }),
      setSelectedSubcategory: (selectedSubcategory) => set({ selectedSubcategory }),
      setShowWriteOnly: (showWriteOnly) => set({ showWriteOnly }),
      setShowReadOnly: (showReadOnly) => set({ showReadOnly }),
      setSelectedComplexity: (selectedComplexity) => set({ selectedComplexity }),

      toggleFavorite: (toolName) => {
        const { favorites } = get();
        if (favorites.includes(toolName)) {
          set({ favorites: favorites.filter((f) => f !== toolName) });
        } else {
          set({ favorites: [...favorites, toolName] });
        }
      },

      isFavorite: (toolName) => get().favorites.includes(toolName),

      clearFavorites: () => set({ favorites: [] }),

      setAlias: (alias, toolName) => {
        const { aliases } = get();
        set({ aliases: { ...aliases, [alias]: toolName } });
      },

      setAliases: (aliases) => set({ aliases }),

      removeAlias: (alias) => {
        const { aliases } = get();
        const newAliases = { ...aliases };
        delete newAliases[alias];
        set({ aliases: newAliases });
      },

      addRecentTool: (toolName) => {
        const { recentTools } = get();
        const filtered = recentTools.filter((t) => t !== toolName);
        set({ recentTools: [toolName, ...filtered].slice(0, 20) });
      },

      clearRecentTools: () => set({ recentTools: [] }),

      addExecution: (record) => {
        const { executionHistory } = get();
        set({ executionHistory: [record, ...executionHistory].slice(0, 100) });
      },

      getFilteredTools: () => {
        const {
          tools,
          searchQuery,
          selectedCategory,
          selectedSubcategory,
          showWriteOnly,
          showReadOnly,
          selectedComplexity,
        } = get();

        return tools.filter((tool) => {
          // Search filter
          if (searchQuery) {
            const query = searchQuery.toLowerCase();
            const matchesSearch =
              tool.name.toLowerCase().includes(query) ||
              tool.description.toLowerCase().includes(query) ||
              tool.tags.some((t) => t.toLowerCase().includes(query));
            if (!matchesSearch) return false;
          }

          // Category filter
          if (selectedCategory && tool.category !== selectedCategory) {
            return false;
          }

          // Subcategory filter
          if (selectedSubcategory && tool.subcategory !== selectedSubcategory) {
            return false;
          }

          // Write/Read filter
          if (showWriteOnly && !tool.isWrite) return false;
          if (showReadOnly && tool.isWrite) return false;

          // Complexity filter
          if (selectedComplexity && tool.complexity !== selectedComplexity) {
            return false;
          }

          return true;
        });
      },
    }),
    {
      name: 'aftrs-tool-preferences',
      partialize: (state) => ({
        favorites: state.favorites,
        aliases: state.aliases,
        recentTools: state.recentTools,
        executionHistory: state.executionHistory,
      }),
    }
  )
);
