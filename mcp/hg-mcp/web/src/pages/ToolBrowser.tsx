import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import {
  Search,
  Star,
  ChevronRight,
  ChevronDown,
  Filter,
  BookOpen,
  Pencil,
  SlidersHorizontal,
  X,
} from 'lucide-react';
import { api } from '../api/client';
import { useToolStore } from '../stores/toolStore';
import { cn, getComplexityColor } from '../lib/utils';
import type { ToolDefinition, CategoryInfo } from '../api/types';

export function ToolBrowser() {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const {
    tools,
    categories,
    searchQuery,
    selectedCategory,
    selectedSubcategory,
    favorites,
    setTools,
    setCategories,
    setSearchQuery,
    setSelectedCategory,
    setSelectedSubcategory,
    toggleFavorite,
    isFavorite,
    getFilteredTools,
    addRecentTool,
  } = useToolStore();

  const { data: toolsData, isLoading: toolsLoading } = useQuery({
    queryKey: ['tools'],
    queryFn: () => api.getTools(),
  });

  const { data: categoriesData } = useQuery({
    queryKey: ['categories'],
    queryFn: () => api.getCategories(),
  });

  useEffect(() => {
    if (toolsData) {
      setTools(toolsData);
    }
  }, [toolsData, setTools]);

  useEffect(() => {
    if (categoriesData) {
      setCategories(categoriesData);
    }
  }, [categoriesData, setCategories]);

  const filteredTools = getFilteredTools();
  const showFavoritesOnly = searchQuery === '@favorites';
  const displayTools = showFavoritesOnly
    ? tools.filter((t) => favorites.includes(t.name))
    : filteredTools;

  const handleCategorySelect = (cat: string | null, subcat: string | null) => {
    setSelectedCategory(cat);
    setSelectedSubcategory(subcat);
    if (searchQuery === '@favorites') setSearchQuery('');
    setSidebarOpen(false);
  };

  return (
    <div className="flex flex-col lg:flex-row gap-4 lg:gap-6 h-[calc(100vh-7rem)]">
      {/* Mobile filter bar */}
      <div className="lg:hidden flex gap-2">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search tools..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-9 pr-4 py-2.5 rounded-md border border-input bg-card text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
          />
        </div>
        <button
          onClick={() => setSidebarOpen(true)}
          className={cn(
            'flex items-center gap-2 px-4 py-2.5 rounded-md border border-input bg-card',
            (selectedCategory || showFavoritesOnly) && 'border-primary text-primary'
          )}
        >
          <SlidersHorizontal className="w-4 h-4" />
          <span className="hidden sm:inline">Filters</span>
        </button>
      </div>

      {/* Mobile sidebar overlay */}
      {sidebarOpen && (
        <div
          className="lg:hidden fixed inset-0 z-40 bg-black/50"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* Sidebar */}
      <aside
        className={cn(
          'lg:w-64 flex-shrink-0 overflow-y-auto bg-card',
          'fixed lg:relative inset-y-0 left-0 z-50 w-80 max-w-[85vw] transform transition-transform lg:transform-none',
          sidebarOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0',
          'lg:bg-transparent border-r lg:border-r-0 border-border p-4 lg:p-0'
        )}
      >
        <div className="space-y-4">
          {/* Mobile close button */}
          <div className="lg:hidden flex items-center justify-between mb-4">
            <h3 className="font-semibold">Filters</h3>
            <button
              onClick={() => setSidebarOpen(false)}
              className="p-2 rounded-md hover:bg-accent"
            >
              <X className="w-5 h-5" />
            </button>
          </div>

          {/* Desktop search */}
          <div className="hidden lg:block relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            <input
              type="text"
              placeholder="Search tools..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-9 pr-4 py-2 rounded-md border border-input bg-card text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
            />
          </div>

          {/* Quick filters */}
          <div className="flex gap-2">
            <button
              onClick={() => {
                setSearchQuery('@favorites');
                setSidebarOpen(false);
              }}
              className={cn(
                'flex items-center gap-1.5 px-3 py-2 lg:py-1.5 rounded-md text-sm transition-colors flex-1 lg:flex-none justify-center lg:justify-start',
                showFavoritesOnly
                  ? 'bg-primary text-primary-foreground'
                  : 'bg-secondary text-secondary-foreground hover:bg-secondary/80'
              )}
            >
              <Star className="w-3.5 h-3.5" />
              Favorites ({favorites.length})
            </button>
          </div>

          {/* Categories */}
          <div>
            <h3 className="text-sm font-semibold text-muted-foreground mb-2 flex items-center gap-2">
              <Filter className="w-4 h-4" />
              Categories
            </h3>
            <div className="space-y-1">
              <button
                onClick={() => handleCategorySelect(null, null)}
                className={cn(
                  'w-full text-left px-3 py-2.5 lg:py-2 rounded-md text-sm transition-colors',
                  !selectedCategory && searchQuery !== '@favorites'
                    ? 'bg-primary/10 text-primary'
                    : 'text-muted-foreground hover:bg-accent hover:text-foreground'
                )}
              >
                All Tools ({tools.length})
              </button>
              {categories.map((category) => (
                <CategoryItem
                  key={category.name}
                  category={category}
                  isSelected={selectedCategory === category.name}
                  selectedSubcategory={selectedSubcategory}
                  onSelect={handleCategorySelect}
                />
              ))}
            </div>
          </div>
        </div>
      </aside>

      {/* Main content */}
      <div className="flex-1 overflow-y-auto">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg lg:text-xl font-semibold">
            {showFavoritesOnly
              ? 'Favorite Tools'
              : selectedCategory
              ? `${selectedCategory}${selectedSubcategory ? ` / ${selectedSubcategory}` : ''}`
              : 'All Tools'}
          </h2>
          <span className="text-muted-foreground text-sm">
            {displayTools.length} {displayTools.length === 1 ? 'tool' : 'tools'}
          </span>
        </div>

        {toolsLoading ? (
          <div className="flex items-center justify-center h-64">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
          </div>
        ) : displayTools.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-64 text-muted-foreground">
            <Search className="w-12 h-12 mb-4 opacity-50" />
            <p>No tools found</p>
            {searchQuery && (
              <button
                onClick={() => setSearchQuery('')}
                className="mt-2 text-primary hover:underline"
              >
                Clear search
              </button>
            )}
          </div>
        ) : (
          <div className="grid gap-3 lg:gap-4">
            {displayTools.map((tool) => (
              <ToolCard
                key={tool.name}
                tool={tool}
                isFavorite={isFavorite(tool.name)}
                onToggleFavorite={() => toggleFavorite(tool.name)}
                onExecute={() => addRecentTool(tool.name)}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function CategoryItem({
  category,
  isSelected,
  selectedSubcategory,
  onSelect,
}: {
  category: CategoryInfo;
  isSelected: boolean;
  selectedSubcategory: string | null;
  onSelect: (category: string | null, subcategory: string | null) => void;
}) {
  const hasSubcategories = category.subcategories.length > 1;

  return (
    <div>
      <button
        onClick={() => onSelect(isSelected ? null : category.name, null)}
        className={cn(
          'w-full flex items-center justify-between px-3 py-2.5 lg:py-2 rounded-md text-sm transition-colors',
          isSelected
            ? 'bg-primary/10 text-primary'
            : 'text-muted-foreground hover:bg-accent hover:text-foreground'
        )}
      >
        <span className="flex items-center gap-2">
          {hasSubcategories &&
            (isSelected ? (
              <ChevronDown className="w-3.5 h-3.5" />
            ) : (
              <ChevronRight className="w-3.5 h-3.5" />
            ))}
          <span className="capitalize">{category.name}</span>
        </span>
        <span className="text-xs opacity-60">{category.count}</span>
      </button>
      {isSelected && hasSubcategories && (
        <div className="ml-6 mt-1 space-y-1">
          {category.subcategories.map((sub) => (
            <button
              key={sub.name}
              onClick={() => onSelect(category.name, sub.name)}
              className={cn(
                'w-full text-left px-3 py-2 lg:py-1.5 rounded-md text-sm transition-colors',
                selectedSubcategory === sub.name
                  ? 'bg-primary/10 text-primary'
                  : 'text-muted-foreground hover:bg-accent hover:text-foreground'
              )}
            >
              <span className="capitalize">{sub.name}</span>
              <span className="text-xs opacity-60 ml-2">({sub.count})</span>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

function ToolCard({
  tool,
  isFavorite,
  onToggleFavorite,
  onExecute,
}: {
  tool: ToolDefinition;
  isFavorite: boolean;
  onToggleFavorite: () => void;
  onExecute: () => void;
}) {
  return (
    <div className="rounded-lg border border-border bg-card p-3 lg:p-4 hover:border-primary/30 transition-colors active:bg-accent/50">
      <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-3 sm:gap-4">
        <div className="flex-1 min-w-0">
          <div className="flex flex-wrap items-center gap-2 mb-1">
            <h3 className="font-mono text-sm font-semibold truncate">{tool.name}</h3>
            <span
              className={cn(
                'text-xs px-1.5 py-0.5 rounded capitalize',
                getComplexityColor(tool.complexity)
              )}
            >
              {tool.complexity}
            </span>
            {tool.isWrite ? (
              <span className="flex items-center gap-1 text-xs text-orange-400">
                <Pencil className="w-3 h-3" />
                Write
              </span>
            ) : (
              <span className="flex items-center gap-1 text-xs text-blue-400">
                <BookOpen className="w-3 h-3" />
                Read
              </span>
            )}
          </div>
          <p className="text-sm text-muted-foreground mb-2 line-clamp-2">
            {tool.description}
          </p>
          <div className="flex flex-wrap gap-1.5">
            {tool.tags.slice(0, 5).map((tag) => (
              <span
                key={tag}
                className="text-xs px-2 py-0.5 rounded-full bg-secondary text-secondary-foreground"
              >
                {tag}
              </span>
            ))}
          </div>
        </div>
        <div className="flex items-center gap-2 sm:flex-col-reverse sm:items-end">
          <button
            onClick={onToggleFavorite}
            className={cn(
              'p-2.5 sm:p-2 rounded-md transition-colors',
              isFavorite
                ? 'text-yellow-400 bg-yellow-400/10'
                : 'text-muted-foreground hover:text-foreground hover:bg-accent'
            )}
          >
            <Star className={cn('w-5 h-5 sm:w-4 sm:h-4', isFavorite && 'fill-current')} />
          </button>
          <Link
            to={`/tools/${tool.name}`}
            onClick={onExecute}
            className="flex-1 sm:flex-none px-4 py-2.5 sm:py-2 rounded-md bg-primary text-primary-foreground hover:bg-primary/90 transition-colors text-sm font-medium text-center"
          >
            Execute
          </Link>
        </div>
      </div>
    </div>
  );
}
