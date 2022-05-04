import React, { Component } from 'react';

import './App.css';

const compact = (movies = []) => {
  return movies.filter((item, index, self) =>
    self.findIndex(i => i.id === item.id) === index
  );
}

const financial = (x) => {
  return Number.parseFloat(x).toFixed(2);
}

class App extends Component {
  constructor(props) {
    super(props);

    this.state = {
      catalog: {
        data: [],
        loaded: false
      },
      rental: {
        data: [],
        loaded: false
      },
      cost: 0,
      session: {
        name: 'Cindy',
        lastName: 'Lopez',
        username: 'cindy'
      },
      fixHeader: false
    };

    this.appRef = React.createRef();
  }

  componentDidMount() {
    this.refreshData();
  }

  handleRent = async (item) => {
    await fetch('/rent', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        catalog_id: item.id,
        price: item.price
      })
    });
    this.refreshData();
  }

  refreshData = async () => {
    const catalogPromise = fetch('/catalog')
      .then(res => res.json())
      .then(result => compact(result));

    const rentalsPromise = fetch('/rentals')
      .then(res => res.json())
      .then(result => compact(result));

    const [catalog, rentals] = await Promise.all([catalogPromise, rentalsPromise]);
    this.setState({
      rental: {
        data: rentals,
        loaded: true
      },
      catalog: {
        data: catalog.map(movie => ({
          ...movie,
          rented: !!rentals.find(c => c.id === movie.id)
        })),
        loaded: true
      },
      cost: financial(rentals.reduce((acc, item) => acc += Number(item?.price ?? 0), 0))
    });
  }

  handleScroll = () => {
    this.setState({
      fixHeader: this.appRef.current.scrollTop > 20
    });
  }

  render() {
    const { catalog, rental, session, cost } = this.state;
    return (
      <div className="App" ref={this.appRef} onScroll={this.handleScroll}>
        <div className={`App__header ${this.state.fixHeader ? 'fixed' : ''}`}>
          <div className="App__logo">
            <MoviesIcon size="22" />
            Movies
          </div>
          <Logo size="24" />
        </div>

        {MODE === 'development' &&
          <DevToast />
        }

        <div className="App__content">
          {/* <div className="App__promo">
            <KubeconLogo size="22" />
            Kubecon 2022 special offer! Get a <strong>50% discount</strong> on all movies today!
          </div> */}
          <TitleList
            title={`${session.name}'s movies`}
            cost={cost}
            titles={rental.data}
            loaded={rental.loaded}
          />
          <TitleList
            title="Store"
            titles={catalog.data}
            loaded={catalog.loaded}
            onRent={this.handleRent}
          />
        </div>
      </div>
    );
  }
}


class Loader extends Component {
  render() {
    return (
      <div className="Loader">
        <svg version="1.1" id="loader" x="0px" y="0px"
          width="40px"
          height="40px"
          viewBox="0 0 50 50"
          style={{
            enableBackground: 'new 0 0 50 50'
          }}>
          <path fill="#000" d="M43.935,25.145c0-10.318-8.364-18.683-18.683-18.683c-10.318,0-18.683,8.365-18.683,18.683h4.068c0-8.071,6.543-14.615,14.615-14.615c8.072,0,14.615,6.543,14.615,14.615H43.935z">
            <animateTransform attributeType="xml"
              attributeName="transform"
              type="rotate"
              from="0 25 25"
              to="360 25 25"
              dur="0.6s"
              repeatCount="indefinite"/>
          </path>
        </svg>
      </div>
    );
  }
}

const DevToast = () => {
  return (
    <div className='DevToast'>
      <Symbol />
      In Development Mode
    </div>
  );
};

class TitleList extends Component {
  renderList() {
    const { titles = [], loaded, onRent } = this.props;
    const movies = titles.filter(item => !item?.rented);

    if (loaded) {
      if (movies.length === 0) {
        return (
          <div className="TitleList--empty">
            {onRent && 'No movies left to rent.'}
          </div>
        );
      }

      return movies.map((item, i) => {
        const backDrop = `/${item.backdrop_path}`;
        return (
          <Item
            key={item.id}
            item={item}
            backdrop={backDrop}
            onRent={onRent}
          />
        );
      });
    }
  }

  render() {
    const { titles, title, cost = 0 } = this.props;

    return (
      <div className="TitleList">
        <div className="Title">
          <h1>
            {title}
          </h1>
          <div className="TitleList__slider">
            {!!cost &&
              <Cart cost={cost} titles={titles} />
            }
            {this.renderList() || <Loader />}
          </div>
        </div>
      </div>
    );
  }
}

const Item = ({ item, onRent, backdrop }) => {
  return (
    <div className="Item">
      <div className="Item__container" style={{ backgroundImage: `url(./${backdrop})` }}>
        <div className="Item__overlay">
          <div className="Item__title">{item?.original_title ?? 'Unknown Title'}</div>
          <div className="Item__rating">{item?.vote_average ?? 0} / 10</div>
          <div className="spring" />
          { onRent ?
            <>
              {!!item?.price &&
                <div className='Item__price'>${item.price}</div>
              }
              <div className="spring" />
              <div className="Item__button" onClick={() => onRent(item)}>
                Rent
              </div>
            </> :
            <>
              <div className="Item__button Item__button--rented">
                Watch Now
              </div>
            </>
          }
        </div>
      </div>
    </div>
  );
}

const CartIcon = () => {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 21 22" width="16">
      <path fill="#fff" d="M6.3 22c-.53 0-.99-.2-1.36-.58a1.94 1.94 0 0 1-.56-1.4c0-.55.18-1.02.56-1.4a1.86 1.86 0 0 1 2.71-.01c.38.39.57.86.57 1.4 0 .56-.19 1.03-.56 1.41A1.8 1.8 0 0 1 6.3 22Zm10.68 0c-.54 0-1-.2-1.36-.58a1.94 1.94 0 0 1-.56-1.4c0-.55.18-1.02.56-1.4a1.85 1.85 0 0 1 2.7-.01c.39.39.58.86.58 1.4 0 .56-.19 1.03-.56 1.41-.38.39-.83.58-1.36.58ZM4.9 3.83l2.94 6.28h7.69l3.33-6.28H4.91Zm-.8-1.65h15.72c.57 0 .93.17 1.08.53.16.36.1.76-.14 1.2l-3.6 6.7c-.18.3-.43.57-.75.8-.32.23-.67.35-1.04.35h-8.1L5.8 14.62h13.1v1.65H6.04c-.74 0-1.28-.25-1.61-.77-.33-.51-.33-1.09.01-1.73l1.7-3.25-4.05-8.87H0V0h3.12l1 2.18Zm3.74 7.93h7.69-7.7Z" opacity=".8"/>
    </svg>
  );
}


const Cart = ({ cost, titles }) => {
  return (
    <div className="Cart">
      <div className="Cart__container">
        <div className="Cart__header">
          <CartIcon />
          Cart
        </div>
        <div className="Cart__list">
          {titles.map((movie, i) => (
            <div className="Cart__item" key={i}>
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
}

const MoviesIcon = ({ size = '28'}) => {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" width={size} fill="none" viewBox="0 0 46 53">
      <path fill="#fff" fillRule="evenodd" d="m34.8 1.47 3.08 7.3L45.3 7.6 44.1 0l-9.3 1.47Zm-3.98.63 3.08 7.3-7.44 1.18-3.06-7.3 7.42-1.18Zm-9.17 9.24-3.06-7.3-7.43 1.18 3.06 7.3 7.43-1.18ZM6.35 5.98c-4.1.65-6.9 4.5-6.26 8.61l.03.16 9.29-1.47-3.06-7.3ZM.1 48.4a3.85 3.85 0 0 0 3.83 3.85H42c2.1 0 3.83-1.73 3.83-3.85V17.64H.09V48.4Zm18.22-21.66 12.27 8.2L18.3 43.1V26.75Z" clipRule="evenodd"/>
    </svg>
  );
};

const Logo = ({ size = '24'}) => {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" height={size} fill="none" viewBox="0 0 420 121">
      <path d="M0 0h420v121H0z"/>
      <path fill="#00D1CA" d="M60.5 121a60.5 60.5 0 1 0 0-121 60.5 60.5 0 0 0 0 121Z"/>
      <path fill="#1A263E" d="M39.46 60.5c0-11.52 9.52-21.04 21.49-21.04a21.6 21.6 0 0 1 16.84 7.97 6.76 6.76 0 1 0 10.51-8.52 35.14 35.14 0 0 0-27.35-12.98c-19.24 0-35.02 15.38-35.02 34.57 0 19.2 15.78 34.57 35.02 34.57A35.14 35.14 0 0 0 88.3 82.1a6.76 6.76 0 0 0-10.5-8.52 21.6 21.6 0 0 1-16.85 7.97c-11.97 0-21.5-9.52-21.5-21.04Z"/>
      <path fill="#1A263E" d="M88.3 67.26a6.76 6.76 0 1 0 0-13.52 6.76 6.76 0 0 0 0 13.52Z"/>
      <path fill="#fff" d="m227.98 89-17.25-25.22V89h-9.6V10h9.6v47.9L227.98 33h12.52L221 60l19.5 29h-12.52Zm38.91 0c-4.19 0-7.66-1.33-10.42-4-2.76-2.65-4.17-6.35-4.24-11.1V42.06h-9.01V33h9V19.72h9.64V33h12.57v9.06h-12.57v28.6c0 3.53.67 5.92 2 7.18a6.38 6.38 0 0 0 4.5 1.88 10 10 0 0 0 4.29-.97l3.04 8.85a25.2 25.2 0 0 1-8.8 1.4Zm36.61-56c8 0 13.67 2.68 19.1 7.65 5.43 4.96 8.97 13.11 8.33 22.6h-43.69c.15 5.21 1.9 8.18 5.28 11.36a17.04 17.04 0 0 0 12.11 4.77c6.46 0 11.86-3.35 14.57-9.55h10.65c-1.44 6.34-6.7 12.7-11.02 15.37-4.32 2.67-9.16 3.8-14.59 3.8-7.92 0-14.45-2.68-19.6-8.03-5.13-5.35-7.75-12.62-7.75-20.79 0-8.1 2.47-13.71 7.46-19.1 4.99-5.39 11.37-8.08 19.15-8.08Zm88.84.1c15.28 0 27.66 12.5 27.66 27.95C420 76.48 407.62 89 392.34 89c-15.27 0-27.66-12.52-27.66-27.95 0-15.44 12.38-27.96 27.66-27.96Zm-223.68 0c15.28 0 27.66 12.5 27.66 27.95 0 15.43-12.38 27.95-27.66 27.95S141 76.48 141 61.05c0-15.44 12.38-27.96 27.66-27.96Zm182.2-13.38V33h12.56v9.06h-12.56v28.6c0 3.53.66 5.92 1.99 7.18a6.38 6.38 0 0 0 4.5 1.88 10 10 0 0 0 4.29-.97l3.04 8.85a25.2 25.2 0 0 1-8.8 1.4c-4.19 0-7.66-1.33-10.42-4-2.76-2.65-4.17-6.35-4.24-11.1V42.06h-9V33h9V19.72h9.63Zm41.48 23.3a17.93 17.93 0 0 0-17.84 18.03c0 9.95 7.99 18.02 17.84 18.02s17.84-8.07 17.84-18.02c0-9.96-7.99-18.03-17.84-18.03Zm-223.68 0a17.93 17.93 0 0 0-17.84 18.03c0 9.95 7.99 18.02 17.84 18.02S186.5 71 186.5 61.05c0-9.96-7.99-18.03-17.84-18.03Zm135.55-1.06c-3.85 0-7.66 1.44-10.44 3.77-2.78 2.34-5.21 5.48-6.1 9.42h33c-.79-3.6-2.76-6.9-5.9-9.42-3.14-2.5-5.9-3.77-10.56-3.77Z"/>
    </svg>
  );
};

const Symbol = ({ size = '18'}) => {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" width={size} fill="none" viewBox="0 0 21 21">
      <path fill="#1A263E" d="M10.5 21a10.5 10.5 0 1 0 0-21 10.5 10.5 0 0 0 0 21Z"/>
      <path fill="#00D1CA" d="M6.85 10.5c0-2 1.65-3.65 3.73-3.65 1.19 0 2.24.54 2.92 1.38a1.17 1.17 0 0 0 1.82-1.48A6.1 6.1 0 0 0 4.5 10.5a6.1 6.1 0 0 0 10.82 3.75 1.17 1.17 0 1 0-1.82-1.48 3.75 3.75 0 0 1-2.92 1.38 3.7 3.7 0 0 1-3.73-3.65Z"/>
      <path fill="#00D1CA" d="M15.33 11.67a1.17 1.17 0 1 0 0-2.34 1.17 1.17 0 0 0 0 2.34Z"/>
    </svg>
  );
};

const KubeconLogo = ({ size = '21'}) => {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" height={size} fill="none" viewBox="0 0 59 62">
      <path fill="#fff" fill-rule="evenodd" d="m21.18 32.8-1.72.37c0-.23-.05-.47-.05-.71 0-1.82.5-3.6 1.43-5.17l3.35 2.97c.06.06.12.1.2.14-.21.5-.34 1.04-.39 1.58-.1.05-.24.05-.34.1l-2.48.71Zm6.5-10.14v4.68c-.47.15-.9.37-1.29.67-.14-.1-.29-.24-.43-.33l-3.4-2.4a9.43 9.43 0 0 1 5.12-2.62ZM4.4 33.46 9.04 13.2c.71-3.2 1.2-4.06 4.1-5.73 4.02-2.35 7.04-4.02 11.34-6.5 2.67-1.54 3.2-1.44 3.2 1.86v16.54c-3 .37-5.77 1.77-7.84 3.97l-1.43-1a.69.69 0 0 1-.34-.48c-.13-.46-.4-.86-.76-1.15a1.56 1.56 0 0 0-1.96 2.44c.21.19.48.32.76.38.2.05.38.05.58.1.15.03.3.1.43.19L18.45 25a13.09 13.09 0 0 0-2.15 8.8l-2.91.62c-.21 0-.42 0-.62.05l-.43.14c-1.92.38-3.83.81-5.79 1.24-2.63.62-2.72.1-2.15-2.39Zm26.67-6.12v-4.68c1.92.34 3.7 1.24 5.12 2.58l-3.68 2.63c-.05.05-.1.1-.14.1-.4-.27-.84-.48-1.3-.63Zm8.23 5.5-4.6-1.05a4.5 4.5 0 0 0-.37-1.44c.04-.04.1-.1.14-.1l3.4-3a9.99 9.99 0 0 1 1.47 5.2.93.93 0 0 0-.04.39ZM33.32 7.55l6.74 3.25c3.44 1.68 6.89 3.35 10.38 4.97a899.6 899.6 0 0 1 2.53 11.19c.52 2.25.76 3.44 1.29 5.64.81 3.5.72 3.78-2.87 2.96l-8.84-2A12.81 12.81 0 0 0 40.3 25l1.24-1.15c.15-.12.34-.2.53-.24.47-.03.91-.2 1.29-.47a1.56 1.56 0 1 0-1.96-2.45c-.24.18-.44.41-.57.68-.1.19-.15.33-.24.52a.74.74 0 0 1-.29.39l-1.43 1a13.05 13.05 0 0 0-7.84-3.97V9.22c.04-2.34.24-2.62 2.3-1.67Zm-3.06 30.12a5.29 5.29 0 0 1-1.77.05l-.14.28-2.1 3.88a9.93 9.93 0 0 0 6.35 0l-2.1-3.83c-.1-.12-.18-.24-.24-.38ZM34 35.13l4.59 1.06a9.92 9.92 0 0 1-3.73 4.54l-1.82-4.4c.36-.36.68-.77.96-1.2Zm-13.81 1 4.58-1.04c.29.45.62.87 1 1.24l-.14.29-1.67 4.1a9.85 9.85 0 0 1-3.77-4.58Zm37.57 8.09c-2.2 2.72-3.2 4.01-5.4 6.74-2.49 3.1-4.5 5.64-6.98 8.75-1.77 2.24-3.06 2.3-5.98 2.3H19.31c-3.82 0-4.1 0-6.45-2.97a959.2 959.2 0 0 1-6.45-8.13c-1.82-2.3-2.3-2.87-4.11-5.16-2.82-3.54-3.3-4.83 1.1-5.79l13.57-3.06a13.42 13.42 0 0 0 5.7 6.89l-.63 1.43a.84.84 0 0 1-.38.43c-.42.22-.76.58-.96 1a1.57 1.57 0 0 0 2.81 1.41l.01-.01c.13-.28.21-.57.24-.87 0-.2 0-.38-.05-.57a.86.86 0 0 1 .1-.48l.76-1.39a13.37 13.37 0 0 0 9.52-.04l.76 1.38c.1.17.14.38.1.58-.09.45-.02.92.19 1.33a1.57 1.57 0 0 0 2.83-1.36v-.02h-.01a2 2 0 0 0-.53-.72 2.9 2.9 0 0 0-.48-.33.6.6 0 0 1-.28-.39l-.62-1.53a13.1 13.1 0 0 0 5.64-6.88l14.96 3.4c3.01.8 2.82 1.86 1.1 4.06" clip-rule="evenodd"/>
      <path fill="#fff" fill-rule="evenodd" d="M31.07 27.34v-4.68c1.92.34 3.7 1.24 5.12 2.58l-3.68 2.63c-.05.05-.1.1-.14.1-.4-.27-.84-.48-1.3-.63Zm8.23 5.5-4.6-1.05a4.5 4.5 0 0 0-.37-1.44c.04-.04.1-.1.14-.1l3.4-3a9.99 9.99 0 0 1 1.47 5.2.93.93 0 0 0-.04.39ZM33.32 7.55l6.74 3.25c3.44 1.68 6.89 3.35 10.38 4.97a940.6 940.6 0 0 1 2.53 11.19c.52 2.25.76 3.44 1.29 5.64.81 3.5.72 3.78-2.87 2.96l-8.84-2A12.81 12.81 0 0 0 40.3 25l1.24-1.15c.15-.12.34-.2.53-.24.47-.03.91-.2 1.29-.47a1.56 1.56 0 1 0-1.96-2.45c-.24.18-.44.41-.57.68-.1.19-.15.33-.24.52a.74.74 0 0 1-.29.39l-1.43 1a13.05 13.05 0 0 0-7.84-3.97V9.22c.04-2.34.24-2.62 2.3-1.67ZM34 35.13l4.59 1.06a9.92 9.92 0 0 1-3.73 4.54l-1.82-4.4c.36-.36.68-.77.96-1.2Zm-4.64 7.22v-4.59c.3 0 .61-.04.91-.1.05.13.11.24.2.34l2.1 3.83a10.3 10.3 0 0 1-3.2.52Zm28.4 1.87c-2.2 2.72-3.2 4.01-5.4 6.74-2.49 3.1-4.5 5.64-6.98 8.75-1.77 2.24-3.06 2.3-5.98 2.3h-10V45.6c1.64 0 3.26-.3 4.78-.91l.77 1.38c.1.17.14.38.1.58-.09.45-.02.92.18 1.33a1.57 1.57 0 0 0 2.82-1.38 2 2 0 0 0-.52-.72 2.9 2.9 0 0 0-.48-.33.6.6 0 0 1-.29-.39l-.62-1.53c2.63-1.56 4.62-4 5.64-6.88l14.97 3.4c2.91.8 2.72 1.86 1 4.06" clip-rule="evenodd"/>
    </svg>
  );
}

export default App;
