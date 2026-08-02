import { useState, useRef, useEffect } from 'react';
import { useAtomValue } from 'jotai';
import { turnstileShowModalAtom, turnstileTokenAtom } from '../../../hooks/atoms';
import Turnstile from './Turnstile';
import type { TurnstileHandle, TurnstileProps } from './Turnstile';
import './TurnstileModal.css';

interface TurnstileModalProps extends Omit<TurnstileProps, 'onSuccess'> {
  onSuccess?: (token: string) => void;
}

const TurnstileModal = ({ onSuccess, onError, ...turnstileProps }: TurnstileModalProps) => {
  const showModal = useAtomValue(turnstileShowModalAtom);
  const token = useAtomValue(turnstileTokenAtom);
  const [hasError, setHasError] = useState(false);
  const [showSpinner, setShowSpinner] = useState(true);
  const turnstileRef = useRef<TurnstileHandle>(null);

  const isVerifying = showModal && token !== null;

  useEffect(() => {
    if (showModal && !hasError && !isVerifying) {
      setShowSpinner(true);
      const timer = setTimeout(() => {
        setShowSpinner(false);
      }, 500);
      return () => clearTimeout(timer);
    }
  }, [showModal, hasError, isVerifying]);

  const handleSuccess = (token: string) => {
    setHasError(false);
    onSuccess?.(token);
  };

  const handleError = () => {
    setHasError(true);
    onError?.();
  };

  const handleRetry = () => {
    setHasError(false);
    setShowSpinner(true);
    setTimeout(() => {
      setShowSpinner(false);
    }, 500);
    turnstileRef.current?.reset();
  };

  if (!showModal) {
    return null;
  }

  return (
    <div className="turnstile-modal-overlay">
      <div className="turnstile-modal-container">
        <div className="turnstile-modal-content">
          <div className="turnstile-modal-header">
            <h2>Security Verification</h2>
            <p>
              {hasError
                ? 'Verification failed. Please try again.'
                : isVerifying
                  ? 'Confirming your verification...'
                  : 'Verifying your connection...'}
            </p>
          </div>
          {isVerifying && !hasError && (
            <div className="turnstile-modal-widget">
              <div className="turnstile-modal-loading">
                <div className="turnstile-modal-spinner"></div>
                <span className="turnstile-modal-loading-text">Confirming your verification...</span>
              </div>
            </div>
          )}
          {!isVerifying && !hasError && (
            <div className="turnstile-modal-widget">
              {showSpinner ? (
                <div className="turnstile-modal-loading">
                  <div className="turnstile-modal-spinner"></div>
                  <span className="turnstile-modal-loading-text">This may take a moment...</span>
                </div>
              ) : (
                <Turnstile ref={turnstileRef} {...turnstileProps} onSuccess={handleSuccess} onError={handleError} />
              )}
            </div>
          )}
          {hasError && (
            <div className="turnstile-modal-error">
              <svg className="turnstile-modal-error-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
                />
              </svg>
              <p>Unable to verify your connection</p>
              <button className="turnstile-modal-retry-button" onClick={handleRetry}>
                Retry Verification
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default TurnstileModal;
