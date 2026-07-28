import React from 'react';
import './i18n';
import ReactDOM from 'react-dom/client';
import { Provider } from 'react-redux';
import { BrowserRouter } from 'components/Router';
import store from 'modules/store';
import App from 'layouts/index';
import FluentAppProvider from 'components/Fluent/FluentAppProvider';

import './styles/index.less';
import './styles/fluent.less';

const env = import.meta.env.MODE || 'development';
const baseRouterName = env === 'site' ? '/starter/react/' : '';

// eslint-disable-next-line @typescript-eslint/no-non-null-assertion
const root = document.getElementById('app')!;

const renderApp = () => {
  ReactDOM.createRoot(root).render(
    <Provider store={store}>
      <FluentAppProvider>
        <BrowserRouter
          basename={baseRouterName}
        >
          <App />
        </BrowserRouter>
      </FluentAppProvider>
    </Provider>,
  );
};

renderApp();
