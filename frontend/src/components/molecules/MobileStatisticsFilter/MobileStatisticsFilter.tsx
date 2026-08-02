import React, { useState } from 'react';
import { Checkbox } from '@radix-ui/themes';
import './MobileStatisticsFilter.css';

interface MobileStatisticsFilterProps {
  filters: {
    key: string;
    label: string;
    value: string[];
    options: string[];
    onFilterChange?: (value: string[]) => void;
  }[];
  onClose: () => void;
  onApplyFilters: (selectedValues: { client: string[]; sync: string[] }) => void;
}

const MobileStatisticsFilter: React.FC<MobileStatisticsFilterProps> = ({ filters, onClose, onApplyFilters }) => {
  const [selectedValues, setSelectedValues] = useState<{ [key: string]: string[] }>(
    filters.reduce((acc, filter) => ({ ...acc, [filter.key]: filter.value }), {})
  );

  const handleValueChange = (filterKey: string, newValue: string) => {
    let updatedValues: string[] = [];

    if (newValue === 'All') {
      updatedValues = selectedValues[filterKey].includes('All') ? [] : ['All'];
    } else {
      if (selectedValues[filterKey].includes(newValue)) {
        updatedValues = selectedValues[filterKey].filter((val) => val !== newValue);
      } else {
        updatedValues = [...selectedValues[filterKey].filter((val) => val !== 'All'), newValue];
      }

      if (newValue === 'Synced' || newValue === 'Unsynced') {
        updatedValues = [newValue];
      }
    }

    if (newValue === 'All' && updatedValues.length === 0) {
      updatedValues = ['All'];
    }

    setSelectedValues({ ...selectedValues, [filterKey]: updatedValues });
  };

  const handleApply = () => {
    onApplyFilters({
      client: selectedValues['client'] || [],
      sync: selectedValues['sync'] || [],
    });
  };

  const handleClearFilters = () => {
    const clearedValues = filters.reduce((acc, filter) => ({ ...acc, [filter.key]: ['All'] }), {});
    setSelectedValues(clearedValues);
  };

  return (
    <div className="mobile-statistics-filter">
      <div className="mobile-statistics-filter-container">
        <div className="mobile-statistics-filter-header">
          <span>Filters</span>
          <div onClick={onClose}>
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none">
              <path
                d="M6.06204 14.7115L5.28809 13.9375L9.22559 10L5.28809 6.06253L6.06204 5.28857L9.99954 9.22607L13.937 5.28857L14.711 6.06253L10.7735 10L14.711 13.9375L13.937 14.7115L9.99954 10.774L6.06204 14.7115Z"
                fill="#8E8C99"
              />
            </svg>
          </div>
        </div>
        <div className="mobile-statistics-filter-content">
          {filters.map((filter) => (
            <div key={filter.key} className="mobile-statistics-filter-section">
              <div className="mobile-statistics-filter-label">{filter.label}</div>
              {filter.options
                .filter((option) => option !== 'All')
                .map((option) => (
                  <div
                    key={option}
                    className="mobile-statistics-filter-item"
                    onClick={() => handleValueChange(filter.key, option)}
                  >
                    <span>{option === '' ? 'Unknown' : option}</span>
                    <Checkbox
                      checked={selectedValues[filter.key].includes(option)}
                      color="yellow"
                      onCheckedChange={() => handleValueChange(filter.key, option)}
                    />
                  </div>
                ))}
            </div>
          ))}
        </div>
        <div className="mobile-statistics-filter-actions">
          <button onClick={handleClearFilters}>Clear filters</button>
          <button onClick={handleApply}>Apply filters</button>
        </div>
      </div>
    </div>
  );
};

export default MobileStatisticsFilter;
