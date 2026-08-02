import { Button, Select } from '@radix-ui/themes';
import { ChevronRightIcon, ChevronLeftIcon, DoubleArrowRightIcon, DoubleArrowLeftIcon } from '@radix-ui/react-icons';
import './Pager.css';

export type PagerProps = {
  currentPage: number;
  totalPages: number;
  rowsPerPage: number;
  rowsPerPageOptions?: number[];
  isLoading?: boolean;
  onRowsPerPageChange: (value: number) => void;
  onPageChange: (newPage: number) => void;
};

type PagerBtnProps = {
  type: 'prev' | 'next' | 'first' | 'last';
  onClick?: (event: React.MouseEvent<HTMLButtonElement, MouseEvent>) => void;
  disabled?: boolean;
  className?: string;
};
const PagerBtn = ({ type, onClick, disabled, className }: PagerBtnProps) => {
  const cssClass = className || '';
  const getIcon = () => {
    switch (type) {
      case 'prev':
        return <ChevronLeftIcon />;
      case 'next':
        return <ChevronRightIcon />;
      case 'first':
        return <DoubleArrowLeftIcon />;
      case 'last':
        return <DoubleArrowRightIcon />;
      default:
        return <ChevronLeftIcon />;
    }
  };
  return (
    <Button radius="none" variant="soft" className={`pager-btn ${cssClass}`} onClick={onClick} disabled={disabled}>
      {getIcon()}
    </Button>
  );
};

const Pager: React.FC<PagerProps> = ({
  currentPage,
  totalPages,
  rowsPerPage,
  rowsPerPageOptions = [10, 25, 50, 100],
  isLoading,
  onRowsPerPageChange,
  onPageChange,
}) => {
  return (
    <div className="pager-container">
      <div className="pager-rows-container">
        <p className="pager-rows-container-text">SHOW</p>
        <Select.Root
          defaultValue={String(rowsPerPage)}
          onValueChange={(e) => onRowsPerPageChange(Number(e))}
          disabled={isLoading}
        >
          <Select.Trigger className="pager-select" radius="none" />
          <Select.Content>
            {rowsPerPageOptions.map((item, idx) => (
              <Select.Item key={idx} value={String(item)}>
                {item}
              </Select.Item>
            ))}
          </Select.Content>
        </Select.Root>
      </div>

      <div className="pager-navigation-container">
        <div className="pager-page-count">
          PAGE <b>{currentPage}</b> of <b>{totalPages}</b>
        </div>

        <div className="pager-btns-container">
          {/* <PagerBtn type="first" onClick={() => onPageChange(1)} disabled={currentPage <= 1 || isLoading} /> */}
          <PagerBtn
            type="prev"
            onClick={() => onPageChange(currentPage - 1)}
            disabled={currentPage <= 1 || isLoading}
          />
          <PagerBtn
            type="next"
            onClick={() => onPageChange(currentPage + 1)}
            disabled={currentPage >= totalPages || isLoading}
          />
          {/* <PagerBtn
            type="last"
            onClick={() => onPageChange(totalPages)}
            disabled={currentPage >= totalPages || isLoading}
          /> */}
        </div>
      </div>
    </div>
  );
};

export default Pager;
