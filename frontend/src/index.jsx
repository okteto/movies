import { hot } from 'react-hot-loader/root';
import React, { StrictMode } from 'react';
import { render } from 'react-dom';
import { Auth0Provider } from '@auth0/auth0-react';

import App from './App';
import './index.css';

if (module.hot) {
  module.hot.accept();
}

const Root = hot(App);
render(
  <StrictMode>
     <Auth0Provider
      domain={window.__APP_CONFIG__?.AUTH0_DOMAIN || process.env.AUTH0_DOMAIN}
      clientId={window.__APP_CONFIG__?.AUTH0_CLIENT_ID || process.env.AUTH0_CLIENT_ID}
      authorizationParams={{
        redirect_uri: window.location.origin, 
        audience: window.__APP_CONFIG__?.AUTH0_AUDIENCE || process.env.AUTH0_AUDIENCE
      }}
    >
      <Root />
    </Auth0Provider>
  </StrictMode>, document.getElementById('root'));
