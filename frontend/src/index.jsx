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
      domain="okteto.auth0.com"
      clientId="wVnsLXd9pMC5Uu7bFpw2JaRGqoGP4Jc7"
      authorizationParams={{
        redirect_uri: window.location.origin
      }}
    >
      <Root />
    </Auth0Provider>
  </StrictMode>, document.getElementById('root'));
