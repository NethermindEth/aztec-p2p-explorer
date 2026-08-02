import React from 'react';
import Icon from '../../atoms/Icons/StatisticsFilterIcon/StatisticsFilterIcon';
import { FilterValues } from '../../../types';
import './SelectedFilters.css';

interface SelectedFiltersProps {
  filterValues: FilterValues;
  handleFilterChange: (newClient: string[], newSync: string[]) => void;
}

const SelectedFilters: React.FC<SelectedFiltersProps> = ({ filterValues, handleFilterChange }) => {
  const renderFilterTag = (filterValue: string, filterType: keyof FilterValues) => {
    if (filterValue === 'All') return null;

    const handleRemove = () => {
      const newFilterValues: FilterValues = {
        client: [...filterValues.client],
        sync: [...filterValues.sync],
      };

      newFilterValues[filterType] = newFilterValues[filterType].filter((value) => value !== filterValue);
      handleFilterChange(newFilterValues.client, newFilterValues.sync);
    };

    const displayValue = filterValue === '' ? 'Unknown' : filterValue;

    return (
      <div key={filterValue} className="filter-tag">
        <span className="selected-filter-text">{displayValue}</span>

        <button onClick={handleRemove} className="selected-filter-button">
          <Icon label={'cross'} />
        </button>
      </div>
    );
  };

  const hasNonAllFilters =
    filterValues.client.some((client) => client !== 'All') || filterValues.sync.some((sync) => sync !== 'All');

  return (
    <div className={`selected-filters ${!hasNonAllFilters ? 'hidden' : ''}`}>
      {filterValues.client.map((client) => renderFilterTag(client, 'client'))}
      {filterValues.sync.map((sync) => renderFilterTag(sync, 'sync'))}
      <button className="selected-filters-clear-button" onClick={() => handleFilterChange(['All'], ['All'])}>
        Clear all filters
      </button>
    </div>
  );
};

export default SelectedFilters;
