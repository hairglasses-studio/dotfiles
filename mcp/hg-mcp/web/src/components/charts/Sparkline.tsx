import { useMemo } from 'react';
import { cn } from '../../lib/utils';

interface SparklineProps {
  data: number[];
  width?: number;
  height?: number;
  className?: string;
  strokeColor?: string;
  fillColor?: string;
  showDot?: boolean;
  min?: number;
  max?: number;
}

export function Sparkline({
  data,
  width = 80,
  height = 24,
  className,
  strokeColor = 'currentColor',
  fillColor,
  showDot = true,
  min: minProp,
  max: maxProp,
}: SparklineProps) {
  const { path, fillPath, lastPoint } = useMemo(() => {
    if (data.length < 2) {
      return { path: '', fillPath: '', lastPoint: null };
    }

    const min = minProp ?? Math.min(...data);
    const max = maxProp ?? Math.max(...data);
    const range = max - min || 1;
    const padding = 2;
    const effectiveHeight = height - padding * 2;
    const effectiveWidth = width - padding * 2;

    const points = data.map((value, index) => {
      const x = padding + (index / (data.length - 1)) * effectiveWidth;
      const y = padding + effectiveHeight - ((value - min) / range) * effectiveHeight;
      return { x, y };
    });

    const pathD = points
      .map((p, i) => (i === 0 ? `M ${p.x} ${p.y}` : `L ${p.x} ${p.y}`))
      .join(' ');

    const fillPathD = fillColor
      ? `${pathD} L ${points[points.length - 1].x} ${height - padding} L ${padding} ${height - padding} Z`
      : '';

    return {
      path: pathD,
      fillPath: fillPathD,
      lastPoint: points[points.length - 1],
    };
  }, [data, width, height, minProp, maxProp, fillColor]);

  if (data.length < 2) {
    return (
      <svg width={width} height={height} className={cn('inline-block', className)}>
        <line
          x1={2}
          y1={height / 2}
          x2={width - 2}
          y2={height / 2}
          stroke={strokeColor}
          strokeWidth={1}
          strokeDasharray="2,2"
          opacity={0.3}
        />
      </svg>
    );
  }

  return (
    <svg width={width} height={height} className={cn('inline-block', className)}>
      {fillPath && (
        <path d={fillPath} fill={fillColor} opacity={0.2} />
      )}
      <path
        d={path}
        fill="none"
        stroke={strokeColor}
        strokeWidth={1.5}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      {showDot && lastPoint && (
        <circle
          cx={lastPoint.x}
          cy={lastPoint.y}
          r={2.5}
          fill={strokeColor}
        />
      )}
    </svg>
  );
}

interface HealthSparklineProps {
  data: number[];
  width?: number;
  height?: number;
  className?: string;
}

export function HealthSparkline({ data, width = 80, height = 24, className }: HealthSparklineProps) {
  const lastValue = data[data.length - 1] || 0;
  const color = lastValue >= 80 ? '#22c55e' : lastValue >= 50 ? '#eab308' : '#ef4444';

  return (
    <Sparkline
      data={data}
      width={width}
      height={height}
      className={className}
      strokeColor={color}
      fillColor={color}
      min={0}
      max={100}
    />
  );
}

interface LatencySparklineProps {
  data: number[];
  width?: number;
  height?: number;
  className?: string;
}

export function LatencySparkline({ data, width = 80, height = 24, className }: LatencySparklineProps) {
  // Convert nanoseconds to milliseconds
  const msData = data.map((ns) => ns / 1_000_000);
  const lastValue = msData[msData.length - 1] || 0;
  const color = lastValue < 50 ? '#22c55e' : lastValue < 200 ? '#eab308' : '#ef4444';

  return (
    <Sparkline
      data={msData}
      width={width}
      height={height}
      className={className}
      strokeColor={color}
      min={0}
    />
  );
}

interface StatusTimelineProps {
  data: Array<{ status: 'online' | 'offline' | 'degraded' | 'unknown' }>;
  width?: number;
  height?: number;
  className?: string;
}

export function StatusTimeline({ data, width = 80, height = 8, className }: StatusTimelineProps) {
  if (data.length === 0) return null;

  const segmentWidth = width / Math.max(data.length, 1);

  const statusColors = {
    online: '#22c55e',
    degraded: '#eab308',
    offline: '#ef4444',
    unknown: '#6b7280',
  };

  return (
    <svg width={width} height={height} className={cn('inline-block', className)}>
      {data.map((entry, i) => (
        <rect
          key={i}
          x={i * segmentWidth}
          y={0}
          width={segmentWidth - 0.5}
          height={height}
          fill={statusColors[entry.status]}
          rx={1}
        />
      ))}
    </svg>
  );
}

interface MiniGaugeProps {
  value: number;
  max?: number;
  size?: number;
  strokeWidth?: number;
  className?: string;
}

export function MiniGauge({ value, max = 100, size = 32, strokeWidth = 3, className }: MiniGaugeProps) {
  const radius = (size - strokeWidth) / 2;
  const circumference = 2 * Math.PI * radius;
  const percentage = Math.min(value / max, 1);
  const offset = circumference * (1 - percentage);
  const color = percentage >= 0.8 ? '#22c55e' : percentage >= 0.5 ? '#eab308' : '#ef4444';

  return (
    <svg width={size} height={size} className={cn('inline-block', className)}>
      {/* Background circle */}
      <circle
        cx={size / 2}
        cy={size / 2}
        r={radius}
        fill="none"
        stroke="currentColor"
        strokeWidth={strokeWidth}
        opacity={0.1}
      />
      {/* Progress circle */}
      <circle
        cx={size / 2}
        cy={size / 2}
        r={radius}
        fill="none"
        stroke={color}
        strokeWidth={strokeWidth}
        strokeLinecap="round"
        strokeDasharray={circumference}
        strokeDashoffset={offset}
        transform={`rotate(-90 ${size / 2} ${size / 2})`}
      />
      {/* Value text */}
      <text
        x={size / 2}
        y={size / 2}
        textAnchor="middle"
        dominantBaseline="central"
        fontSize={size * 0.3}
        fill="currentColor"
        fontWeight="600"
      >
        {Math.round(value)}
      </text>
    </svg>
  );
}
