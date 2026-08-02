import React from 'react';

interface PaginationButtonProps {
  onClick: () => void;
  disabled: boolean;
  direction: 'previous' | 'next';
  arrowIconDirectionPath: string;
}

const arrowIcon = (drawingPath: string) => {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="25" viewBox="0 0 24 25" fill="none">
      <path d={drawingPath} stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
};
const PaginationButton: React.FC<PaginationButtonProps> = ({ onClick, disabled, arrowIconDirectionPath }) => {
  return (
    <button
      onClick={onClick}
      className={`network-stats__pagination-button arrow ${disabled ? 'disabled' : ''}`}
      disabled={disabled}
    >
      {arrowIcon(arrowIconDirectionPath)}
    </button>
  );
};

export default PaginationButton;
