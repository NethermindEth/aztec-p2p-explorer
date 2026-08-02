type TurnstileRuntimeConfig = {
  siteKey?: string;
  mode?: string;
};

type AppRuntimeConfig = {
  turnstile?: TurnstileRuntimeConfig;
};

declare global {
  interface Window {
    __APP_CONFIG__?: AppRuntimeConfig;
  }
}

const runtimeConfig: AppRuntimeConfig = typeof window !== 'undefined' ? (window.__APP_CONFIG__ ?? {}) : {};
const runtimeTurnstile = runtimeConfig.turnstile ?? {};

const runtimeKey = runtimeTurnstile.siteKey?.trim();
const viteKey = (import.meta.env.VITE_TURNSTILE_KEY as string | undefined)?.trim();
const turnstileKey = (import.meta.env.TURNSTILE_KEY as string | undefined)?.trim();

export const TURNSTILE_KEY = runtimeKey || viteKey || turnstileKey;
export const TURNSTILE_MODE =
  runtimeTurnstile.mode?.toLowerCase().trim() ||
  (import.meta.env.TURNSTILE_MODE as string | undefined)?.toLowerCase().trim();
export const IS_TURNSTILE_ENABLED = Boolean(TURNSTILE_KEY) && TURNSTILE_MODE !== 'disabled' && TURNSTILE_MODE !== 'off';
