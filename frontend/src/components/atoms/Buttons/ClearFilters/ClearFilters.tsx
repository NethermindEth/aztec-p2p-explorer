import './ClearFilters.css';

interface ClearFiltersProps {
  clearFilters: () => void;
}
const ClearFilters = ({ clearFilters }: ClearFiltersProps) => {
  return (
    <button className="clear-filters-button" onClick={clearFilters}>
      Clear Filters
    </button>
  );
};

export default ClearFilters;
