import React, { useState } from 'react';

import { post } from './api';

const Banned = ({ user, onLogout }) => {
  const [deed, setDeed] = useState('');
  const [error, setError] = useState('');
  const [submitted, setSubmitted] = useState(false);

  const submit = async (event) => {
    event.preventDefault();
    setError('');

    try {
      await post('/redemptions', { good_deed: deed });
      setSubmitted(true);
    } catch (err) {
      setError(err.message);
    }
  };

  return (
    <div className="Banned">
      <h1>You have been banned</h1>
      <img className="Banned__meme" src="/banned-meme.png" alt="A video store clerk refusing to hand over any more movies" />
      <p className="Banned__reason">
        Reason: <strong>{user.ban_reason || 'misbehaving'}</strong>
      </p>
      {submitted ? (
        <p className="Banned__thanks">
          Thanks! An admin is reviewing your good deed. Refresh this page once they approve it.
        </p>
      ) : (
        <>
          <p>
            To get back in, go perform a good deed and tell us about it. An admin will review it and lift your ban.
          </p>
          <form className="Banned__form" onSubmit={submit}>
            <textarea
              className="Banned__input"
              placeholder="I returned all my movies and helped a neighbor debug their Kubernetes cluster"
              value={deed}
              onChange={(event) => setDeed(event.target.value)}
              required
            />
            <button className="button Banned__button" type="submit">Request forgiveness</button>
          </form>
        </>
      )}
      {!!error && <div className="Banned__error">{error}</div>}
      <button className="button" type="button" onClick={onLogout}>Sign out</button>
    </div>
  );
};

export default Banned;
