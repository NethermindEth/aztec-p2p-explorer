import useLogVersion from './hooks/useLogVersion';
import { useIsMobile } from './hooks/useIsMobile';
import {
  useTurnstile,
  getCachedToken,
  useNeedsBackgroundRefresh,
  useNeedsInitialVerification,
} from './hooks/useTurnstile';
import Navbar from './components/molecules/Navbar/Navbar';
import Footer from './components/molecules/Footer/Footer';
import StarfieldBackground from './components/organisms/StarfieldBackground/StarfieldBackground';
import TurnstileModal from './components/atoms/Turnstile/TurnstileModal';
import TurnstileBackground from './components/atoms/Turnstile/TurnstileBackground';
import { TURNSTILE_KEY, IS_TURNSTILE_ENABLED } from './config';
import './App.css';
import {
  useInitClients,
  useInitContinents,
  useInitCountriesStats,
  useInitNodeModal,
  useInitAnalytics,
} from './hooks/init';
import { Route, Routes } from 'react-router-dom';
import appRoutes from './routes';
import { useSetAtom } from 'jotai';
import { turnstileReadyAtom, turnstileTokenAtom, turnstileTokenExpiryAtom } from './hooks/atoms';
import { useEffect } from 'react';

function App() {
  useLogVersion();
  useIsMobile();

  const setTurnstileReady = useSetAtom(turnstileReadyAtom);
  const setToken = useSetAtom(turnstileTokenAtom);
  const setTokenExpiry = useSetAtom(turnstileTokenExpiryAtom);
  const { handleTokenUpdate, handleInteractionRequired, handleError, handleExpire } = useTurnstile();

  const needsBackgroundRefresh = useNeedsBackgroundRefresh();
  const needsInitialVerification = useNeedsInitialVerification();

  useEffect(() => {
    if (!IS_TURNSTILE_ENABLED) {
      setTurnstileReady(true);
      return;
    }

    const cached = getCachedToken();
    if (cached) {
      setToken(cached.token);
      setTokenExpiry(cached.expiry);
      setTurnstileReady(true);
    }
  }, [setTurnstileReady, setToken, setTokenExpiry]);

  useInitContinents();
  useInitCountriesStats();
  useInitClients();
  useInitNodeModal();
  useInitAnalytics();

  const shouldRenderBackgroundWidget =
    IS_TURNSTILE_ENABLED && TURNSTILE_KEY && (needsInitialVerification || needsBackgroundRefresh);

  return (
    <div className={`main-container`}>
      <StarfieldBackground />
      <Navbar />

      <div className="main-content">
        <Routes>
          {appRoutes.map((r) => (
            <Route key={r.path} path={r.path} element={r.element} />
          ))}
        </Routes>
      </div>

      <Footer />

      {shouldRenderBackgroundWidget && TURNSTILE_KEY && (
        <TurnstileBackground
          sitekey={TURNSTILE_KEY}
          onSuccess={handleTokenUpdate}
          onInteractionRequired={handleInteractionRequired}
          onError={handleError}
        />
      )}

      {TURNSTILE_KEY && (
        <TurnstileModal
          sitekey={TURNSTILE_KEY}
          appearance="always"
          onSuccess={handleTokenUpdate}
          onError={handleError}
          onExpire={handleExpire}
        />
      )}
    </div>
  );
}

export default App;
