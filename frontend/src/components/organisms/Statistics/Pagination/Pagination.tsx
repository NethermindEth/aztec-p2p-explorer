import './Pagination.css';
import PaginationButton from './PaginationButton';

interface PaginationProps {
  currentPage: number;
  totalPages: number;
  onFirst: () => void;
  onPrevious: () => void;
  onNext: () => void;
  onLast: () => void;
}

const Pagination: React.FC<PaginationProps> = ({ currentPage, totalPages, onFirst, onPrevious, onNext, onLast }) => {
  return (
    <div className="network-stats__pagination">
      <button
        onClick={onFirst}
        className={`network-stats__pagination-button ${currentPage === 1 ? 'disabled' : ''}`}
        disabled={currentPage === 1}
      >
        First
      </button>

      <div className="network-stats__pagination-center">
        <PaginationButton
          onClick={onPrevious}
          disabled={currentPage === 1}
          direction="previous"
          arrowIconDirectionPath="M15 19.7747L8 12.7747L15 5.77469"
        />
        <span className="network-stats__pagination-info">
          Page {currentPage} of {totalPages}
        </span>
        <PaginationButton
          onClick={onNext}
          disabled={currentPage === totalPages}
          direction="next"
          arrowIconDirectionPath="M9 5.77469L16 12.7747L9 19.7747"
        />
      </div>

      <button
        onClick={onLast}
        className={`network-stats__pagination-button last ${currentPage === totalPages ? 'disabled' : ''}`}
        disabled={currentPage === totalPages}
      >
        Last
      </button>
    </div>
  );
};

export default Pagination;
