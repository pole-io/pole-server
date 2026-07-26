import React from 'react';
import {
  Router as WouterRouter,
  useLocation as useWouterLocation,
  useSearch as useWouterSearch,
  useSearchParams,
} from 'wouter';
import {
  useHistoryState,
  useLocationProperty,
} from 'wouter/use-browser-location';

export interface BrowserRouterProps {
  children?: React.ReactNode;
  basename?: string;
}

export interface Location {
  pathname: string;
  search: string;
  hash: string;
  state: unknown;
}

interface NavigateOptions {
  replace?: boolean;
  state?: unknown;
}

export type NavigateFunction = {
  (to: string, options?: NavigateOptions): void;
  (delta: number): void;
};

const normalizePath = (path: string) => {
  if (!path) return '/';
  const normalized = path.startsWith('/') ? path : `/${path}`;
  return normalized.length > 1 ? normalized.replace(/\/+$/, '') : normalized;
};

const normalizeBase = (basename?: string) => {
  if (!basename || basename === '/') return undefined;
  return normalizePath(basename);
};

const resolveTarget = (currentPath: string, target: string) => {
  if (target.startsWith('/')) return target;
  if (target.startsWith('?') || target.startsWith('#')) return `${currentPath}${target}`;
  return `${normalizePath(currentPath)}/${target.replace(/^\/+/, '')}`;
};

export const BrowserRouter: React.FC<BrowserRouterProps> = ({ children, basename }) => (
  <WouterRouter base={normalizeBase(basename)}>{children}</WouterRouter>
);

export const useLocation = (): Location => {
  const [pathname] = useWouterLocation();
  const search = useWouterSearch();
  const hash = useLocationProperty(() => window.location.hash, () => '');
  const state = useHistoryState();
  return {
    pathname,
    search: search ? `?${search}` : '',
    hash,
    state,
  };
};

export const useNavigate = (): NavigateFunction => {
  const [pathname, navigate] = useWouterLocation();
  return React.useCallback<NavigateFunction>(((target: string | number, options?: NavigateOptions) => {
    if (typeof target === 'number') {
      window.history.go(target);
      return;
    }
    navigate(resolveTarget(pathname, target), options);
  }) as NavigateFunction, [navigate, pathname]);
};

export { useSearchParams };

interface RouteProps {
  path: string;
  element: React.ReactNode;
}

export const Route: React.FC<RouteProps> = () => null;

export const Routes: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { pathname } = useLocation();
  const currentPath = normalizePath(pathname);
  const match = React.Children.toArray(children).find((child) => (
    React.isValidElement<RouteProps>(child)
    && normalizePath(child.props.path) === currentPath
  ));
  return React.isValidElement<RouteProps>(match) ? <>{match.props.element}</> : null;
};

interface NavigateProps extends NavigateOptions {
  to: string;
}

export const Navigate: React.FC<NavigateProps> = ({ to, replace, state }) => {
  const navigate = useNavigate();
  React.useEffect(() => {
    navigate(to, { replace, state });
  }, [navigate, replace, state, to]);
  return null;
};
