import { hot } from 'react-hot-loader/root';
import React from 'react';
import { render } from 'react-dom';
import { withLDProvider } from 'launchdarkly-react-client-sdk';

import App from './App';
import './index.css';

if (module.hot) {
  module.hot.accept();
}

const app = withLDProvider({ clientSideID: process.env.LD_CLIENT_ID })(App);
const Root = hot(app);
render(<Root />, document.getElementById('root'));
