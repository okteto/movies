import React, { useState, useEffect } from 'react';
import Loader from './Loader';
import './Users.css';

const Users = () => {
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(false);

  const fetchUsers = async () => {
    setLoading(true);
    const reqUsers = await fetch('/users');
    const usersResult = await reqUsers.json();
    setUsers(usersResult);
    return setLoading(false);
  };

  useEffect(() => {
    fetchUsers();
  }, []);

  return (
    <div className="Users">
        <h1>Users </h1>
      { loading && <Loader />}
      { !!users.length && <Table users={users} /> }
    </div>
  )
};

const Table = ({ users }) => {
  const allUsers = [...users];
  const [currentUsers, setCurrentUsers] = useState(allUsers.slice(0, 100));
  const [currentPage, setCurrentPage] = useState(1);

  const pageSize = 100;
  const totalPages = Math.ceil(allUsers.length / pageSize);

  useEffect(() => {
    handleUsersToDisplay();
  }, [currentPage]);

  const handleUsersToDisplay = () => {
    const endIndex = currentPage * pageSize;
    const startIndex = endIndex - pageSize;
    setCurrentUsers(allUsers.slice(startIndex, endIndex));
  }

  return (
    <div className="Table">
      <h5>Total Users: {allUsers.length ? allUsers.length : '0'}</h5>
      <table className='Table__table'>
        <thead className="Table__head">
          <tr className="Table__row">
            {
              Object.keys(allUsers[0]).map((key) => {
                return <th className="Table__header" key={key}>{key}</th>
              })
            }
          </tr>
        </thead>
        <tbody className="Table__body">
          {
            currentUsers.map((user) => (
              <tr className="Table__row" key={user.Userid}>
                {
                  Object.keys(user).map((property) => {
                    return (
                    <td className="Table__data" key={`${user.Userid}--${property}`}>
                      {user[property]}
                    </td>
                    )
                  })
                }
              </tr>
            ))
          }
        </tbody>
      </table>
      <Pagination
        totalPages={totalPages}
        currentPage={currentPage}
        setCurrentPage={setCurrentPage}
      />
    </div>
  )
};

const Pagination = (props) => {
  const {
    totalPages,
    currentPage,
    setCurrentPage,
  } = props;

  const [pageNumbers, setPageNumbers] = useState([...Array(totalPages + 1).keys()].slice(1, 11));

  const handlePageNumbers = () => {
    const start = totalPages - currentPage < 9 ? totalPages - 9 : currentPage > 8 ? currentPage - 5 : 1;
    const end = start + 10;
    setPageNumbers([...Array(totalPages + 1).keys()].slice(start, end))
  }

  useEffect(() => {
    return handlePageNumbers();
  }, [currentPage]);

  const setPage = {
    prevPage: function () {
      const newCurrentPage = currentPage > 1 ? currentPage - 1 : 1;
      setCurrentPage(newCurrentPage);
    },
    nextPage: () => {
      const newCurrentPage = currentPage !== totalPages ? currentPage + 1 : totalPages;
      setCurrentPage(newCurrentPage);
    },
    pageNumber: (num) => {
      setCurrentPage(num);
    }
  }

  return (
    <div className='Pagination__wrapper'>
      <ul className="Pagination__list">
        <li
          className={`Pagination__item-nav ${currentPage === 1 ? 'disabled' : ''}`}
          onClick={() => setPage.prevPage()}
        >
          <CaretIcon disabled={currentPage === 1} rotate="left" />
        </li>
        {pageNumbers.map((pageNum) => (
          <li
            className={`Pagination__item-number ${pageNum === currentPage ? 'Pagination__selected' : ''}`}
            onClick={() => setPage.pageNumber(pageNum)}
            key={pageNum}
          >
            { pageNum === currentPage && <span className="Pagination__circle" /> }
            <span>{pageNum}</span>
          </li>
        ))}
        <li
          className={`Pagination__item-nav ${currentPage === totalPages ? 'disabled' : ''}`}
          onClick={() => setPage.nextPage(currentPage)}
        >
          <CaretIcon disabled={currentPage === totalPages} rotate="right" />
        </li>
      </ul>
    </div>
  )
};

const CaretIcon = (props) => {
  const {
    disabled,
    rotate
  } = props;
  const fill = disabled ? 'gray' : '#fff';
  const transform = rotate === 'left' ? 'rotate(90 12 12)' : 'rotate(-90 12 12)';
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="36"
      height="36"
      viewBox="0 0 24 24"
    >
      <g fill="none" fillRule="evenodd" stroke="none" strokeWidth="1">
        <path d="M0 0L24 0 24 24 0 24z"></path>
        <g transform={transform}>
          <path d="M0 0L24 0 24 24 0 24z"></path>
          <path
            fill={fill}
            d="M7.146 10.146a.5.5 0 01.638-.057l.07.057 3.792 3.793a.5.5 0 00.638.058l.07-.058 3.792-3.793a.5.5 0 01.765.638l-.057.07-3.793 3.793a1.5 1.5 0 01-2.008.102l-.114-.102-3.793-3.793a.5.5 0 010-.708z"
          ></path>
        </g>
      </g>
    </svg>
  );
}

export default Users;
