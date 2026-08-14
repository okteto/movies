import React, { useCallback, useEffect, useRef, useState } from 'react';
import { BrowserRouter as Router, Switch, Route, Link } from 'react-router-dom';

import Admin from './Admin';
import Banned from './Banned';
import Loader from './Loader';
import Login from './Login';
import { CartIcon, Logo, MoviesIcon, Symbol } from './Icons';
import { financial, formatDate, get, post } from './api';

import './App.css';

const App = () => {
  const [session, setSession] = useState({ user: null, loaded: false });
  const [fixHeader, setFixHeader] = useState(false);
  const appRef = useRef(null);

  const loadSession = useCallback(async () => {
    try {
      const user = await get('/me');
      setSession({ user, loaded: true });
    } catch (err) {
      setSession({ user: null, loaded: true });
    }
  }, []);

  useEffect(() => {
    loadSession();
  }, [loadSession]);

  const logout = async () => {
    await post('/auth/logout');
    setSession({ user: null, loaded: true });
  };

  return (
    <Router>
      <div className="App" ref={appRef} onScroll={() => setFixHeader(appRef.current.scrollTop > 20)}>
        <div className={`App__header ${fixHeader ? 'fixed' : ''}`}>
          <Link to="/">
            <div className="App__logo">
              <MoviesIcon size="22" />
              Movies
            </div>
          </Link>
          <Logo size="24" />
        </div>

        {MODE === 'development' && <DevToast />}

        <div className="App__content">
          <Switch>
            <Route path="/admin">
              <Admin />
            </Route>
            <Route exact path="/">
              <Store
                session={session}
                onLogin={(user) => setSession({ user, loaded: true })}
                onLogout={logout}
                onRefreshSession={loadSession}
              />
            </Route>
          </Switch>
        </div>
      </div>
    </Router>
  );
};

const Store = ({ session, onLogin, onLogout, onRefreshSession }) => {
  const [catalog, setCatalog] = useState({ data: [], loaded: false });
  const [history, setHistory] = useState([]);
  const [showHistory, setShowHistory] = useState(false);
  const [error, setError] = useState('');

  const { user, loaded } = session;
  const email = user?.email;
  const banned = user?.banned;

  const refreshData = useCallback(async () => {
    if (!email || banned) {
      return;
    }

    const [movies, rentalHistory] = await Promise.all([get('/availability'), get('/rentals/history')]);
    setCatalog({ data: movies, loaded: true });
    setHistory(rentalHistory);
  }, [email, banned]);

  // an admin may have banned the user while the app was on another screen
  useEffect(() => {
    onRefreshSession();
  }, []);

  useEffect(() => {
    refreshData();
  }, [refreshData]);

  if (!loaded) {
    return <Loader />;
  }

  if (!user) {
    return <Login onLogin={onLogin} />;
  }

  if (user.banned) {
    return <Banned user={user} onLogout={onLogout} />;
  }

  const act = async (path, item) => {
    setError('');
    try {
      await post(path, { catalog_id: String(item.id) });
      // the rent service publishes to kafka, give the worker a moment to catch up
      await new Promise((resolve) => setTimeout(resolve, 700));
      await Promise.all([onRefreshSession(), refreshData()]);
    } catch (err) {
      setError(err.message);
    }
  };

  const rented = catalog.data.filter((movie) => movie.rented);
  const cost = financial(rented.reduce((total, movie) => total + Number(movie.price ?? 0), 0));

  return (
    <div>
      <div className="App__nav">
        <div className="App__session">
          Signed in as <strong>{user.display_name || user.email}</strong>
        </div>
        <button className="button" type="button" onClick={onLogout}>Sign out</button>
        <Link className="button" role="button" to="/admin">Admin</Link>
      </div>

      {!!error && <div className="App__error">{error}</div>}

      <TitleList
        title={`${user.display_name || user.email}'s movies`}
        cost={cost}
        titles={rented}
        loaded={catalog.loaded}
        onReturn={(item) => act('/rent/return', item)}
      />

      <TitleList
        title="Store"
        titles={catalog.data.filter((movie) => !movie.rented)}
        loaded={catalog.loaded}
        onRent={(item) => act('/rent', item)}
      />

      <div className="History">
        <h1 className="History__title" onClick={() => setShowHistory(!showHistory)}>
          Rental history ({history.length})
          <span className="History__toggle">{showHistory ? 'hide' : 'show'}</span>
        </h1>
        {showHistory && (
          history.length === 0 ? (
            <div className="TitleList--empty">You haven't rented anything yet.</div>
          ) : (
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
                {history.map((rental) => (
                  <tr className="Table__row" key={rental.id}>
                    <td className="Table__data">{rental.title}</td>
                    <td className="Table__data">${financial(rental.price)}</td>
                    <td className="Table__data">{formatDate(rental.rented_at)}</td>
                    <td className="Table__data">{rental.returned_at ? formatDate(rental.returned_at) : 'still rented'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )
        )}
      </div>
    </div>
  );
};

const DevToast = () => {
  return (
    <div className='DevToast'>
      <Symbol />
      In Development Mode
    </div>
  );
};

const TitleList = ({ titles = [], title, cost = 0, loaded, onRent, onReturn }) => {
  const renderList = () => {
    if (!loaded) {
      return null;
    }

    if (titles.length === 0) {
      return (
        <div className="TitleList--empty">
          {onRent ? 'No movies left to rent.' : 'You have no movies rented.'}
        </div>
      );
    }

    return titles.map((item) => (
      <Item
        key={item.id}
        item={item}
        backdrop={`/${item.backdrop_path}`}
        onRent={onRent}
        onReturn={onReturn}
      />
    ));
  };

  return (
    <div className="TitleList">
      <div className="Title">
        <h1>{title}</h1>
        <div className="TitleList__slider">
          {!!cost && Number(cost) > 0 && <Cart cost={cost} titles={titles} />}
          {renderList() || <Loader />}
        </div>
      </div>
    </div>
  );
};

const Item = ({ item, onRent, onReturn, backdrop }) => {
  const soldOut = onRent && item.available === 0;

  return (
    <div className="Item">
      <div className="Item__container" style={{ backgroundImage: `url(./${backdrop})` }}>
        {onRent && (
          <div className={`Item__copies ${soldOut ? 'Item__copies--out' : ''}`}>
            {soldOut ? 'all copies rented' : `${item.available} of ${item.copies} available`}
          </div>
        )}
        <div className="Item__overlay">
          <div className="Item__title">{item?.original_title ?? 'Unknown Title'}</div>
          <div className="Item__rating">{item?.vote_average ?? 0} / 10</div>
          <div className="spring" />
          {onRent ? (
            <>
              {!!item?.price && <div className='Item__price'>${item.price}</div>}
              <div className="spring" />
              {soldOut ? (
                <div className="Item__button Item__button--rented button">Sold out</div>
              ) : (
                <div className="Item__button button" onClick={() => onRent(item)}>Rent</div>
              )}
            </>
          ) : (
            <>
              <div className="Item__button Item__button--rented button">Watch Now</div>
              <div className="Item__button button" onClick={() => onReturn(item)}>Return</div>
            </>
          )}
        </div>
      </div>
    </div>
  );
};

const Cart = ({ cost, titles }) => {
  return (
    <div className="Cart">
      <div className="Cart__container">
        <div className="Cart__header">
          <CartIcon />
          Cart
        </div>
        <div className="Cart__list">
          {titles.map((movie) => (
            <div className="Cart__item" key={movie.id}>
              <div className="Cart__item-name">{movie.original_title}</div>
              <div className="Cart__item-price">${movie.price}</div>
            </div>
          ))}
        </div>
        <div className="Cart__total">
          <div className="Cart__total-title">Total due:</div>
          <div className="Cart__total-price">${cost}</div>
        </div>
      </div>
    </div>
  );
};

export default App;
