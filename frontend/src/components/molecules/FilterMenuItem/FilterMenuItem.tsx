import React, { useCallback } from 'react';
import { Checkbox } from '@radix-ui/themes';
import './FilterMenuItem.css';

interface FilterMenuItemProps {
  option: string;
  selectedValues: string[];
  counts?: { [key: string]: number };
  handleValueChange: (option: string) => void;
}

const FilterMenuItem: React.FC<FilterMenuItemProps> = ({ option, selectedValues, counts = {}, handleValueChange }) => {
  const displayOption = option === '' ? 'Unknown' : option;
  const isSelected = selectedValues.includes(option) || (option === '' && selectedValues.includes('Unknown'));
  const count = counts[option.toLowerCase()] || 0;

  const handleClick = useCallback(
    (e: React.MouseEvent<HTMLDivElement>) => {
      e.stopPropagation();
      handleValueChange(option);
    },
    [handleValueChange, option]
  );

  return (
    <div
      key={option}
      className={`dropdown-menu-item ${isSelected ? 'dropdown-menu-item-selected' : ''}`}
      onClick={handleClick}
    >
      <div className="dropdown-menu-item-label">
        <div className="dropdown-menu-item-container">
          <span className="dropdown-menu-item-option">{displayOption}</span>
          {count > 0 && <span className="dropdown-menu-item-count">{count}</span>}
        </div>
        <Checkbox
          checked={isSelected}
          onCheckedChange={() => handleValueChange(option)}
          size="1"
          variant="surface"
          highContrast={false}
          color="sky"
          style={{ cursor: 'pointer' }}
        />
      </div>
    </div>
  );
};

export default FilterMenuItem;
