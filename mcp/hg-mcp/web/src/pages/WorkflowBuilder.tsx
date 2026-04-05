import { useState, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Plus,
  Play,
  Save,
  Trash2,
  GripVertical,
  ChevronDown,
  ChevronRight,
  X,
  Check,
  Loader,
  AlertTriangle,
  ArrowRight,
  Settings2,
  GitBranch,
} from 'lucide-react';
import { api } from '../api/client';
import { useToolStore } from '../stores/toolStore';
import { cn } from '../lib/utils';
import type { Workflow, WorkflowStep, ToolDefinition } from '../api/types';

type EditorMode = 'list' | 'edit' | 'create';

interface StepWithId extends WorkflowStep {
  id: string;
}

export function WorkflowBuilder() {
  const queryClient = useQueryClient();
  const { tools } = useToolStore();
  const [mode, setMode] = useState<EditorMode>('list');
  const [currentWorkflow, setCurrentWorkflow] = useState<Workflow | null>(null);
  const [steps, setSteps] = useState<StepWithId[]>([]);
  const [workflowName, setWorkflowName] = useState('');
  const [workflowDescription, setWorkflowDescription] = useState('');
  const [draggedStep, setDraggedStep] = useState<number | null>(null);
  const [expandedStep, setExpandedStep] = useState<number | null>(null);
  const [runningWorkflow, setRunningWorkflow] = useState<string | null>(null);
  const [runResults, setRunResults] = useState<{ success: boolean; steps: { tool: string; success: boolean; error?: string }[] } | null>(null);

  const { data: workflows, isLoading } = useQuery({
    queryKey: ['workflows'],
    queryFn: () => api.getWorkflows(),
  });

  const { data: allTools } = useQuery({
    queryKey: ['tools'],
    queryFn: () => api.getTools(),
  });

  const createMutation = useMutation({
    mutationFn: (workflow: Workflow) => api.createWorkflow(workflow),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workflows'] });
      setMode('list');
      resetEditor();
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ name, workflow }: { name: string; workflow: Workflow }) =>
      api.updateWorkflow(name, workflow),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workflows'] });
      setMode('list');
      resetEditor();
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (name: string) => api.deleteWorkflow(name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workflows'] });
    },
  });

  const runMutation = useMutation({
    mutationFn: (name: string) => api.runWorkflow(name),
  });

  const resetEditor = () => {
    setCurrentWorkflow(null);
    setSteps([]);
    setWorkflowName('');
    setWorkflowDescription('');
    setExpandedStep(null);
    setRunResults(null);
  };

  const startCreate = () => {
    resetEditor();
    setMode('create');
  };

  const startEdit = (workflow: Workflow) => {
    setCurrentWorkflow(workflow);
    setWorkflowName(workflow.name);
    setWorkflowDescription(workflow.description);
    setSteps(workflow.steps.map((s, i) => ({ ...s, id: `step-${i}-${Date.now()}` })));
    setMode('edit');
  };

  const handleSave = () => {
    const workflow: Workflow = {
      name: workflowName.toLowerCase().replace(/\s+/g, '_'),
      description: workflowDescription,
      steps: steps.map(({ id, ...step }) => step),
    };

    if (mode === 'create') {
      createMutation.mutate(workflow);
    } else if (mode === 'edit' && currentWorkflow) {
      updateMutation.mutate({ name: currentWorkflow.name, workflow });
    }
  };

  const handleRun = async (name: string) => {
    setRunningWorkflow(name);
    setRunResults(null);
    try {
      const results = await runMutation.mutateAsync(name);
      setRunResults({
        success: results.every((r) => r.success),
        steps: results.map((r, i) => ({
          tool: workflows?.find((w) => w.name === name)?.steps[i]?.tool || 'unknown',
          success: r.success,
          error: r.error,
        })),
      });
    } catch (error) {
      setRunResults({
        success: false,
        steps: [{ tool: 'workflow', success: false, error: error instanceof Error ? error.message : 'Unknown error' }],
      });
    } finally {
      setRunningWorkflow(null);
    }
  };

  const addStep = () => {
    const newStep: StepWithId = {
      id: `step-${Date.now()}`,
      tool: '',
      args: {},
    };
    setSteps([...steps, newStep]);
    setExpandedStep(steps.length);
  };

  const removeStep = (index: number) => {
    setSteps(steps.filter((_, i) => i !== index));
    if (expandedStep === index) setExpandedStep(null);
    else if (expandedStep !== null && expandedStep > index) setExpandedStep(expandedStep - 1);
  };

  const updateStep = (index: number, updates: Partial<StepWithId>) => {
    setSteps(steps.map((s, i) => (i === index ? { ...s, ...updates } : s)));
  };

  const handleDragStart = (index: number) => {
    setDraggedStep(index);
  };

  const handleDragOver = useCallback((e: React.DragEvent, index: number) => {
    e.preventDefault();
    if (draggedStep === null || draggedStep === index) return;

    setSteps((prevSteps) => {
      const newSteps = [...prevSteps];
      const [draggedItem] = newSteps.splice(draggedStep, 1);
      newSteps.splice(index, 0, draggedItem);
      setDraggedStep(index);
      return newSteps;
    });
  }, [draggedStep]);

  const handleDragEnd = () => {
    setDraggedStep(null);
  };

  const availableTools = allTools || tools;

  if (mode !== 'list') {
    return (
      <WorkflowEditor
        mode={mode}
        workflowName={workflowName}
        setWorkflowName={setWorkflowName}
        workflowDescription={workflowDescription}
        setWorkflowDescription={setWorkflowDescription}
        steps={steps}
        expandedStep={expandedStep}
        setExpandedStep={setExpandedStep}
        addStep={addStep}
        removeStep={removeStep}
        updateStep={updateStep}
        handleDragStart={handleDragStart}
        handleDragOver={handleDragOver}
        handleDragEnd={handleDragEnd}
        draggedStep={draggedStep}
        availableTools={availableTools}
        onSave={handleSave}
        onCancel={() => {
          setMode('list');
          resetEditor();
        }}
        isSaving={createMutation.isPending || updateMutation.isPending}
      />
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold flex items-center gap-2">
            <GitBranch className="w-6 h-6 text-primary" />
            Workflow Builder
          </h1>
          <p className="text-muted-foreground">
            Create and manage automated tool sequences
          </p>
        </div>
        <button
          onClick={startCreate}
          className="flex items-center gap-2 px-4 py-2 rounded-md bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
        >
          <Plus className="w-4 h-4" />
          New Workflow
        </button>
      </div>

      {/* Run Results Banner */}
      {runResults && (
        <div
          className={cn(
            'p-4 rounded-lg border flex items-center justify-between',
            runResults.success
              ? 'bg-green-500/10 border-green-500/30 text-green-400'
              : 'bg-red-500/10 border-red-500/30 text-red-400'
          )}
        >
          <div className="flex items-center gap-3">
            {runResults.success ? (
              <Check className="w-5 h-5" />
            ) : (
              <AlertTriangle className="w-5 h-5" />
            )}
            <span>
              {runResults.success
                ? `Workflow completed successfully (${runResults.steps.length} steps)`
                : `Workflow failed: ${runResults.steps.find((s) => !s.success)?.error || 'Unknown error'}`}
            </span>
          </div>
          <button onClick={() => setRunResults(null)} className="p-1 hover:bg-white/10 rounded">
            <X className="w-4 h-4" />
          </button>
        </div>
      )}

      {/* Workflows List */}
      {isLoading ? (
        <div className="flex items-center justify-center py-12">
          <Loader className="w-6 h-6 animate-spin text-muted-foreground" />
        </div>
      ) : !workflows || workflows.length === 0 ? (
        <div className="text-center py-12 border border-dashed border-border rounded-lg">
          <GitBranch className="w-12 h-12 mx-auto text-muted-foreground mb-4" />
          <h3 className="text-lg font-medium mb-2">No Workflows Yet</h3>
          <p className="text-muted-foreground mb-4">
            Create your first workflow to automate tool sequences
          </p>
          <button
            onClick={startCreate}
            className="inline-flex items-center gap-2 px-4 py-2 rounded-md bg-primary text-primary-foreground hover:bg-primary/90"
          >
            <Plus className="w-4 h-4" />
            Create Workflow
          </button>
        </div>
      ) : (
        <div className="grid gap-4">
          {workflows.map((workflow) => (
            <WorkflowCard
              key={workflow.name}
              workflow={workflow}
              isRunning={runningWorkflow === workflow.name}
              onEdit={() => startEdit(workflow)}
              onRun={() => handleRun(workflow.name)}
              onDelete={() => {
                if (confirm(`Delete workflow "${workflow.name}"?`)) {
                  deleteMutation.mutate(workflow.name);
                }
              }}
            />
          ))}
        </div>
      )}
    </div>
  );
}

function WorkflowCard({
  workflow,
  isRunning,
  onEdit,
  onRun,
  onDelete,
}: {
  workflow: Workflow;
  isRunning: boolean;
  onEdit: () => void;
  onRun: () => void;
  onDelete: () => void;
}) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <h3 className="font-semibold text-lg capitalize">{workflow.name.replace(/_/g, ' ')}</h3>
          <p className="text-sm text-muted-foreground mt-1">{workflow.description}</p>
          <div className="flex items-center gap-2 mt-3">
            {workflow.steps.slice(0, 5).map((step, i) => (
              <div key={i} className="flex items-center gap-1">
                <span className="px-2 py-1 text-xs rounded bg-secondary font-mono">
                  {step.tool.replace('webb_', '')}
                </span>
                {i < Math.min(workflow.steps.length - 1, 4) && (
                  <ArrowRight className="w-3 h-3 text-muted-foreground" />
                )}
              </div>
            ))}
            {workflow.steps.length > 5 && (
              <span className="text-xs text-muted-foreground">+{workflow.steps.length - 5} more</span>
            )}
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={onRun}
            disabled={isRunning}
            className={cn(
              'p-2 rounded-md bg-green-600 text-white hover:bg-green-700 transition-colors',
              isRunning && 'opacity-50 cursor-not-allowed'
            )}
            title="Run workflow"
          >
            {isRunning ? <Loader className="w-4 h-4 animate-spin" /> : <Play className="w-4 h-4" />}
          </button>
          <button
            onClick={onEdit}
            className="p-2 rounded-md bg-secondary text-secondary-foreground hover:bg-secondary/80 transition-colors"
            title="Edit workflow"
          >
            <Settings2 className="w-4 h-4" />
          </button>
          <button
            onClick={onDelete}
            className="p-2 rounded-md bg-red-500/20 text-red-400 hover:bg-red-500/30 transition-colors"
            title="Delete workflow"
          >
            <Trash2 className="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  );
}

function WorkflowEditor({
  mode,
  workflowName,
  setWorkflowName,
  workflowDescription,
  setWorkflowDescription,
  steps,
  expandedStep,
  setExpandedStep,
  addStep,
  removeStep,
  updateStep,
  handleDragStart,
  handleDragOver,
  handleDragEnd,
  draggedStep,
  availableTools,
  onSave,
  onCancel,
  isSaving,
}: {
  mode: EditorMode;
  workflowName: string;
  setWorkflowName: (name: string) => void;
  workflowDescription: string;
  setWorkflowDescription: (desc: string) => void;
  steps: StepWithId[];
  expandedStep: number | null;
  setExpandedStep: (index: number | null) => void;
  addStep: () => void;
  removeStep: (index: number) => void;
  updateStep: (index: number, updates: Partial<StepWithId>) => void;
  handleDragStart: (index: number) => void;
  handleDragOver: (e: React.DragEvent, index: number) => void;
  handleDragEnd: () => void;
  draggedStep: number | null;
  availableTools: ToolDefinition[];
  onSave: () => void;
  onCancel: () => void;
  isSaving: boolean;
}) {
  const [toolSearch, setToolSearch] = useState('');
  const [showToolPicker, setShowToolPicker] = useState<number | null>(null);

  const filteredTools = availableTools.filter(
    (t) =>
      t.name.toLowerCase().includes(toolSearch.toLowerCase()) ||
      t.description.toLowerCase().includes(toolSearch.toLowerCase())
  );

  const canSave = workflowName.trim() && steps.length > 0 && steps.every((s) => s.tool);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">
            {mode === 'create' ? 'Create Workflow' : 'Edit Workflow'}
          </h1>
          <p className="text-muted-foreground">
            {mode === 'create'
              ? 'Design a new automated tool sequence'
              : `Editing "${workflowName}"`}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={onCancel}
            className="px-4 py-2 rounded-md bg-secondary text-secondary-foreground hover:bg-secondary/80 transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={onSave}
            disabled={!canSave || isSaving}
            className={cn(
              'flex items-center gap-2 px-4 py-2 rounded-md bg-primary text-primary-foreground hover:bg-primary/90 transition-colors',
              (!canSave || isSaving) && 'opacity-50 cursor-not-allowed'
            )}
          >
            {isSaving ? <Loader className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
            Save Workflow
          </button>
        </div>
      </div>

      {/* Workflow Metadata */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 p-4 rounded-lg border border-border bg-card">
        <div>
          <label className="block text-sm font-medium mb-1">Workflow Name</label>
          <input
            type="text"
            value={workflowName}
            onChange={(e) => setWorkflowName(e.target.value)}
            placeholder="e.g., startup_sequence"
            className="w-full px-3 py-2 rounded-md border border-border bg-background focus:outline-none focus:ring-2 focus:ring-primary"
          />
        </div>
        <div>
          <label className="block text-sm font-medium mb-1">Description</label>
          <input
            type="text"
            value={workflowDescription}
            onChange={(e) => setWorkflowDescription(e.target.value)}
            placeholder="What does this workflow do?"
            className="w-full px-3 py-2 rounded-md border border-border bg-background focus:outline-none focus:ring-2 focus:ring-primary"
          />
        </div>
      </div>

      {/* Steps */}
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold">Steps ({steps.length})</h2>
          <button
            onClick={addStep}
            className="flex items-center gap-1 text-sm text-primary hover:text-primary/80"
          >
            <Plus className="w-4 h-4" />
            Add Step
          </button>
        </div>

        {steps.length === 0 ? (
          <div className="text-center py-8 border border-dashed border-border rounded-lg">
            <p className="text-muted-foreground mb-2">No steps yet</p>
            <button onClick={addStep} className="text-primary hover:underline">
              Add your first step
            </button>
          </div>
        ) : (
          <div className="space-y-2">
            {steps.map((step, index) => (
              <div
                key={step.id}
                draggable
                onDragStart={() => handleDragStart(index)}
                onDragOver={(e) => handleDragOver(e, index)}
                onDragEnd={handleDragEnd}
                className={cn(
                  'rounded-lg border border-border bg-card transition-all',
                  draggedStep === index && 'opacity-50 border-primary'
                )}
              >
                {/* Step Header */}
                <div
                  className="flex items-center gap-3 p-3 cursor-pointer"
                  onClick={() => setExpandedStep(expandedStep === index ? null : index)}
                >
                  <div className="cursor-grab hover:bg-secondary p-1 rounded">
                    <GripVertical className="w-4 h-4 text-muted-foreground" />
                  </div>
                  <div className="w-6 h-6 rounded-full bg-primary/20 text-primary text-xs flex items-center justify-center font-medium">
                    {index + 1}
                  </div>
                  <div className="flex-1">
                    {step.tool ? (
                      <div>
                        <span className="font-mono text-sm">{step.tool}</span>
                        {step.onFail && (
                          <span className="ml-2 text-xs text-orange-400">
                            on fail: {step.onFail}
                          </span>
                        )}
                      </div>
                    ) : (
                      <span className="text-muted-foreground text-sm italic">Select a tool...</span>
                    )}
                  </div>
                  {expandedStep === index ? (
                    <ChevronDown className="w-4 h-4 text-muted-foreground" />
                  ) : (
                    <ChevronRight className="w-4 h-4 text-muted-foreground" />
                  )}
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      removeStep(index);
                    }}
                    className="p-1 rounded hover:bg-red-500/20 text-red-400"
                  >
                    <X className="w-4 h-4" />
                  </button>
                </div>

                {/* Step Details */}
                {expandedStep === index && (
                  <div className="p-4 pt-0 border-t border-border space-y-4">
                    {/* Tool Selector */}
                    <div>
                      <label className="block text-sm font-medium mb-2">Tool</label>
                      <div className="relative">
                        <button
                          onClick={() => setShowToolPicker(showToolPicker === index ? null : index)}
                          className="w-full px-3 py-2 rounded-md border border-border bg-background text-left flex items-center justify-between hover:border-primary/50"
                        >
                          <span className={step.tool ? '' : 'text-muted-foreground'}>
                            {step.tool || 'Select a tool...'}
                          </span>
                          <ChevronDown className="w-4 h-4 text-muted-foreground" />
                        </button>

                        {showToolPicker === index && (
                          <div className="absolute z-10 mt-1 w-full max-h-64 overflow-auto rounded-md border border-border bg-card shadow-lg">
                            <div className="sticky top-0 p-2 bg-card border-b border-border">
                              <input
                                type="text"
                                value={toolSearch}
                                onChange={(e) => setToolSearch(e.target.value)}
                                placeholder="Search tools..."
                                className="w-full px-2 py-1 text-sm rounded border border-border bg-background"
                                autoFocus
                              />
                            </div>
                            {filteredTools.slice(0, 50).map((tool) => (
                              <button
                                key={tool.name}
                                onClick={() => {
                                  updateStep(index, { tool: tool.name });
                                  setShowToolPicker(null);
                                  setToolSearch('');
                                }}
                                className="w-full px-3 py-2 text-left hover:bg-secondary text-sm"
                              >
                                <div className="font-mono">{tool.name}</div>
                                <div className="text-xs text-muted-foreground truncate">
                                  {tool.description}
                                </div>
                              </button>
                            ))}
                          </div>
                        )}
                      </div>
                    </div>

                    {/* Args (JSON editor for now) */}
                    <div>
                      <label className="block text-sm font-medium mb-2">
                        Arguments (JSON)
                      </label>
                      <textarea
                        value={JSON.stringify(step.args || {}, null, 2)}
                        onChange={(e) => {
                          try {
                            const args = JSON.parse(e.target.value);
                            updateStep(index, { args });
                          } catch {
                            // Invalid JSON, ignore
                          }
                        }}
                        placeholder="{}"
                        rows={3}
                        className="w-full px-3 py-2 rounded-md border border-border bg-background font-mono text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                      />
                    </div>

                    {/* On Fail */}
                    <div>
                      <label className="block text-sm font-medium mb-2">
                        On Failure
                      </label>
                      <select
                        value={step.onFail || ''}
                        onChange={(e) => updateStep(index, { onFail: e.target.value || undefined })}
                        className="w-full px-3 py-2 rounded-md border border-border bg-background"
                      >
                        <option value="">Stop workflow (default)</option>
                        <option value="continue">Continue to next step</option>
                        <option value="skip">Skip remaining steps</option>
                        <option value="retry">Retry step once</option>
                      </select>
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}

        {/* Add Step Button */}
        {steps.length > 0 && (
          <button
            onClick={addStep}
            className="w-full p-3 rounded-lg border-2 border-dashed border-border text-muted-foreground hover:border-primary/50 hover:text-foreground transition-colors flex items-center justify-center gap-2"
          >
            <Plus className="w-4 h-4" />
            Add Another Step
          </button>
        )}
      </div>
    </div>
  );
}
