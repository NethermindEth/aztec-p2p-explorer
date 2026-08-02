import React, { useState, useEffect, useRef } from 'react';
import './Dropdown.css';
import FilterMenuItem from '../FilterMenuItem/FilterMenuItem';
import { isMobileAtom } from '../../../hooks/atoms';
import { useAtom } from 'jotai';
import '../FilterMenuItem/FilterMenuItem.css';

interface DropdownProps {
  label: string;
  value: string[];
  options: string[];
  counts: { [key: string]: number };
  onFilterChange?: (value: string[]) => void;
  isOpen: boolean;
  onClose: () => void;
  filterBarRef: React.RefObject<HTMLDivElement>;
}

const dropdownArrowIcon = (
  <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none">
    <path d="M12 13L7 8H17L12 13Z" fill="#959595" />
  </svg>
);

const Dropdown: React.FC<DropdownProps> = ({
  label,
  value,
  options,
  counts,
  onFilterChange,
  isOpen,
  onClose,
  filterBarRef,
}) => {
  const [selectedValues, setSelectedValues] = useState<string[]>(value);
  const [isMobile] = useAtom(isMobileAtom);
  const menuRef = useRef<HTMLDivElement>(null);
  const hoverTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const handleValueChange = (newValue: string) => {
    newValue = newValue === '' ? 'Unknown' : newValue;

    let updatedValues: string[] = [];
    if (newValue === 'All') {
      if (selectedValues.includes('All')) {
        updatedValues = [];
      } else {
        updatedValues = ['All'];
      }
    } else {
      if (selectedValues.includes(newValue)) {
        updatedValues = selectedValues.filter((val) => val !== newValue);
      } else {
        updatedValues = [...selectedValues.filter((val) => val !== 'All'), newValue];
      }
    }

    // @note ensures at least one value is selected
    if (updatedValues.length === 0) {
      updatedValues = ['All'];
    }

    setSelectedValues(updatedValues);
    if (onFilterChange) {
      onFilterChange(updatedValues);
    }
  };

  const handleMouseEnter = () => {
    if (hoverTimeoutRef.current) {
      clearTimeout(hoverTimeoutRef.current);
      hoverTimeoutRef.current = null;
    }
  };

  const handleMouseLeave = () => {
    hoverTimeoutRef.current = setTimeout(() => {
      onClose();
    }, 250);
  };

  useEffect(() => {
    const menuElement = menuRef.current;
    const filterBarElement = filterBarRef.current;

    if (menuElement && filterBarElement) {
      // Add event listeners for both the menu and filter bar
      menuElement.addEventListener('mouseenter', handleMouseEnter);
      menuElement.addEventListener('mouseleave', handleMouseLeave);
      filterBarElement.addEventListener('mouseenter', handleMouseEnter);
      filterBarElement.addEventListener('mouseleave', handleMouseLeave);
    }

    return () => {
      // Clean up event listeners and timeouts
      if (menuElement && filterBarElement) {
        menuElement.removeEventListener('mouseenter', handleMouseEnter);
        menuElement.removeEventListener('mouseleave', handleMouseLeave);
        filterBarElement.removeEventListener('mouseenter', handleMouseEnter);
        filterBarElement.removeEventListener('mouseleave', handleMouseLeave);
      }
      if (hoverTimeoutRef.current) {
        clearTimeout(hoverTimeoutRef.current);
      }
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isOpen]);
  const handleDropdownWheel = (e: React.WheelEvent<HTMLDivElement>) => {
    e.stopPropagation();
  };

  const renderDropdownMenu = () => (
    <div className="dropdown-menu" ref={menuRef}>
      <div className="dropdown-menu-inner custom-scrollbar" onWheel={handleDropdownWheel}>
        {options.map((option) => (
          <FilterMenuItem
            key={option}
            option={option}
            selectedValues={selectedValues}
            counts={counts}
            handleValueChange={handleValueChange}
          />
        ))}
      </div>
    </div>
  );

  return (
    <>
      <div className="dropdown" onClick={onClose}>
        <div className="dropdown-header">
          <span className="dropdown-label">{label}</span>
          <span className="dropdown-arrow">{dropdownArrowIcon}</span>
        </div>
        <div className="dropdown-selected-text">
          {selectedValues.length > 0 ? `${selectedValues[0]}${selectedValues.length > 1 ? ' +' : ''}` : 'Select...'}
        </div>

        {!isMobile && isOpen && renderDropdownMenu()}
      </div>
      <svg xmlns="http://www.w3.org/2000/svg" width="2" height="72" viewBox="0 0 2 72" fill="none">
        <path opacity="0.07" d="M1.3335 72L1.3335 -1.78814e-07" stroke="#EBEBEB" />
      </svg>
      {isMobile && isOpen && renderDropdownMenu()}
    </>
  );
};

export default Dropdown;
