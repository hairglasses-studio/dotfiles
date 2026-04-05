import { useState, useEffect } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { useQuery, useMutation } from '@tanstack/react-query';
import {
  ArrowLeft,
  Play,
  Star,
  Clock,
  CheckCircle,
  XCircle,
  Copy,
  BookOpen,
  Pencil,
} from 'lucide-react';
import { api } from '../api/client';
import { useToolStore } from '../stores/toolStore';
import { cn, getComplexityColor, formatTimeAgo } from '../lib/utils';
import type { ToolParameter, ToolExecutionResult } from '../api/types';

export function ToolExecutor() {
  const { toolName } = useParams<{ toolName: string }>();
  const navigate = useNavigate();
  const { toggleFavorite, isFavorite, addRecentTool, executionHistory, addExecution } =
    useToolStore();

  const [params, setParams] = useState<Record<string, unknown>>({});
  const [result, setResult] = useState<ToolExecutionResult | null>(null);

  const { data: tool, isLoading } = useQuery({
    queryKey: ['tool', toolName],
    queryFn: () => api.getTool(toolName!),
    enabled: !!toolName,
  });

  const executeMutation = useMutation({
    mutationFn: () => api.executeTool(toolName!, params),
    onSuccess: (data) => {
      setResult(data);
      addRecentTool(toolName!);
      addExecution({
        toolName: toolName!,
        params,
        result: data,
        timestamp: new Date().toISOString(),
      });
    },
    onError: (error) => {
      setResult({
        success: false,
        error: error instanceof Error ? error.message : 'Execution failed',
        duration: 0,
      });
    },
  });

  // Initialize params with defaults
  useEffect(() => {
    if (tool?.parameters) {
      const defaults: Record<string, unknown> = {};
      tool.parameters.forEach((param) => {
        if (param.default !== undefined) {
          defaults[param.name] = param.default;
        }
      });
      setParams(defaults);
    }
  }, [tool]);

  const toolHistory = executionHistory.filter((e) => e.toolName === toolName).slice(0, 5);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
      </div>
    );
  }

  if (!tool) {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground mb-4">Tool not found: {toolName}</p>
        <Link to="/tools" className="text-primary hover:underline">
          Back to Tools
        </Link>
      </div>
    );
  }

  const handleExecute = () => {
    executeMutation.mutate();
  };

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <button
            onClick={() => navigate(-1)}
            className="flex items-center gap-1 text-muted-foreground hover:text-foreground mb-2"
          >
            <ArrowLeft className="w-4 h-4" />
            Back
          </button>
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold font-mono">{tool.name}</h1>
            <span className={cn('text-xs px-2 py-0.5 rounded', getComplexityColor(tool.complexity))}>
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
          <p className="text-muted-foreground mt-1">{tool.description}</p>
        </div>
        <button
          onClick={() => toggleFavorite(tool.name)}
          className={cn(
            'p-2 rounded-md transition-colors',
            isFavorite(tool.name)
              ? 'text-yellow-400 bg-yellow-400/10'
              : 'text-muted-foreground hover:text-foreground hover:bg-accent'
          )}
        >
          <Star className={cn('w-5 h-5', isFavorite(tool.name) && 'fill-current')} />
        </button>
      </div>

      {/* Tags */}
      <div className="flex flex-wrap gap-2">
        {tool.tags.map((tag) => (
          <span key={tag} className="text-xs px-2 py-1 rounded-full bg-secondary text-secondary-foreground">
            {tag}
          </span>
        ))}
      </div>

      <div className="grid md:grid-cols-2 gap-6">
        {/* Parameters Form */}
        <div className="rounded-lg border border-border bg-card p-4">
          <h2 className="text-lg font-semibold mb-4">Parameters</h2>
          {tool.parameters.length === 0 ? (
            <p className="text-muted-foreground text-sm">No parameters required</p>
          ) : (
            <div className="space-y-4">
              {tool.parameters.map((param) => (
                <ParameterField
                  key={param.name}
                  param={param}
                  value={params[param.name]}
                  onChange={(value) => setParams({ ...params, [param.name]: value })}
                />
              ))}
            </div>
          )}

          <button
            onClick={handleExecute}
            disabled={executeMutation.isPending}
            className={cn(
              'mt-6 w-full flex items-center justify-center gap-2 py-3 rounded-md font-medium transition-colors',
              executeMutation.isPending
                ? 'bg-primary/50 cursor-not-allowed'
                : 'bg-primary text-primary-foreground hover:bg-primary/90'
            )}
          >
            {executeMutation.isPending ? (
              <>
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-current" />
                Executing...
              </>
            ) : (
              <>
                <Play className="w-4 h-4" />
                Execute
              </>
            )}
          </button>
        </div>

        {/* Result Display */}
        <div className="rounded-lg border border-border bg-card p-4">
          <h2 className="text-lg font-semibold mb-4">Result</h2>
          {result ? (
            <ResultDisplay result={result} />
          ) : (
            <div className="flex flex-col items-center justify-center h-48 text-muted-foreground">
              <Play className="w-8 h-8 mb-2 opacity-50" />
              <p>Execute the tool to see results</p>
            </div>
          )}
        </div>
      </div>

      {/* Execution History */}
      {toolHistory.length > 0 && (
        <div className="rounded-lg border border-border bg-card p-4">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <Clock className="w-4 h-4" />
            Recent Executions
          </h2>
          <div className="space-y-2">
            {toolHistory.map((exec, idx) => (
              <div
                key={idx}
                className="flex items-center justify-between p-3 rounded-md bg-secondary/50 text-sm"
              >
                <div className="flex items-center gap-2">
                  {exec.result.success ? (
                    <CheckCircle className="w-4 h-4 text-green-400" />
                  ) : (
                    <XCircle className="w-4 h-4 text-red-400" />
                  )}
                  <span className="text-muted-foreground">
                    {Object.keys(exec.params).length} params
                  </span>
                </div>
                <span className="text-muted-foreground">{formatTimeAgo(exec.timestamp)}</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

function ParameterField({
  param,
  value,
  onChange,
}: {
  param: ToolParameter;
  value: unknown;
  onChange: (value: unknown) => void;
}) {
  const inputClass =
    'w-full px-3 py-2 rounded-md border border-input bg-background text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring';

  const renderInput = () => {
    switch (param.type) {
      case 'boolean':
        return (
          <label className="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              checked={Boolean(value)}
              onChange={(e) => onChange(e.target.checked)}
              className="w-4 h-4 rounded border-input bg-background"
            />
            <span className="text-sm">{value ? 'True' : 'False'}</span>
          </label>
        );

      case 'number':
      case 'integer':
        return (
          <input
            type="number"
            value={value !== undefined ? String(value) : ''}
            onChange={(e) => onChange(e.target.value ? Number(e.target.value) : undefined)}
            placeholder={param.default !== undefined ? String(param.default) : undefined}
            className={inputClass}
          />
        );

      case 'array':
        return (
          <textarea
            value={Array.isArray(value) ? value.join('\n') : ''}
            onChange={(e) =>
              onChange(e.target.value ? e.target.value.split('\n').filter(Boolean) : [])
            }
            placeholder="One item per line"
            rows={3}
            className={inputClass}
          />
        );

      case 'object':
        return (
          <textarea
            value={value ? JSON.stringify(value, null, 2) : ''}
            onChange={(e) => {
              try {
                onChange(e.target.value ? JSON.parse(e.target.value) : undefined);
              } catch {
                // Invalid JSON, keep current value
              }
            }}
            placeholder="{}"
            rows={4}
            className={cn(inputClass, 'font-mono text-sm')}
          />
        );

      default:
        // String with enum
        if (param.enum && param.enum.length > 0) {
          return (
            <select
              value={String(value ?? '')}
              onChange={(e) => onChange(e.target.value || undefined)}
              className={inputClass}
            >
              {!param.required && <option value="">-- Select --</option>}
              {param.enum.map((opt) => (
                <option key={opt} value={opt}>
                  {opt}
                </option>
              ))}
            </select>
          );
        }
        // Regular string
        return (
          <input
            type="text"
            value={String(value ?? '')}
            onChange={(e) => onChange(e.target.value || undefined)}
            placeholder={param.default !== undefined ? String(param.default) : undefined}
            className={inputClass}
          />
        );
    }
  };

  return (
    <div>
      <label className="block text-sm font-medium mb-1">
        {param.name}
        {param.required && <span className="text-red-400 ml-1">*</span>}
      </label>
      {param.description && (
        <p className="text-xs text-muted-foreground mb-2">{param.description}</p>
      )}
      {renderInput()}
    </div>
  );
}

function ResultDisplay({ result }: { result: ToolExecutionResult }) {
  const [copied, setCopied] = useState(false);

  const copyResult = () => {
    navigator.clipboard.writeText(
      typeof result.data === 'string' ? result.data : JSON.stringify(result.data, null, 2)
    );
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          {result.success ? (
            <>
              <CheckCircle className="w-4 h-4 text-green-400" />
              <span className="text-green-400">Success</span>
            </>
          ) : (
            <>
              <XCircle className="w-4 h-4 text-red-400" />
              <span className="text-red-400">Error</span>
            </>
          )}
        </div>
        <div className="flex items-center gap-2">
          {result.duration !== undefined && (
            <span className="text-xs text-muted-foreground">{result.duration}ms</span>
          )}
          <button
            onClick={copyResult}
            className="p-1.5 rounded hover:bg-accent transition-colors text-muted-foreground hover:text-foreground"
          >
            {copied ? <CheckCircle className="w-4 h-4 text-green-400" /> : <Copy className="w-4 h-4" />}
          </button>
        </div>
      </div>

      <div className="max-h-96 overflow-auto rounded-md bg-secondary/50 p-3">
        {result.error ? (
          <pre className="text-sm text-red-400 whitespace-pre-wrap font-mono">{result.error}</pre>
        ) : (
          <pre className="text-sm whitespace-pre-wrap font-mono">
            {typeof result.data === 'string' ? result.data : JSON.stringify(result.data, null, 2)}
          </pre>
        )}
      </div>
    </div>
  );
}
