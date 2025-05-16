import React, { useState, useEffect } from 'react';
import Loader from './Loader';
import './RentalsHistory.css';

const RentalsHistory = () => {
  const [rentalsHistory, setRentalsHistory] = useState([]);
  const [loading, setLoading] = useState(false);

  const fetchRentalsHistory = async () => {
    setLoading(true);
    const reqRentalsHistory = await fetch('/rentals-history');
    const rentalsHistoryResult = await reqRentalsHistory.json();
    setRentalsHistory(rentalsHistoryResult);
    return setLoading(false);
  };

  useEffect(() => {
    fetchRentalsHistory();
  }, []);

  return (
    <div className="RentalsHistory">
        <h1>Rentals History </h1>
      { loading && <Loader />}
      { !!rentalsHistory.length && <Table rentalsHistory={rentalsHistory} /> }
    </div>
  )
};

const Table = ({ rentalsHistory }) => {
  const allRentalsHistory = [...rentalsHistory];
  const [currentRentalHistory, setCurrentRentalsHistory] = useState(allRentalsHistory.slice(0, 100));
  const [currentPage, setCurrentPage] = useState(1);

  const pageSize = 10;
  const totalPages = Math.ceil(allRentalsHistory.length / pageSize);

  useEffect(() => {
    handleRentalsHistoryToDisplay();
  }, [currentPage]);

  const handleRentalsHistoryToDisplay = () => {
    const endIndex = currentPage * pageSize;
    const startIndex = endIndex - pageSize;
    setCurrentRentalsHistory(allRentalsHistory.slice(startIndex, endIndex));
  }

  return (
    <div className="Table">
      <h5>Total Users: {allRentalsHistory.length ? allRentalsHistory.length : '0'}</h5>
      <table className='Table__table'>
        <thead className="Table__head">
          <tr className="Table__row">
            {
              Object.keys(allRentalsHistory[0]).map((key) => {
                return <th className="Table__header" key={key}>{key}</th>
              })
            }
          </tr>
        </thead>
        <tbody className="Table__body">
          {
            currentRentalHistory.map((rentalHistory) => (
              <tr className="Table__row" key={rentalHistory.id}>
                {
                  Object.keys(rentalHistory).map((property) => {
                    return (
                    <td className="Table__data" key={`${rentalHistory.id}--${property}`}>
                      {property === 'Price' 
                        ? new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(rentalHistory[property])
                        : rentalHistory[property]
                      }
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

  const [pageNumbers, setPageNumbers] = useState([]);

  const handlePageNumbers = () => {
    const maxVisiblePages = 10;
    let start = 1;
    let end = Math.min(maxVisiblePages, totalPages);

    if (totalPages > maxVisiblePages) {
      if (currentPage > 5) {
        start = Math.min(currentPage - 4, totalPages - maxVisiblePages + 1);
        end = start + maxVisiblePages - 1;
      }
    }

    setPageNumbers(Array.from({ length: end - start + 1 }, (_, i) => start + i));
  }

  useEffect(() => {
    handlePageNumbers();
  }, [currentPage, totalPages]);

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

export default RentalsHistory;
