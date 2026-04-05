import { useState, useEffect, useCallback } from 'react';
import { useQuery, useMutation } from '@tanstack/react-query';
import { Play, Plus, Trash2, Zap, GripVertical, CheckCircle, XCircle, Loader, Keyboard, AlertTriangle, Music, Lightbulb, Video, Settings2 } from 'lucide-react';
import { api } from '../api/client';
import { useToolStore } from '../stores/toolStore';
import { useDashboardStore } from '../stores/dashboardStore';
import { cn } from '../lib/utils';
import type { Workflow } from '../api/types';

type ActionCategory = 'system' | 'video' | 'audio' | 'lighting' | 'custom';

type ActionButton = {
  id: string;
  type: 'tool' | 'workflow';
  name: string;
  label: string;
  color: string;
  icon?: string;
  category?: ActionCategory;
  shortcut?: string;
};

const CATEGORY_ICONS = {
  system: Settings2,
  video: Video,
  audio: Music,
  lighting: Lightbulb,
  custom: Zap,
};

const CATEGORY_COLORS = {
  system: 'border-blue-500/50',
  video: 'border-purple-500/50',
  audio: 'border-green-500/50',
  lighting: 'border-yellow-500/50',
  custom: 'border-primary/50',
};

const DEFAULT_COLORS = [
  'bg-blue-500',
  'bg-green-500',
  'bg-purple-500',
  'bg-orange-500',
  'bg-pink-500',
  'bg-cyan-500',
  'bg-yellow-500',
  'bg-red-500',
];

const DEFAULT_ACTIONS: ActionButton[] = [
  { id: '1', type: 'workflow', name: 'startup', label: 'STARTUP', color: 'bg-green-500', icon: '🚀', category: 'system', shortcut: '1' },
  { id: '2', type: 'workflow', name: 'shutdown', label: 'SHUTDOWN', color: 'bg-red-500', icon: '🛑', category: 'system', shortcut: '2' },
  { id: '3', type: 'tool', name: 'webb_dashboard_quick', label: 'STATUS', color: 'bg-blue-500', icon: '📊', category: 'system', shortcut: '3' },
  { id: '4', type: 'tool', name: 'webb_resolume_trigger', label: 'VJ TRIGGER', color: 'bg-purple-500', icon: '🎬', category: 'video', shortcut: 'Q' },
  { id: '5', type: 'tool', name: 'webb_grandma3_blackout', label: 'BLACKOUT', color: 'bg-gray-700', icon: '⬛', category: 'lighting', shortcut: 'B' },
  { id: '6', type: 'tool', name: 'webb_obs_scene', label: 'OBS SCENE', color: 'bg-orange-500', icon: '📹', category: 'video', shortcut: 'O' },
  { id: '7', type: 'tool', name: 'webb_ableton_play', label: 'PLAY', color: 'bg-green-600', icon: '▶️', category: 'audio', shortcut: 'P' },
  { id: '8', type: 'tool', name: 'webb_ableton_stop', label: 'STOP', color: 'bg-red-600', icon: '⏹️', category: 'audio', shortcut: 'S' },
];

export function QuickActions() {
  const { favorites, tools } = useToolStore();
  const { overallHealth, onlineCount, totalSystems, criticalIssues } = useDashboardStore();
  const [actions, setActions] = useState<ActionButton[]>(DEFAULT_ACTIONS);
  const [editMode, setEditMode] = useState(false);
  const [executingId, setExecutingId] = useState<string | null>(null);
  const [lastResult, setLastResult] = useState<{ id: string; success: boolean } | null>(null);
  const [showShortcuts, setShowShortcuts] = useState(false);
  const [panicMode, setPanicMode] = useState(false);

  const { data: workflows } = useQuery({
    queryKey: ['workflows'],
    queryFn: () => api.getWorkflows(),
  });

  const executeTool = useMutation({
    mutationFn: (toolName: string) => api.executeTool(toolName, {}),
  });

  const executeWorkflow = useMutation({
    mutationFn: (workflowName: string) => api.runWorkflow(workflowName),
  });

  const handleExecute = useCallback(async (action: ActionButton) => {
    if (editMode) return;
    setExecutingId(action.id);
    setLastResult(null);

    try {
      if (action.type === 'tool') {
        await executeTool.mutateAsync(action.name);
      } else {
        await executeWorkflow.mutateAsync(action.name);
      }
      setLastResult({ id: action.id, success: true });
    } catch {
      setLastResult({ id: action.id, success: false });
    } finally {
      setExecutingId(null);
      setTimeout(() => setLastResult(null), 2000);
    }
  }, [editMode, executeTool, executeWorkflow]);

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (editMode) return;
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return;

      // Escape for panic mode
      if (e.key === 'Escape') {
        setPanicMode(true);
        // Execute blackout if available
        const blackoutAction = actions.find(a => a.name.includes('blackout'));
        if (blackoutAction) handleExecute(blackoutAction);
        setTimeout(() => setPanicMode(false), 3000);
        return;
      }

      const key = e.key.toUpperCase();
      const action = actions.find(a => a.shortcut?.toUpperCase() === key);
      if (action) {
        e.preventDefault();
        handleExecute(action);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [actions, editMode, handleExecute]);

  const addAction = (type: 'tool' | 'workflow', name: string, label: string) => {
    const newAction: ActionButton = {
      id: crypto.randomUUID(),
      type,
      name,
      label: label.toUpperCase(),
      color: DEFAULT_COLORS[actions.length % DEFAULT_COLORS.length],
      category: 'custom',
    };
    setActions([...actions, newAction]);
  };

  const removeAction = (id: string) => {
    setActions(actions.filter((a) => a.id !== id));
  };

  const favoriteTools = tools.filter((t) => favorites.includes(t.name));

  // Group actions by category
  const actionsByCategory = actions.reduce((acc, action) => {
    const cat = action.category || 'custom';
    if (!acc[cat]) acc[cat] = [];
    acc[cat].push(action);
    return acc;
  }, {} as Record<ActionCategory, ActionButton[]>);

  return (
    <div className="space-y-6">
      {/* Panic Mode Overlay */}
      {panicMode && (
        <div className="fixed inset-0 z-50 bg-red-900/90 flex items-center justify-center animate-pulse">
          <div className="text-center">
            <AlertTriangle className="w-24 h-24 text-white mx-auto mb-4" />
            <h1 className="text-4xl font-bold text-white">PANIC MODE</h1>
            <p className="text-white/80 mt-2">Executing blackout...</p>
          </div>
        </div>
      )}

      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold flex items-center gap-2">
            <Zap className="w-6 h-6 text-yellow-400" />
            Quick Actions
          </h1>
          <p className="text-muted-foreground">
            Stream Deck-style buttons for rapid tool execution
          </p>
        </div>
        <div className="flex items-center gap-3">
          {/* System Status */}
          <div className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-card border border-border">
            <div className={cn(
              'w-2 h-2 rounded-full',
              overallHealth >= 80 ? 'bg-green-500' : overallHealth >= 50 ? 'bg-yellow-500' : 'bg-red-500'
            )} />
            <span className="text-sm">{onlineCount}/{totalSystems}</span>
            {criticalIssues > 0 && (
              <span className="flex items-center gap-1 text-xs text-red-400">
                <AlertTriangle className="w-3 h-3" />
                {criticalIssues}
              </span>
            )}
          </div>

          {/* Shortcuts Toggle */}
          <button
            onClick={() => setShowShortcuts(!showShortcuts)}
            className={cn(
              'p-2 rounded-md transition-colors',
              showShortcuts ? 'bg-primary text-primary-foreground' : 'bg-secondary hover:bg-secondary/80'
            )}
            title="Toggle keyboard shortcuts"
          >
            <Keyboard className="w-4 h-4" />
          </button>

          {/* Edit Mode */}
          <button
            onClick={() => setEditMode(!editMode)}
            className={cn(
              'px-4 py-2 rounded-md transition-colors',
              editMode
                ? 'bg-primary text-primary-foreground'
                : 'bg-secondary text-secondary-foreground hover:bg-secondary/80'
            )}
          >
            {editMode ? 'Done' : 'Edit'}
          </button>
        </div>
      </div>

      {/* Keyboard Shortcuts Help */}
      {showShortcuts && (
        <div className="rounded-lg border border-border bg-card p-4">
          <h3 className="font-semibold mb-2 flex items-center gap-2">
            <Keyboard className="w-4 h-4" />
            Keyboard Shortcuts
          </h3>
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-2 text-sm">
            {actions.filter(a => a.shortcut).map(action => (
              <div key={action.id} className="flex items-center gap-2">
                <kbd className="px-2 py-1 rounded bg-secondary font-mono text-xs">{action.shortcut}</kbd>
                <span className="text-muted-foreground">{action.label}</span>
              </div>
            ))}
            <div className="flex items-center gap-2">
              <kbd className="px-2 py-1 rounded bg-red-500/20 text-red-400 font-mono text-xs">ESC</kbd>
              <span className="text-red-400">PANIC</span>
            </div>
          </div>
        </div>
      )}

      {/* Action Grid by Category */}
      {(['system', 'video', 'audio', 'lighting', 'custom'] as ActionCategory[]).map(category => {
        const categoryActions = actionsByCategory[category];
        if (!categoryActions || categoryActions.length === 0) return null;

        const CategoryIcon = CATEGORY_ICONS[category];

        return (
          <div key={category} className={cn('rounded-lg border-2 p-4', CATEGORY_COLORS[category])}>
            <h2 className="text-sm font-semibold text-muted-foreground mb-3 flex items-center gap-2 capitalize">
              <CategoryIcon className="w-4 h-4" />
              {category}
            </h2>
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-4">
              {categoryActions.map((action) => (
                <ActionButtonCard
                  key={action.id}
                  action={action}
                  editMode={editMode}
                  isExecuting={executingId === action.id}
                  lastResult={lastResult?.id === action.id ? lastResult.success : null}
                  showShortcut={showShortcuts}
                  onExecute={() => handleExecute(action)}
                  onRemove={() => removeAction(action.id)}
                />
              ))}
            </div>
          </div>
        );
      })}

      {/* Add Button */}
      {editMode && (
        <button
          className="w-full py-8 rounded-xl border-2 border-dashed border-border flex flex-col items-center justify-center gap-2 text-muted-foreground hover:text-foreground hover:border-primary/50 transition-colors"
          onClick={() => {
            const name = prompt('Tool or workflow name:');
            if (name) {
              const label = prompt('Button label:') || name;
              addAction('tool', name, label);
            }
          }}
        >
          <Plus className="w-6 h-6" />
          <span className="text-sm">Add New Action</span>
        </button>
      )}

      {/* Favorite Tools Quick Access */}
      {favoriteTools.length > 0 && (
        <div>
          <h2 className="text-lg font-semibold mb-3">Favorite Tools</h2>
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-3">
            {favoriteTools.slice(0, 12).map((tool) => (
              <button
                key={tool.name}
                onClick={() => handleExecute({ id: tool.name, type: 'tool', name: tool.name, label: tool.name, color: 'bg-yellow-500' })}
                disabled={executingId === tool.name}
                className="p-3 rounded-lg border border-border bg-card hover:border-primary/50 transition-colors text-left"
              >
                <div className="font-mono text-sm truncate">{tool.name.replace('webb_', '')}</div>
                <div className="text-xs text-muted-foreground truncate mt-1">{tool.description.slice(0, 40)}...</div>
              </button>
            ))}
          </div>
        </div>
      )}

      {/* Workflows */}
      {workflows && workflows.length > 0 && (
        <div>
          <h2 className="text-lg font-semibold mb-3">Workflows</h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
            {workflows.map((workflow) => (
              <WorkflowCard
                key={workflow.name}
                workflow={workflow}
                isExecuting={executingId === workflow.name}
                onExecute={() => handleExecute({ id: workflow.name, type: 'workflow', name: workflow.name, label: workflow.name, color: 'bg-primary' })}
              />
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

function ActionButtonCard({
  action,
  editMode,
  isExecuting,
  lastResult,
  showShortcut,
  onExecute,
  onRemove,
}: {
  action: ActionButton;
  editMode: boolean;
  isExecuting: boolean;
  lastResult: boolean | null;
  showShortcut?: boolean;
  onExecute: () => void;
  onRemove: () => void;
}) {
  return (
    <div className="relative group">
      <button
        onClick={onExecute}
        disabled={editMode || isExecuting}
        className={cn(
          'aspect-square w-full rounded-xl flex flex-col items-center justify-center gap-2 font-bold text-white shadow-lg transition-all',
          action.color,
          !editMode && !isExecuting && 'hover:scale-105 hover:shadow-xl active:scale-95',
          editMode && 'opacity-80 cursor-default',
          isExecuting && 'animate-pulse'
        )}
      >
        {isExecuting ? (
          <Loader className="w-8 h-8 animate-spin" />
        ) : lastResult !== null ? (
          lastResult ? (
            <CheckCircle className="w-8 h-8 text-white" />
          ) : (
            <XCircle className="w-8 h-8 text-white" />
          )
        ) : (
          <>
            <span className="text-2xl">{action.icon || '⚡'}</span>
            <span className="text-xs px-2 text-center leading-tight">{action.label}</span>
          </>
        )}
      </button>

      {/* Keyboard Shortcut Badge */}
      {showShortcut && action.shortcut && (
        <div className="absolute -bottom-1 left-1/2 -translate-x-1/2 px-1.5 py-0.5 rounded bg-black/80 text-white text-[10px] font-mono">
          {action.shortcut}
        </div>
      )}

      {editMode && (
        <>
          <button className="absolute -top-2 -left-2 p-1 rounded-full bg-card border border-border shadow cursor-grab">
            <GripVertical className="w-4 h-4 text-muted-foreground" />
          </button>
          <button
            onClick={onRemove}
            className="absolute -top-2 -right-2 p-1 rounded-full bg-red-500 text-white shadow hover:bg-red-600"
          >
            <Trash2 className="w-4 h-4" />
          </button>
        </>
      )}
    </div>
  );
}

function WorkflowCard({
  workflow,
  isExecuting,
  onExecute,
}: {
  workflow: Workflow;
  isExecuting: boolean;
  onExecute: () => void;
}) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="flex items-start justify-between">
        <div>
          <h3 className="font-semibold capitalize">{workflow.name}</h3>
          <p className="text-sm text-muted-foreground mt-1">{workflow.description}</p>
          <p className="text-xs text-muted-foreground mt-2">{workflow.steps.length} steps</p>
        </div>
        <button
          onClick={onExecute}
          disabled={isExecuting}
          className={cn(
            'p-2 rounded-md bg-primary text-primary-foreground hover:bg-primary/90 transition-colors',
            isExecuting && 'opacity-50 cursor-not-allowed'
          )}
        >
          {isExecuting ? <Loader className="w-4 h-4 animate-spin" /> : <Play className="w-4 h-4" />}
        </button>
      </div>
    </div>
  );
}
