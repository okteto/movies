import React, { useCallback, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';

import Loader from './Loader';
import { del, financial, formatDate, get, post, put } from './api';

import './Admin.css';

const emptyMovie = {
  original_title: '',
  overview: '',
  backdrop_path: '/poster-kube.png',
  price: '2.99',
  vote_average: '5',
  copies: '3'
};

const Admin = () => {
  const [admin, setAdmin] = useState({ isAdmin: false, loaded: false });
  const [tab, setTab] = useState('users');

  const loadSession = useCallback(async () => {
    const session = await get('/adminapi/session');
    setAdmin({ isAdmin: session.admin, loaded: true });
  }, []);

  useEffect(() => {
    loadSession();
  }, [loadSession]);

  const logout = async () => {
    await post('/auth/admin-logout');
    setAdmin({ isAdmin: false, loaded: true });
  };

  return (
    <div className="Admin">
      <div className="App__nav">
        <Link className="button" role="button" to="/">Back to Movies</Link>
        {admin.isAdmin && <button className="button" type="button" onClick={logout}>Sign out</button>}
      </div>

      {!admin.loaded && <Loader />}

      {admin.loaded && !admin.isAdmin && <AdminLogin onLogin={() => setAdmin({ isAdmin: true, loaded: true })} />}

      {admin.loaded && admin.isAdmin && (
        <>
          <div className="Admin__tabs">
            <button
              type="button"
              className={`button ${tab === 'users' ? 'Admin__tab--active' : ''}`}
              onClick={() => setTab('users')}
            >
              Users
            </button>
            <button
              type="button"
              className={`button ${tab === 'catalog' ? 'Admin__tab--active' : ''}`}
              onClick={() => setTab('catalog')}
            >
              Catalog
            </button>
          </div>
          {tab === 'users' ? <Users /> : <Catalog />}
        </>
      )}
    </div>
  );
};

const AdminLogin = ({ onLogin }) => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');

  const submit = async (event) => {
    event.preventDefault();
    setError('');

    try {
      await post('/auth/admin-login', { username, password });
      onLogin();
    } catch (err) {
      setError(err.message);
    }
  };

  return (
    <div className="Login">
      <h1>Admin panel</h1>
      <form className="Login__form" onSubmit={submit}>
        <input
          className="Login__input"
          placeholder="username"
          value={username}
          onChange={(event) => setUsername(event.target.value)}
          required
        />
        <input
          className="Login__input"
          type="password"
          placeholder="password"
          value={password}
          onChange={(event) => setPassword(event.target.value)}
          required
        />
        <button className="button Login__button" type="submit">Sign in</button>
      </form>
      {!!error && <div className="Login__error">{error}</div>}
    </div>
  );
};

const Users = () => {
  const [users, setUsers] = useState([]);
  const [redemptions, setRedemptions] = useState([]);
  const [expanded, setExpanded] = useState(null);
  const [rentals, setRentals] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const [userList, redemptionList] = await Promise.all([get('/adminapi/users'), get('/adminapi/redemptions')]);
      setUsers(userList);
      setRedemptions(redemptionList);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const toggle = async (email) => {
    if (expanded === email) {
      setExpanded(null);
      return;
    }

    setExpanded(email);
    setRentals(await get(`/adminapi/users/${encodeURIComponent(email)}/rentals`));
  };

  const ban = async (email) => {
    const reason = window.prompt('Why are you banning this user?', 'renting too many movies');
    if (reason === null) {
      return;
    }

    await post(`/adminapi/users/${encodeURIComponent(email)}/ban`, { reason });
    await refresh();
  };

  const unban = async (email) => {
    await post(`/adminapi/users/${encodeURIComponent(email)}/unban`);
    await refresh();
  };

  const resolve = async (id, status) => {
    await post(`/adminapi/redemptions/${id}/resolve`, { status });
    await refresh();
  };

  const pending = redemptions.filter((redemption) => redemption.status === 'pending');

  if (loading) {
    return <Loader />;
  }

  return (
    <div className="Users">
      {!!error && <div className="App__error">{error}</div>}

      {pending.length > 0 && (
        <>
          <h1>Good deeds waiting for review</h1>
          <table className="Table__table">
            <thead className="Table__head">
              <tr className="Table__row">
                <th className="Table__header">user</th>
                <th className="Table__header">good deed</th>
                <th className="Table__header">requested</th>
                <th className="Table__header" />
              </tr>
            </thead>
            <tbody className="Table__body">
              {pending.map((redemption) => (
                <tr className="Table__row" key={redemption.id}>
                  <td className="Table__data">{redemption.user_email}</td>
                  <td className="Table__data">{redemption.good_deed}</td>
                  <td className="Table__data">{formatDate(redemption.created_at)}</td>
                  <td className="Table__data Admin__actions">
                    <button className="button" type="button" onClick={() => resolve(redemption.id, 'approved')}>
                      Forgive
                    </button>
                    <button className="button" type="button" onClick={() => resolve(redemption.id, 'rejected')}>
                      Reject
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </>
      )}

      <h1>Users</h1>
      <table className="Table__table">
        <thead className="Table__head">
          <tr className="Table__row">
            <th className="Table__header">email</th>
            <th className="Table__header">name</th>
            <th className="Table__header">rented now</th>
            <th className="Table__header">total rentals</th>
            <th className="Table__header">last rental</th>
            <th className="Table__header">status</th>
            <th className="Table__header" />
          </tr>
        </thead>
        <tbody className="Table__body">
          {users.map((user) => (
            <React.Fragment key={user.email}>
              <tr className="Table__row">
                <td className="Table__data">{user.email}</td>
                <td className="Table__data">{user.display_name}</td>
                <td className="Table__data">{user.active_rentals}</td>
                <td className="Table__data">{user.total_rentals}</td>
                <td className="Table__data">{formatDate(user.last_rental_at) || 'never'}</td>
                <td className="Table__data">
                  {user.banned ? <span className="Admin__banned">banned: {user.ban_reason}</span> : 'active'}
                </td>
                <td className="Table__data Admin__actions">
                  <button className="button" type="button" onClick={() => toggle(user.email)}>
                    {expanded === user.email ? 'Hide history' : 'History'}
                  </button>
                  {user.banned ? (
                    <button className="button" type="button" onClick={() => unban(user.email)}>Unban</button>
                  ) : (
                    <button className="button" type="button" onClick={() => ban(user.email)}>Ban</button>
                  )}
                </td>
              </tr>
              {expanded === user.email && (
                <tr className="Table__row">
                  <td className="Table__data" colSpan="7">
                    <History rentals={rentals} />
                  </td>
                </tr>
              )}
            </React.Fragment>
          ))}
        </tbody>
      </table>
    </div>
  );
};

const History = ({ rentals }) => {
  if (rentals.length === 0) {
    return <div className="TitleList--empty">This user hasn't rented anything yet.</div>;
  }

  return (
    <table className="Table__table">
      <thead className="Table__head">
        <tr className="Table__row">
          <th className="Table__header">movie</th>
          <th className="Table__header">price</th>
          <th className="Table__header">rented</th>
          <th className="Table__header">returned</th>
        </tr>
      </thead>
      <tbody className="Table__body">
        {rentals.map((rental) => (
          <tr className="Table__row" key={rental.id}>
            <td className="Table__data">{rental.title}</td>
            <td className="Table__data">${financial(rental.price)}</td>
            <td className="Table__data">{formatDate(rental.rented_at)}</td>
            <td className="Table__data">{rental.returned_at ? formatDate(rental.returned_at) : 'still rented'}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
};

const Catalog = () => {
  const [movies, setMovies] = useState([]);
  const [editing, setEditing] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const refresh = useCallback(async () => {
    setLoading(true);
    setMovies(await get('/catalog'));
    setLoading(false);
  }, []);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const save = async (movie) => {
    setError('');
    try {
      if (movie.id) {
        await put(`/catalog/${movie.id}`, movie);
      } else {
        await post('/catalog', movie);
      }
      setEditing(null);
      await refresh();
    } catch (err) {
      setError(err.message);
    }
  };

  const remove = async (movie) => {
    if (!window.confirm(`Remove "${movie.original_title}" from the catalog?`)) {
      return;
    }

    setError('');
    try {
      await del(`/catalog/${movie.id}`);
      await refresh();
    } catch (err) {
      setError(err.message);
    }
  };

  if (loading) {
    return <Loader />;
  }

  return (
    <div className="Users">
      {!!error && <div className="App__error">{error}</div>}

      <h1>Catalog</h1>
      <div className="App__nav">
        <button className="button" type="button" onClick={() => setEditing({ ...emptyMovie })}>Add movie</button>
      </div>

      {editing && <MovieForm movie={editing} onCancel={() => setEditing(null)} onSave={save} />}

      <table className="Table__table">
        <thead className="Table__head">
          <tr className="Table__row">
            <th className="Table__header">id</th>
            <th className="Table__header">title</th>
            <th className="Table__header">price</th>
            <th className="Table__header">rating</th>
            <th className="Table__header">copies</th>
            <th className="Table__header" />
          </tr>
        </thead>
        <tbody className="Table__body">
          {movies.map((movie) => (
            <tr className="Table__row" key={movie.id}>
              <td className="Table__data">{movie.id}</td>
              <td className="Table__data">{movie.original_title}</td>
              <td className="Table__data">${financial(movie.price)}</td>
              <td className="Table__data">{movie.vote_average}</td>
              <td className="Table__data">{movie.copies}</td>
              <td className="Table__data Admin__actions">
                <button className="button" type="button" onClick={() => setEditing(movie)}>Edit</button>
                <button className="button" type="button" onClick={() => remove(movie)}>Remove</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

const MovieForm = ({ movie, onCancel, onSave }) => {
  const [form, setForm] = useState(movie);

  const update = (field) => (event) => setForm({ ...form, [field]: event.target.value });

  return (
    <form
      className="Admin__form"
      onSubmit={(event) => {
        event.preventDefault();
        onSave(form);
      }}
    >
      <label className="Admin__field">
        Title
        <input className="Login__input" value={form.original_title} onChange={update('original_title')} required />
      </label>
      <label className="Admin__field">
        Poster
        <input className="Login__input" value={form.backdrop_path} onChange={update('backdrop_path')} />
      </label>
      <label className="Admin__field">
        Price
        <input className="Login__input" type="number" step="0.01" min="0" value={form.price} onChange={update('price')} />
      </label>
      <label className="Admin__field">
        Rating
        <input className="Login__input" type="number" step="0.1" min="0" max="10" value={form.vote_average} onChange={update('vote_average')} />
      </label>
      <label className="Admin__field">
        Copies
        <input className="Login__input" type="number" min="0" value={form.copies} onChange={update('copies')} />
      </label>
      <label className="Admin__field Admin__field--wide">
        Overview
        <textarea className="Login__input" value={form.overview} onChange={update('overview')} />
      </label>
      <div className="Admin__actions">
        <button className="button" type="submit">Save</button>
        <button className="button" type="button" onClick={onCancel}>Cancel</button>
      </div>
    </form>
  );
};

export default Admin;
