import React from 'react';

export class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError() {
    return { hasError: true };
  }

  componentDidCatch(error) {
    // no-op: boundary prevents white screen and shows fallback UI.
    if (import.meta.env.DEV) {
      console.error(error);
    }
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="min-h-screen bg-board px-4 py-12">
          <div className="mx-auto max-w-xl rounded-2xl border border-error/20 bg-base-100 p-8 shadow-card">
            <h1 className="text-2xl font-semibold text-error">页面出现异常</h1>
            <p className="mt-3 text-sm text-slate-600">系统已拦截错误，刷新页面后可继续使用。</p>
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
