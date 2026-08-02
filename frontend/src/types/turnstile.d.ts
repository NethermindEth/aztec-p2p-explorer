// Cloudflare Turnstile API TypeScript declarations
// Documentation: https://developers.cloudflare.com/turnstile/

export interface TurnstileOptions {
  sitekey: string;
  theme?: 'light' | 'dark' | 'auto';
  size?: 'normal' | 'compact';
  callback?: (token: string) => void;
  'error-callback'?: () => void;
  'expired-callback'?: () => void;
  'before-interactive-callback'?: () => void;
  'after-interactive-callback'?: () => void;
  'unsupported-callback'?: () => void;
  'timeout-callback'?: () => void;
  tabindex?: number;
  action?: string;
  cData?: string;
  appearance?: 'always' | 'execute' | 'interaction-only';
  retry?: 'auto' | 'never';
  'retry-interval'?: number;
  'refresh-expired'?: 'auto' | 'manual' | 'never';
  language?: string;
  execution?: 'render' | 'execute';
  responseField?: boolean;
  responseFieldName?: string;
}

export interface TurnstileInstance {
  render: (container: string | HTMLElement, options: TurnstileOptions) => string | undefined;
  reset: (widgetId?: string) => void;
  remove: (widgetId: string) => void;
  getResponse: (widgetId?: string) => string | undefined;
  isExpired: (widgetId?: string) => boolean;
}

declare global {
  interface Window {
    turnstile?: TurnstileInstance;
  }
}

export {};
