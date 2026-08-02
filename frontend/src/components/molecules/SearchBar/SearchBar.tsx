import './SearchBar.css';
import { Kbd } from '@radix-ui/themes';

interface ISearchBarProps {
  onChange: (event: React.ChangeEvent<HTMLInputElement>) => void;
}

const searchIconSvg = `
<svg xmlns="http://www.w3.org/2000/svg" width="17" height="17" viewBox="0 0 17 17" fill="none">
  <path fill-rule="evenodd" clip-rule="evenodd" d="M11.8998 7.77456C11.8998 8.3655 11.7834 8.95067 11.5572 9.49663C11.3311 10.0426 10.9996 10.5387 10.5818 10.9566C10.1639 11.3744 9.66784 11.7059 9.12187 11.9321C8.57591 12.1582 7.99075 12.2746 7.3998 12.2746C6.80885 12.2746 6.22369 12.1582 5.67772 11.9321C5.13176 11.7059 4.63568 11.3744 4.21782 10.9566C3.79995 10.5387 3.46849 10.0426 3.24234 9.49663C3.01619 8.95067 2.8998 8.3655 2.8998 7.77456C2.8998 6.58108 3.3739 5.43649 4.21782 4.59258C5.06173 3.74866 6.20632 3.27456 7.3998 3.27456C8.59327 3.27456 9.73786 3.74866 10.5818 4.59258C11.4257 5.43649 11.8998 6.58108 11.8998 7.77456ZM11.0798 12.5146C9.87418 13.4506 8.35717 13.8919 6.8376 13.7487C5.31803 13.6056 3.91014 12.8887 2.90055 11.744C1.89095 10.5993 1.35555 9.11285 1.40333 7.58731C1.45112 6.06176 2.0785 4.61176 3.15775 3.53251C4.23701 2.45326 5.687 1.82588 7.21255 1.77809C8.7381 1.73031 10.2245 2.26571 11.3692 3.27531C12.5139 4.2849 13.2308 5.69279 13.3739 7.21236C13.5171 8.73192 13.0758 10.2489 12.1398 11.4546L15.1798 14.4946C15.2535 14.5633 15.3126 14.6461 15.3536 14.7381C15.3946 14.8301 15.4166 14.9294 15.4184 15.0301C15.4202 15.1308 15.4016 15.2308 15.3639 15.3242C15.3262 15.4176 15.27 15.5024 15.1988 15.5736C15.1276 15.6449 15.0428 15.701 14.9494 15.7387C14.856 15.7764 14.756 15.795 14.6553 15.7932C14.5546 15.7914 14.4552 15.7694 14.3633 15.7284C14.2713 15.6874 14.1885 15.6283 14.1198 15.5546L11.0798 12.5146Z" fill="white"/>
</svg>
`;
const dataUri = `data:image/svg+xml;utf8,${encodeURIComponent(searchIconSvg)}`;

const SearchBar: React.FC<ISearchBarProps> = ({ onChange }) => {
  return (
    <div className="network-stats__search">
      <img src={dataUri} alt="Search Icon" className="network-stats__search-icon" />
      <input
        type="text"
        placeholder="Search by Peer ID"
        className="network-stats__search-input"
        style={{
          outline: 'none',
        }}
        onChange={onChange}
      />
      <Kbd size="1" className="search-kbd-icon">
        S
      </Kbd>
    </div>
  );
};

export default SearchBar;
