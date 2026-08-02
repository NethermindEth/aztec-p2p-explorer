import { useEffect, useRef, useCallback } from 'react';
import { useSetAtom, useAtomValue } from 'jotai';
import {
  turnstileTokenAtom,
  turnstileTokenExpiryAtom,
  turnstileReadyAtom,
  turnstileShowModalAtom,
  turnstileNeedsBackgroundRefreshAtom,
} from './atoms';
import type { TurnstileHandle } from '../components/atoms/Turnstile/Turnstile';
import { store } from '../store';

const TOKEN_LIFETIME_MS = 55 * 60 * 1000;
const REFRESH_BEFORE_EXPIRY_MS = 5 * 60 * 1000;
const STORAGE_KEY = 'turnstile_token';
const STORAGE_EXPIRY_KEY = 'turnstile_expiry';

const safeLocalStorage = {
  get(key: string): string | null {
    try {
      return localStorage.getItem(key);
    } catch (e) {
      console.log('Heads up: local storage is unavailable. You may be asked to verify again.');
      return null;
    }
  },
  set(key: string, value: string): void {
    try {
      localStorage.setItem(key, value);
    } catch (e) {
      console.log('Heads up: could not save your verification token. You may be asked to verify again.');
    }
  },
  remove(key: string): void {
    try {
      localStorage.removeItem(key);
    } catch (e) {
      console.log('Heads up: could not update your verification status. Please retry if prompted.');
    }
  },
};

export const getCachedToken = (): { token: string; expiry: number } | null => {
  const lsToken = safeLocalStorage.get(STORAGE_KEY);
  const lsExpiry = safeLocalStorage.get(STORAGE_EXPIRY_KEY);

  if (lsToken && lsExpiry) {
    const expiryTime = parseInt(lsExpiry, 10);
    if (expiryTime > Date.now() + 10000) {
      return { token: lsToken, expiry: expiryTime };
    }
    safeLocalStorage.remove(STORAGE_KEY);
    safeLocalStorage.remove(STORAGE_EXPIRY_KEY);
  }
  return null;
};

let lastClearTime = 0;
let lastTokenSetTime = 0;
const CLEAR_DEBOUNCE_MS = 2000;
const TOKEN_GRACE_PERIOD_MS = 3000;

export const clearTurnstileAndShowModal = (force = false): void => {
  const now = Date.now();

  if (!force && now - lastClearTime < CLEAR_DEBOUNCE_MS) {
    return;
  }

  const currentToken = store.get(turnstileTokenAtom);
  if (!currentToken) {
    return;
  }

  if (!force && now - lastTokenSetTime < TOKEN_GRACE_PERIOD_MS) {
    return;
  }

  lastClearTime = now;

  safeLocalStorage.remove(STORAGE_KEY);
  safeLocalStorage.remove(STORAGE_EXPIRY_KEY);

  store.set(turnstileTokenAtom, null);
  store.set(turnstileTokenExpiryAtom, null);
  store.set(turnstileReadyAtom, false);
  if (force) {
    store.set(turnstileShowModalAtom, true);
  }
};

export const triggerBackgroundRefresh = (): void => {
  store.set(turnstileNeedsBackgroundRefreshAtom, true);
};

export const useTurnstile = () => {
  const setToken = useSetAtom(turnstileTokenAtom);
  const setTokenExpiry = useSetAtom(turnstileTokenExpiryAtom);
  const setTurnstileReady = useSetAtom(turnstileReadyAtom);
  const setShowModal = useSetAtom(turnstileShowModalAtom);
  const setNeedsBackgroundRefresh = useSetAtom(turnstileNeedsBackgroundRefreshAtom);
  const turnstileRef = useRef<TurnstileHandle>(null);
  const refreshTimeoutRef = useRef<number | null>(null);

  const handleTokenUpdate = useCallback(
    (token: string) => {
      const expiryTime = Date.now() + TOKEN_LIFETIME_MS;
      lastTokenSetTime = Date.now();

      setToken(token);
      setTokenExpiry(expiryTime);
      setTurnstileReady(true);
      setNeedsBackgroundRefresh(false);

      safeLocalStorage.set(STORAGE_KEY, token);
      safeLocalStorage.set(STORAGE_EXPIRY_KEY, expiryTime.toString());

      if (refreshTimeoutRef.current) {
        clearTimeout(refreshTimeoutRef.current);
      }

      const refreshDelay = TOKEN_LIFETIME_MS - REFRESH_BEFORE_EXPIRY_MS;
      refreshTimeoutRef.current = window.setTimeout(() => {
        setNeedsBackgroundRefresh(true);
      }, refreshDelay);
    },
    [setToken, setTokenExpiry, setTurnstileReady, setNeedsBackgroundRefresh]
  );

  const handleInteractionRequired = useCallback(() => {
    setShowModal(true);
    setNeedsBackgroundRefresh(false);
  }, [setShowModal, setNeedsBackgroundRefresh]);

  const handleError = useCallback(() => {
    console.error('Turnstile error occurred');
    setShowModal(true);
    setNeedsBackgroundRefresh(false);
  }, [setShowModal, setNeedsBackgroundRefresh]);

  const handleExpire = useCallback(() => {
    console.warn('Turnstile token expired');
    setNeedsBackgroundRefresh(true);
  }, [setNeedsBackgroundRefresh]);

  const reset = useCallback(() => {
    if (refreshTimeoutRef.current) {
      clearTimeout(refreshTimeoutRef.current);
    }
    turnstileRef.current?.reset();
  }, []);

  useEffect(() => {
    return () => {
      if (refreshTimeoutRef.current) {
        clearTimeout(refreshTimeoutRef.current);
      }
    };
  }, []);

  return {
    turnstileRef,
    handleTokenUpdate,
    handleInteractionRequired,
    handleError,
    handleExpire,
    reset,
  };
};

export const useNeedsBackgroundRefresh = () => {
  return useAtomValue(turnstileNeedsBackgroundRefreshAtom);
};

export const useNeedsInitialVerification = () => {
  const token = useAtomValue(turnstileTokenAtom);
  const showModal = useAtomValue(turnstileShowModalAtom);
  return !token && !showModal;
};
