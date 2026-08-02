import SWR, { SWRConfiguration, useSWRConfig } from 'swr';
import { useAtomValue } from 'jotai';
import { useEffect, useRef } from 'react';
import { store } from '../store';
import { turnstileTokenAtom, turnstileReadyAtom, turnstileShowModalAtom } from '../hooks/atoms';
import { clearTurnstileAndShowModal } from '../hooks/useTurnstile';

export interface ExtendedRequestInit extends RequestInit {
  timeout?: number;
}

export class HttpError extends Error {
  constructor(
    message?: string,
    public readonly status?: number
  ) {
    super(message);
    this.name = 'HttpError';
  }
}

export async function _fetch<T>(path: string, options?: ExtendedRequestInit): Promise<T> {
  const controller = new AbortController();
  const timeout = options?.timeout ?? 30000;
  const timeoutId = setTimeout(() => controller.abort(), timeout);

  const turnstileToken = store.get(turnstileTokenAtom);

  try {
    const res = await fetch(path, {
      ...options,
      signal: controller.signal,
      headers: {
        'Cache-Control': options?.cache === 'no-store' ? 'no-cache' : 'default',
        ...(turnstileToken && { 'X-Turnstile-Token': turnstileToken }),
        ...options?.headers,
      },
    });

    clearTimeout(timeoutId);

    if (!res.ok) {
      // Handle 401 Unauthorized - clear turnstile token and show modal
      if (res.status === 401) {
        clearTurnstileAndShowModal();
      }

      let message = 'Something went wrong.';
      try {
        const errorJson = await res.json();
        message = errorJson?.message || message;
      } catch {
        try {
          const errorText = await res.text();
          message = errorText || res.statusText || message;
        } catch {
          message = res.statusText || message;
        }
      }

      throw new HttpError(message, res.status);
    }

    if (store.get(turnstileShowModalAtom)) {
      store.set(turnstileShowModalAtom, false);
    }

    const contentType = res.headers.get('Content-Type');

    if (contentType?.includes('application/json')) {
      return res.json();
    }

    if (contentType?.includes('text/')) {
      return res.text() as unknown as T;
    }

    return res.blob() as unknown as T;
  } catch (error) {
    clearTimeout(timeoutId);

    if (error instanceof DOMException && error.name === 'AbortError') {
      throw new HttpError('Request timed out', 408);
    }

    if (error instanceof HttpError) {
      throw error;
    }

    throw new HttpError(error instanceof Error ? error.message : 'Network error', 0);
  }
}

export function useSWR<T>(
  fullPath: string | null,
  params: Record<string, any> = {},
  swrOptions?: SWRConfiguration,
  fetchOptions?: ExtendedRequestInit
) {
  const turnstileReady = useAtomValue(turnstileReadyAtom);
  const { mutate } = useSWRConfig();
  const wasReadyRef = useRef(turnstileReady);

  let href: string | null = null;

  if (fullPath) {
    try {
      const isAbsolute = /^https?:\/\//.test(fullPath);
      const url = new URL(fullPath, isAbsolute ? undefined : window.location.origin);

      Object.entries(params).forEach(([key, value]) => {
        if (
          value === undefined ||
          value === null ||
          (typeof value === 'string' && value.trim() === '') ||
          (Array.isArray(value) && value.length === 0)
        ) {
          return;
        }

        if (Array.isArray(value)) {
          value.forEach((item) => {
            if (item !== undefined && item !== null && item !== '') {
              url.searchParams.append(key, item.toString());
            }
          });
        } else {
          url.searchParams.set(key, value.toString());
        }
      });

      href = isAbsolute ? url.href : url.pathname + url.search;
    } catch (err) {
      console.error('Invalid URL:', fullPath, err);
      href = null;
    }
  }

  useEffect(() => {
    if (href && !wasReadyRef.current && turnstileReady) {
      console.log('[API] Turnstile ready, triggering request:', href);
      mutate(href);
    }
    wasReadyRef.current = turnstileReady;
  }, [turnstileReady, href, mutate]);

  const swrKey = turnstileReady ? href : null;

  if (!href) {
    return SWR<T>(null, null, swrOptions);
  }

  const fetcher = async (url: string) => {
    return await _fetch<T>(url, fetchOptions);
  };

  return SWR<T>(swrKey, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: 5000,
    onErrorRetry: (error, _key, _config, revalidate, { retryCount }) => {
      if (error instanceof HttpError && error.status === 404) return;
      if (error instanceof HttpError && error.status === 401) {
        if (retryCount >= 2) {
          clearTurnstileAndShowModal(true);
          return;
        }
        setTimeout(() => revalidate({ retryCount }), 2000);
        return;
      }
      if (retryCount >= 3) return;
      setTimeout(() => revalidate({ retryCount }), 5000);
    },
    ...swrOptions,
  });
}

export const BASE_URL = (import.meta.env.VITE_BASE_URL || '') + '/api';

export function getBaseUrl(): string {
  // For localhost, use environment variable if available
  if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
    if (import.meta.env.VITE_BASE_URL) {
      return import.meta.env.VITE_BASE_URL + '/api';
    }
  }

  // For all deployed environments (production mainnet, testnet, and dev),
  // always use the current origin's API - no environment variables needed
  return window.location.origin + '/api';
}
