import React, { Component } from 'react';

import userAvatarImage from './assets/images/cindy.jpg';

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
        username: 'cindy',
        avatar: userAvatarImage
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
        <header className={`App__header ${this.state.fixHeader ? 'fixed' : ''}`}>
          <div className="App__logo">
            <MoviesIcon />
            Movies
          </div>
          <UserProfile user={session} />
        </header>
        <div className="App__content">
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


class UserProfile extends Component {
  render() {
    const { user } = this.props;
    return (
      <div className="UserProfile">
        <div className="User">
          <div className="name">{`${user.name} ${user.lastName}`}</div>
          <div className="image"><img src={user.avatar} alt="profile" /></div>
        </div>
      </div>
    );
  }
}

class TitleList extends Component {
  renderList() {
    const { titles = [], loaded, onRent } = this.props;
    const movies = titles.filter(item => !item?.rented);

    if (loaded) {
      if (movies.length === 0) {
        return (
          <div className="TitleListEmpty">
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
          <div style={{ flex: '1 auto' }} />
          { onRent ?
            <>
              {!!item?.price &&
                <div className='Item__price'>${item.price}</div>
              }
              <div style={{ flex: '1 auto' }} />
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

const MoviesIcon = () => {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" width="32" fill="none" viewBox="0 0 46 53">
      <path fill="#fff" fillRule="evenodd" d="m34.8 1.47 3.08 7.3L45.3 7.6 44.1 0l-9.3 1.47Zm-3.98.63 3.08 7.3-7.44 1.18-3.06-7.3 7.42-1.18Zm-9.17 9.24-3.06-7.3-7.43 1.18 3.06 7.3 7.43-1.18ZM6.35 5.98c-4.1.65-6.9 4.5-6.26 8.61l.03.16 9.29-1.47-3.06-7.3ZM.1 48.4a3.85 3.85 0 0 0 3.83 3.85H42c2.1 0 3.83-1.73 3.83-3.85V17.64H.09V48.4Zm18.22-21.66 12.27 8.2L18.3 43.1V26.75Z" clipRule="evenodd"/>
    </svg>
  );
};

export default App;
