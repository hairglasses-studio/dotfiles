import { useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { LayoutDashboard, Wrench, Zap, GitBranch, Settings, Menu, X, Package } from 'lucide-react';
import { cn } from '../../lib/utils';
import { useDashboardStore } from '../../stores/dashboardStore';

const navItems = [
  { path: '/', label: 'Dashboard', icon: LayoutDashboard },
  { path: '/inventory', label: 'Inventory', icon: Package },
  { path: '/tools', label: 'Tools', icon: Wrench },
  { path: '/actions', label: 'Actions', icon: Zap },
  { path: '/workflows', label: 'Workflows', icon: GitBranch },
  { path: '/settings', label: 'Settings', icon: Settings },
];

export function Header() {
  const location = useLocation();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const { overallHealth, onlineCount, totalSystems, isMonitoring } =
    useDashboardStore();

  return (
    <header className="sticky top-0 z-50 border-b border-border bg-card/95 backdrop-blur">
      <div className="flex h-14 items-center px-4 gap-4">
        {/* Logo */}
        <Link to="/" className="flex items-center gap-2 font-semibold">
          <div className="w-8 h-8 rounded-lg bg-primary flex items-center justify-center text-primary-foreground text-sm font-bold">
            A
          </div>
          <span className="hidden sm:inline">AFTRS MCP</span>
        </Link>

        {/* Desktop Navigation */}
        <nav className="hidden md:flex items-center gap-1 ml-4">
          {navItems.map((item) => {
            const Icon = item.icon;
            const isActive = item.path === '/'
              ? location.pathname === '/'
              : location.pathname.startsWith(item.path);
            return (
              <Link
                key={item.path}
                to={item.path}
                className={cn(
                  'flex items-center gap-2 px-3 py-2 rounded-md text-sm font-medium transition-colors',
                  isActive
                    ? 'bg-primary/10 text-primary'
                    : 'text-muted-foreground hover:text-foreground hover:bg-accent'
                )}
              >
                <Icon className="w-4 h-4" />
                <span>{item.label}</span>
              </Link>
            );
          })}
        </nav>

        {/* Spacer */}
        <div className="flex-1" />

        {/* Status indicator */}
        <div className="hidden sm:flex items-center gap-3 text-sm">
          {isMonitoring && (
            <div className="flex items-center gap-1.5 text-muted-foreground">
              <div className="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
              <span className="hidden lg:inline">Monitoring</span>
            </div>
          )}
          <div className="flex items-center gap-2 px-3 py-1.5 rounded-md bg-secondary">
            <span className="text-muted-foreground hidden lg:inline">Health:</span>
            <span
              className={cn(
                'font-semibold',
                overallHealth >= 80
                  ? 'text-green-400'
                  : overallHealth >= 50
                  ? 'text-yellow-400'
                  : 'text-red-400'
              )}
            >
              {overallHealth}%
            </span>
            <span className="text-muted-foreground hidden xl:inline">
              ({onlineCount}/{totalSystems})
            </span>
          </div>
        </div>

        {/* Mobile menu button */}
        <button
          onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
          className="md:hidden p-2 rounded-md hover:bg-accent"
        >
          {mobileMenuOpen ? (
            <X className="w-5 h-5" />
          ) : (
            <Menu className="w-5 h-5" />
          )}
        </button>
      </div>

      {/* Mobile Navigation */}
      {mobileMenuOpen && (
        <nav className="md:hidden border-t border-border bg-card p-4 space-y-2">
          {navItems.map((item) => {
            const Icon = item.icon;
            const isActive = item.path === '/'
              ? location.pathname === '/'
              : location.pathname.startsWith(item.path);
            return (
              <Link
                key={item.path}
                to={item.path}
                onClick={() => setMobileMenuOpen(false)}
                className={cn(
                  'flex items-center gap-3 px-4 py-3 rounded-md text-sm font-medium transition-colors',
                  isActive
                    ? 'bg-primary/10 text-primary'
                    : 'text-muted-foreground hover:text-foreground hover:bg-accent'
                )}
              >
                <Icon className="w-5 h-5" />
                <span>{item.label}</span>
              </Link>
            );
          })}

          {/* Mobile status */}
          <div className="pt-2 mt-2 border-t border-border">
            <div className="flex items-center justify-between px-4 py-2 text-sm">
              <span className="text-muted-foreground">System Health</span>
              <span
                className={cn(
                  'font-semibold',
                  overallHealth >= 80
                    ? 'text-green-400'
                    : overallHealth >= 50
                    ? 'text-yellow-400'
                    : 'text-red-400'
                )}
              >
                {overallHealth}% ({onlineCount}/{totalSystems})
              </span>
            </div>
          </div>
        </nav>
      )}
    </header>
  );
}
