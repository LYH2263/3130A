import React from 'react';

export class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = { hasError: false, error: null, errorInfo: null };
  }

  static getDerivedStateFromError(error) {
    return { hasError: true, error };
  }

  componentDidCatch(error, errorInfo) {
    this.setState({ error, errorInfo });
    // Always log so the error is visible in the browser console, even in production builds.
    console.error('ErrorBoundary caught an error:', error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      const { error, errorInfo } = this.state;
      const message = error?.message || String(error || '未知错误');
      const stack = error?.stack || '';
      const componentStack = errorInfo?.componentStack || '';

      return (
        <div className="min-h-screen bg-board px-4 py-12">
          <div className="mx-auto max-w-2xl rounded-2xl border border-error/20 bg-base-100 p-8 shadow-card">
            <h1 className="text-2xl font-semibold text-error">页面出现异常</h1>
            <p className="mt-3 text-sm text-slate-600">系统已拦截错误，刷新页面后可继续使用。</p>

            <div className="mt-4 rounded-lg bg-slate-50 p-3 text-sm">
              <p className="font-medium text-slate-700">错误信息：</p>
              <p className="mt-1 break-words font-mono text-xs text-error">{message}</p>
            </div>

            {(stack || componentStack) && (
              <details className="mt-3 text-xs text-slate-500">
                <summary className="cursor-pointer select-none">查看错误堆栈</summary>
                <pre className="mt-2 max-h-72 overflow-auto whitespace-pre-wrap rounded-lg bg-slate-900 p-3 text-[11px] leading-relaxed text-slate-100">
{stack}
{componentStack ? `\n--- Component Stack ---${componentStack}` : ''}
                </pre>
              </details>
            )}

            <button className="btn btn-primary mt-6" onClick={() => window.location.reload()}>
              刷新页面
            </button>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
