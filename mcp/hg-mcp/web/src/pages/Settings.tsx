import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Settings as SettingsIcon, Star, Tag, Trash2, Plus, Save, RefreshCw } from 'lucide-react';
import { api } from '../api/client';
import { useToolStore } from '../stores/toolStore';

export function Settings() {
  const queryClient = useQueryClient();
  const { favorites, aliases, clearFavorites, setAliases, clearRecentTools, recentTools } = useToolStore();
  const [newAlias, setNewAlias] = useState({ alias: '', toolName: '' });

  const { data: serverAliases } = useQuery({
    queryKey: ['aliases'],
    queryFn: () => api.getAliases(),
  });

  const saveAliases = useMutation({
    mutationFn: (aliases: Record<string, string>) => api.updateAliases(aliases),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['aliases'] });
    },
  });

  const handleAddAlias = () => {
    if (newAlias.alias && newAlias.toolName) {
      setAliases({ ...aliases, [newAlias.alias]: newAlias.toolName });
      setNewAlias({ alias: '', toolName: '' });
    }
  };

  const handleRemoveAlias = (alias: string) => {
    const updated = { ...aliases };
    delete updated[alias];
    setAliases(updated);
  };

  const handleSyncAliases = () => {
    saveAliases.mutate(aliases);
  };

  return (
    <div className="max-w-3xl mx-auto space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold flex items-center gap-2">
          <SettingsIcon className="w-6 h-6" />
          Settings
        </h1>
        <p className="text-muted-foreground">Manage your preferences and tool configuration</p>
      </div>

      {/* Favorites Section */}
      <section className="rounded-lg border border-border bg-card p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold flex items-center gap-2">
            <Star className="w-5 h-5 text-yellow-400" />
            Favorites
          </h2>
          {favorites.length > 0 && (
            <button
              onClick={() => {
                if (confirm('Clear all favorites?')) clearFavorites();
              }}
              className="text-sm text-red-400 hover:text-red-300"
            >
              Clear All
            </button>
          )}
        </div>

        {favorites.length === 0 ? (
          <p className="text-muted-foreground text-sm">No favorite tools yet. Star tools to add them here.</p>
        ) : (
          <div className="flex flex-wrap gap-2">
            {favorites.map((name) => (
              <span
                key={name}
                className="inline-flex items-center gap-1 px-3 py-1.5 rounded-full bg-yellow-400/10 text-yellow-400 text-sm"
              >
                <Star className="w-3 h-3 fill-current" />
                {name}
              </span>
            ))}
          </div>
        )}
      </section>

      {/* Aliases Section */}
      <section className="rounded-lg border border-border bg-card p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold flex items-center gap-2">
            <Tag className="w-5 h-5 text-blue-400" />
            Tool Aliases
          </h2>
          <button
            onClick={handleSyncAliases}
            disabled={saveAliases.isPending}
            className="flex items-center gap-1 text-sm text-primary hover:text-primary/80"
          >
            {saveAliases.isPending ? (
              <RefreshCw className="w-4 h-4 animate-spin" />
            ) : (
              <Save className="w-4 h-4" />
            )}
            Sync to Server
          </button>
        </div>

        <p className="text-muted-foreground text-sm mb-4">
          Create short aliases for frequently used tools
        </p>

        {/* Add new alias */}
        <div className="flex gap-2 mb-4">
          <input
            type="text"
            placeholder="Alias (e.g., st)"
            value={newAlias.alias}
            onChange={(e) => setNewAlias({ ...newAlias, alias: e.target.value })}
            className="flex-1 px-3 py-2 rounded-md border border-input bg-background text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
          />
          <input
            type="text"
            placeholder="Tool name (e.g., webb_resolume_status)"
            value={newAlias.toolName}
            onChange={(e) => setNewAlias({ ...newAlias, toolName: e.target.value })}
            className="flex-1 px-3 py-2 rounded-md border border-input bg-background text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
          />
          <button
            onClick={handleAddAlias}
            disabled={!newAlias.alias || !newAlias.toolName}
            className="px-4 py-2 rounded-md bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
          >
            <Plus className="w-4 h-4" />
          </button>
        </div>

        {/* Alias list */}
        {Object.keys(aliases).length === 0 ? (
          <p className="text-muted-foreground text-sm">No aliases configured</p>
        ) : (
          <div className="space-y-2">
            {Object.entries(aliases).map(([alias, toolName]) => (
              <div
                key={alias}
                className="flex items-center justify-between p-3 rounded-md bg-secondary/50"
              >
                <div className="flex items-center gap-3">
                  <code className="px-2 py-0.5 rounded bg-primary/20 text-primary text-sm font-mono">
                    {alias}
                  </code>
                  <span className="text-muted-foreground">→</span>
                  <code className="text-sm font-mono">{toolName}</code>
                </div>
                <button
                  onClick={() => handleRemoveAlias(alias)}
                  className="p-1 text-muted-foreground hover:text-red-400"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            ))}
          </div>
        )}

        {/* Server aliases */}
        {serverAliases && Object.keys(serverAliases).length > 0 && (
          <div className="mt-4 pt-4 border-t border-border">
            <h3 className="text-sm font-medium text-muted-foreground mb-2">Server Aliases</h3>
            <div className="flex flex-wrap gap-2">
              {Object.entries(serverAliases).map(([alias, toolName]) => (
                <span key={alias} className="text-xs px-2 py-1 rounded bg-secondary text-secondary-foreground">
                  {alias} → {toolName}
                </span>
              ))}
            </div>
          </div>
        )}
      </section>

      {/* Recent Tools Section */}
      <section className="rounded-lg border border-border bg-card p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold">Recent Tools</h2>
          {recentTools.length > 0 && (
            <button
              onClick={() => {
                if (confirm('Clear recent tools history?')) clearRecentTools();
              }}
              className="text-sm text-red-400 hover:text-red-300"
            >
              Clear History
            </button>
          )}
        </div>

        {recentTools.length === 0 ? (
          <p className="text-muted-foreground text-sm">No recent tools</p>
        ) : (
          <div className="flex flex-wrap gap-2">
            {recentTools.slice(0, 20).map((name, idx) => (
              <span key={`${name}-${idx}`} className="text-xs px-2 py-1 rounded bg-secondary text-secondary-foreground">
                {name}
              </span>
            ))}
          </div>
        )}
      </section>

      {/* Data Management */}
      <section className="rounded-lg border border-border bg-card p-6">
        <h2 className="text-lg font-semibold mb-4">Data Management</h2>
        <div className="space-y-3">
          <button
            onClick={() => {
              if (confirm('Clear all local data? This includes favorites, aliases, and history.')) {
                localStorage.clear();
                window.location.reload();
              }
            }}
            className="w-full px-4 py-3 rounded-md border border-red-500/30 text-red-400 hover:bg-red-500/10 transition-colors text-left"
          >
            <div className="font-medium">Clear All Local Data</div>
            <div className="text-sm text-muted-foreground">
              Remove all favorites, aliases, and execution history stored in your browser
            </div>
          </button>
        </div>
      </section>

      {/* About */}
      <section className="rounded-lg border border-border bg-card p-6">
        <h2 className="text-lg font-semibold mb-4">About</h2>
        <div className="space-y-2 text-sm text-muted-foreground">
          <p>
            <strong className="text-foreground">AFTRS MCP Web UI</strong>
          </p>
          <p>Unified interface for controlling 672+ MCP tools</p>
          <p className="pt-2">
            Built with React, TypeScript, and Tailwind CSS.
            <br />
            Backend powered by Go REST API with SSE streaming.
          </p>
        </div>
      </section>
    </div>
  );
}
