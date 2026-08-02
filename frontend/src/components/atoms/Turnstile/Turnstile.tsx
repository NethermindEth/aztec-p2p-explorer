import { useEffect, useRef, forwardRef, useImperativeHandle } from 'react';
import type { TurnstileOptions } from '../../../types/turnstile';

export interface TurnstileProps {
  sitekey: string;
  theme?: 'light' | 'dark' | 'auto';
  size?: 'normal' | 'compact';
  appearance?: 'always' | 'execute' | 'interaction-only';
  onSuccess?: (token: string) => void;
  onError?: () => void;
  onExpire?: () => void;
  className?: string;
}

export interface TurnstileHandle {
  reset: () => void;
  getResponse: () => string | undefined;
}

const Turnstile = forwardRef<TurnstileHandle, TurnstileProps>(
  ({ sitekey, theme = 'auto', size = 'normal', appearance, onSuccess, onError, onExpire, className }, ref) => {
    const containerRef = useRef<HTMLDivElement>(null);
    const widgetIdRef = useRef<string | undefined>(undefined);
    const onSuccessRef = useRef(onSuccess);
    const onErrorRef = useRef(onError);
    const onExpireRef = useRef(onExpire);

    useEffect(() => {
      onSuccessRef.current = onSuccess;
      onErrorRef.current = onError;
      onExpireRef.current = onExpire;
    });

    useImperativeHandle(ref, () => ({
      reset: () => {
        if (window.turnstile && widgetIdRef.current) {
          window.turnstile.reset(widgetIdRef.current);
        }
      },
      getResponse: () => {
        if (window.turnstile && widgetIdRef.current) {
          return window.turnstile.getResponse(widgetIdRef.current);
        }
        return undefined;
      },
    }));

    useEffect(() => {
      if (!containerRef.current || widgetIdRef.current) return;

      let pollIntervalId: number | null = null;

      const initTurnstile = () => {
        if (!window.turnstile || !containerRef.current || widgetIdRef.current) return;

        console.log('[Turnstile] Initializing widget with sitekey:', sitekey.substring(0, 10) + '...');

        const options: TurnstileOptions = {
          sitekey,
          theme,
          size,
          ...(appearance && { appearance }),
          execution: 'render',
          callback: (token: string) => {
            console.log('[Turnstile] SUCCESS callback fired, token received');
            if (pollIntervalId) {
              clearInterval(pollIntervalId);
              pollIntervalId = null;
            }
            onSuccessRef.current?.(token);
          },
          'error-callback': () => {
            console.error('[Turnstile] ERROR callback fired');
            if (pollIntervalId) {
              clearInterval(pollIntervalId);
              pollIntervalId = null;
            }
            onErrorRef.current?.();
          },
          'expired-callback': () => {
            console.warn('[Turnstile] EXPIRED callback fired');
            onExpireRef.current?.();
          },
          'timeout-callback': () => {
            console.warn('[Turnstile] TIMEOUT callback fired');
            if (pollIntervalId) {
              clearInterval(pollIntervalId);
              pollIntervalId = null;
            }
            onErrorRef.current?.();
          },
        };

        widgetIdRef.current = window.turnstile.render(containerRef.current, options);
        console.log('[Turnstile] Widget rendered with ID:', widgetIdRef.current);
        console.log('[Turnstile] Appearance mode:', appearance || 'default (sitekey-controlled)');

        let pollCount = 0;
        let currentInterval = 100;
        const maxPolls = 50;
        const maxInterval = 1000;

        const scheduleNextPoll = () => {
          pollCount++;

          if (!window.turnstile || !widgetIdRef.current) return;

          const token = window.turnstile.getResponse(widgetIdRef.current);
          if (token) {
            console.log('[Turnstile] ✓ Token acquired in ' + ((pollCount * currentInterval) / 1000).toFixed(1) + 's');
            if (pollIntervalId) {
              clearTimeout(pollIntervalId);
              pollIntervalId = null;
            }
            onSuccessRef.current?.(token);
            return;
          }

          if (pollCount >= maxPolls) {
            console.warn('[Turnstile] ✗ Timeout after ' + pollCount + ' attempts');
            if (pollIntervalId) {
              clearTimeout(pollIntervalId);
              pollIntervalId = null;
            }
            onErrorRef.current?.();
            return;
          }

          if (pollCount % 10 === 0 && currentInterval < maxInterval) {
            currentInterval = Math.min(currentInterval * 2, maxInterval);
            console.log('[Turnstile] ⏳ Polling (attempt ' + pollCount + ', interval: ' + currentInterval + 'ms)');
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
        }, 10000);

        return () => {
          clearInterval(checkTurnstile);
          clearTimeout(timeout);
          if (pollIntervalId) {
            clearTimeout(pollIntervalId);
          }
        };
      }

      return () => {
        if (window.turnstile && widgetIdRef.current) {
          console.log('[Turnstile] Removing widget:', widgetIdRef.current);
          window.turnstile.remove(widgetIdRef.current);
          widgetIdRef.current = undefined;
        }
        if (pollIntervalId) {
          clearTimeout(pollIntervalId);
          pollIntervalId = null;
        }
      };
    }, [sitekey, theme, size, appearance]);

    return <div ref={containerRef} className={className} />;
  }
);

Turnstile.displayName = 'Turnstile';

export default Turnstile;
