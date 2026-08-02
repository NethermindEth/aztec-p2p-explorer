import React, { useState, useEffect, useRef, useCallback } from 'react';

import Icon from '../../atoms/Icons/StatisticsFilterIcon/StatisticsFilterIcon';
import './StatisticsFilter.css';
import '../../../globalStyles.css';
import upArrowIcon from '../../../assets/icons/upArrow.svg';
import downArrowIcon from '../../../assets/icons/downArrow.svg';
import FilterMenuItem from '../FilterMenuItem/FilterMenuItem';
import '../FilterMenuItem/FilterMenuItem.css';

interface StatisticsFilterProps {
  key: string;
  label: string;
  value: string[];
  options: string[];
  onFilterChange?: (value: string[]) => void;
  isOpen: boolean;
  onClose: () => void;
}

const StatisticsFilter: React.FC<StatisticsFilterProps> = ({
  label,
  value,
  options,
  onFilterChange,
  isOpen,
  onClose,
}) => {
  const [selectedValues, setSelectedValues] = useState<string[]>(value);
  const menuRef = useRef<HTMLDivElement>(null);

  const handleValueChange = useCallback(
    (newValue: string) => {
      let updatedValues: string[] = [];

      // If there are more than one selected valuesthe user selects 'All'
      // set the selected values to only 'All'.
      if (newValue === 'All' && selectedValues.length > 1) {
        setSelectedValues(['All']);

        // If the user tries to unselect 'All' when it is the only selected value,
        // do nothing to ensure at least one value remains selected.
      } else if (newValue === 'All' && selectedValues.length === 1 && selectedValues[0] === 'All') {
        return;
      } else {
        updatedValues = selectedValues.includes(newValue)
          ? selectedValues.filter((val) => val !== newValue)
          : [...selectedValues.filter((val) => val !== 'All'), newValue];
      }

      setSelectedValues(updatedValues);
      onFilterChange?.(updatedValues);
    },
    [selectedValues, onFilterChange]
  );

  const handleHoverAway = useCallback(
    (event: MouseEvent) => {
      if (isOpen && menuRef.current && !menuRef.current.contains(event.target as Node)) {
        onClose();
      }
    },
    [isOpen, onClose]
  );

  useEffect(() => {
    if (isOpen) {
      document.addEventListener('mousemove', handleHoverAway);
    } else {
      document.removeEventListener('mousemove', handleHoverAway);
    }
    return () => {
      document.removeEventListener('mousemove', handleHoverAway);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isOpen]);

  return (
    <div className={`statistics-filter ${isOpen ? 'statistics-filter-open' : ''}`} ref={menuRef}>
      <div className="statistics-filter-header" onClick={onClose}>
        <Icon label={label} />
        <span className="statistics-filter-label">{label}</span>
        {isOpen ? <img src={upArrowIcon} alt="up arrow icon" /> : <img src={downArrowIcon} alt="down arrow icon" />}
      </div>
      {isOpen && (
        <div className="dropdown-menu dropdown-menu-statistics custom-scrollbar">
          {options.map((option) => (
            <FilterMenuItem
              key={option}
              option={option}
              selectedValues={selectedValues}
              handleValueChange={handleValueChange}
            />
          ))}
        </div>
      )}
    </div>
  );
};

export default StatisticsFilter;
