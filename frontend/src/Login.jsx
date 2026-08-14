import React, { useState } from 'react';

import { post } from './api';

const Login = ({ onLogin }) => {
  const [email, setEmail] = useState('');
  const [error, setError] = useState('');
  const [pending, setPending] = useState(false);

  const submit = async (event) => {
    event.preventDefault();
    setPending(true);
    setError('');

    try {
      const user = await post('/auth/login', { email });
      onLogin(user);
    } catch (err) {
      setError(err.message);
    } finally {
      setPending(false);
    }
  };

  return (
    <div className="Login">
      <h1>Welcome to Movies</h1>
      <p>Sign in with your email to rent movies. No password needed, this is a demo.</p>
      <form className="Login__form" onSubmit={submit}>
        <input
          className="Login__input"
          type="email"
          name="email"
          placeholder="you@example.com"
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          required
        />
        <button className="button Login__button" type="submit" disabled={pending}>
          {pending ? 'Signing in...' : 'Sign in'}
        </button>
      </form>
      {!!error && <div className="Login__error">{error}</div>}
    </div>
  );
};

export default Login;
