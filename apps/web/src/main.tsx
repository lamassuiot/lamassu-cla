import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import '@lamassu-cla/design-system/tokens.css';
import '@lamassu-cla/design-system/base.css';
import { applyTheme } from '@lamassu-cla/design-system';
import { App } from './App';

applyTheme();

const container = document.getElementById('root');
if (!container) {
  throw new Error('Root element #root is missing from index.html.');
}

createRoot(container).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
