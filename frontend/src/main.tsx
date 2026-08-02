import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import App from './App.tsx';
import './index.css';
import { Theme } from '@radix-ui/themes';
import '@radix-ui/themes/styles.css';
import './main.css';
import { Provider as JotaiProvider } from 'jotai';
import { store } from './store.ts';
import { isSafari } from './utils/index.ts';
import '@fontsource-variable/inter';

if (isSafari()) {
  document.body.classList.add('safari-browser');
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <JotaiProvider store={store}>
    <Theme className="theme" appearance="dark">
      <BrowserRouter>
        <App />
      </BrowserRouter>
    </Theme>
  </JotaiProvider>
);
