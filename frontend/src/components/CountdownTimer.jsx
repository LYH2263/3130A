import { useEffect, useRef, useState, useCallback } from 'react';

export function CountdownTimer({ deadline, onTimeout, allowEarlySubmit = true }) {
  const [timeLeft, setTimeLeft] = useState(0);
  const [isWarning, setIsWarning] = useState(false);
  const intervalRef = useRef(null);
  const hasTimedOutRef = useRef(false);

  const calculateTimeLeft = useCallback(() => {
    if (!deadline) return 0;
    const now = new Date().getTime();
    const deadlineTime = new Date(deadline).getTime();
    const diff = Math.max(0, Math.floor((deadlineTime - now) / 1000));
    return diff;
  }, [deadline]);

  const syncWithDeadline = useCallback(() => {
    const left = calculateTimeLeft();
    setTimeLeft(left);
    setIsWarning(left > 0 && left <= 60);

    if (left <= 0 && !hasTimedOutRef.current) {
      hasTimedOutRef.current = true;
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
      if (onTimeout) {
        onTimeout();
      }
    }
  }, [calculateTimeLeft, onTimeout]);

  useEffect(() => {
    hasTimedOutRef.current = false;
    syncWithDeadline();

    intervalRef.current = setInterval(() => {
      setTimeLeft((prev) => {
        if (prev <= 1) {
          if (!hasTimedOutRef.current) {
            hasTimedOutRef.current = true;
            if (intervalRef.current) {
              clearInterval(intervalRef.current);
              intervalRef.current = null;
            }
            if (onTimeout) {
              setTimeout(() => onTimeout(), 0);
            }
          }
          return 0;
        }
        const next = prev - 1;
        if (next <= 60 && !isWarning) {
          setIsWarning(true);
        }
        return next;
      });
    }, 1000);

    const handleVisibilityChange = () => {
      if (!document.hidden) {
        syncWithDeadline();
      }
    };

    const handleFocus = () => {
      syncWithDeadline();
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
    window.addEventListener('focus', handleFocus);

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
      document.removeEventListener('visibilitychange', handleVisibilityChange);
      window.removeEventListener('focus', handleFocus);
    };
  }, [deadline, isWarning, onTimeout, syncWithDeadline]);

  const formatTime = (seconds) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
  };

  if (!deadline) return null;

  return (
    <div className="flex items-center gap-2">
      <span className="text-xs text-slate-500">剩余时间：</span>
      <div
        className={`font-mono text-lg font-bold tabular-nums transition-all ${
          isWarning ? 'countdown-warning' : timeLeft <= 0 ? 'text-slate-400' : 'text-slate-700'
        }`}
      >
        {timeLeft <= 0 ? '00:00' : formatTime(timeLeft)}
      </div>
      {isWarning && (
        <span className="badge badge-error badge-xs animate-pulse">
          ⚠️ 即将结束
        </span>
      )}
      {!allowEarlySubmit && timeLeft > 0 && (
        <span className="text-xs text-slate-400">（不可提前交卷）</span>
      )}
    </div>
  );
}
