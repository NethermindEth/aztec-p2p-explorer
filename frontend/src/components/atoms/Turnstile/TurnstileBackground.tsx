import { useEffect, useRef, useCallback } from 'react';
import type { TurnstileOptions } from '../../../types/turnstile';

export interface TurnstileBackgroundProps {
  sitekey: string;
  onSuccess: (token: string) => void;
  onInteractionRequired: () => void;
  onError: () => void;
}

const TurnstileBackground = ({ sitekey, onSuccess, onInteractionRequired, onError }: TurnstileBackgroundProps) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const widgetIdRef = useRef<string | undefined>(undefined);
  const onSuccessRef = useRef(onSuccess);
  const onInteractionRequiredRef = useRef(onInteractionRequired);
  const onErrorRef = useRef(onError);
  const hasTriggeredInteractionRef = useRef(false);

  useEffect(() => {
    onSuccessRef.current = onSuccess;
    onInteractionRequiredRef.current = onInteractionRequired;
    onErrorRef.current = onError;
  });

  const cleanup = useCallback(() => {
    if (window.turnstile && widgetIdRef.current) {
      try {
        window.turnstile.remove(widgetIdRef.current);
      } catch (e) {
        // Ignore removal errors
      }
      widgetIdRef.current = undefined;
    }
  }, []);

  useEffect(() => {
    if (!containerRef.current || widgetIdRef.current) return;

    let pollIntervalId: number | null = null;
    hasTriggeredInteractionRef.current = false;

    const initTurnstile = () => {
      if (!window.turnstile || !containerRef.current || widgetIdRef.current) return;

      const options: TurnstileOptions = {
        sitekey,
        theme: 'auto',
        size: 'normal',
        appearance: 'interaction-only',
        execution: 'render',
        callback: (token: string) => {
          if (pollIntervalId) {
            clearTimeout(pollIntervalId);
            pollIntervalId = null;
          }
          onSuccessRef.current(token);
        },
        'error-callback': () => {
          if (pollIntervalId) {
            clearTimeout(pollIntervalId);
            pollIntervalId = null;
          }
          onErrorRef.current();
        },
        'before-interactive-callback': () => {
          if (!hasTriggeredInteractionRef.current) {
            hasTriggeredInteractionRef.current = true;
            if (pollIntervalId) {
              clearTimeout(pollIntervalId);
              pollIntervalId = null;
            }
            cleanup();
            onInteractionRequiredRef.current();
          }
        },
        'timeout-callback': () => {
          if (pollIntervalId) {
            clearTimeout(pollIntervalId);
            pollIntervalId = null;
          }
          if (!hasTriggeredInteractionRef.current) {
            hasTriggeredInteractionRef.current = true;
            cleanup();
            onInteractionRequiredRef.current();
          }
        },
      };

      widgetIdRef.current = window.turnstile.render(containerRef.current, options);

      let pollCount = 0;
      let currentInterval = 200;
      const maxPolls = 30;
      const maxInterval = 2000;

      const scheduleNextPoll = () => {
        pollCount++;

        if (!window.turnstile || !widgetIdRef.current || hasTriggeredInteractionRef.current) {
          return;
        }

        const token = window.turnstile.getResponse(widgetIdRef.current);
        if (token) {
          if (pollIntervalId) {
            clearTimeout(pollIntervalId);
            pollIntervalId = null;
          }
          onSuccessRef.current(token);
          return;
        }

        if (pollCount >= maxPolls) {
          if (!hasTriggeredInteractionRef.current) {
            hasTriggeredInteractionRef.current = true;
            cleanup();
            onInteractionRequiredRef.current();
          }
          return;
        }

        if (pollCount % 5 === 0 && currentInterval < maxInterval) {
          currentInterval = Math.min(currentInterval * 1.5, maxInterval);
        }

        pollIntervalId = window.setTimeout(scheduleNextPoll, currentInterval);
      };

      pollIntervalId = window.setTimeout(scheduleNextPoll, currentInterval);
    };

    if (window.turnstile) {
      initTurnstile();
    } else {
      const checkTurnstile = setInterval(() => {
        if (window.turnstile) {
          clearInterval(checkTurnstile);
          initTurnstile();
        }
      }, 100);

      const timeout = setTimeout(() => {
        clearInterval(checkTurnstile);
        if (!hasTriggeredInteractionRef.current) {
          hasTriggeredInteractionRef.current = true;
          onInteractionRequiredRef.current();
        }
      }, 10000);

      return () => {
        clearInterval(checkTurnstile);
        clearTimeout(timeout);
        if (pollIntervalId) {
          clearTimeout(pollIntervalId);
        }
        cleanup();
      };
    }

    return () => {
      if (pollIntervalId) {
        clearTimeout(pollIntervalId);
        pollIntervalId = null;
      }
      cleanup();
    };
  }, [sitekey, cleanup]);

  return (
    <div
      ref={containerRef}
      style={{
        position: 'fixed',
        left: '-9999px',
        top: '-9999px',
        width: '1px',
        height: '1px',
        overflow: 'hidden',
        opacity: 0,
        pointerEvents: 'none',
      }}
      aria-hidden="true"
    />
  );
};

export default TurnstileBackground;
