import { useState, useRef, useEffect } from 'react';
import { getNetworkFromDomain, navigateToNetwork, getNetworkDisplayName, NetworkType } from '../../../utils/network';
import './NetworkIndicator.css';

const NetworkIndicator = () => {
  const [isOpen, setIsOpen] = useState(false);
  const wrapperRef = useRef<HTMLDivElement>(null);

  // Detect current network from domain
  const currentNetwork = getNetworkFromDomain();

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (wrapperRef.current && !wrapperRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  const toggleDropdown = () => {
    setIsOpen(!isOpen);
  };

  const handleNetworkChange = (network: NetworkType) => {
    if (network !== currentNetwork) {
      // Navigate to the other network's domain
      navigateToNetwork(network);
    }
    setIsOpen(false);
  };

  return (
    <div className="network-indicator-wrapper" ref={wrapperRef}>
      <div className="network-indicator-container" onClick={toggleDropdown}>
        <div className="network-status">
          <span className="green-dot"></span>
        </div>
        <span className="network-label">{getNetworkDisplayName(currentNetwork)}</span>
        <div className="network-indicator-dropdown-icon">
          <svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path
              d="M3.13523 6.15803C3.3241 5.95657 3.64052 5.94637 3.84197 6.13523L7.5 9.56464L11.158 6.13523C11.3595 5.94637 11.6759 5.95657 11.8648 6.15803C12.0536 6.35949 12.0434 6.67591 11.842 6.86477L7.84197 10.6148C7.64964 10.7951 7.35036 10.7951 7.15803 10.6148L3.15803 6.86477C2.95657 6.67591 2.94637 6.35949 3.13523 6.15803Z"
              fill="currentColor"
              fillRule="evenodd"
              clipRule="evenodd"
            ></path>
          </svg>
        </div>
      </div>
      {isOpen && (
        <div className="network-dropdown-menu">
          <div
            className={`network-dropdown-item ${currentNetwork === 'aztec-testnet' ? 'selected' : ''}`}
            onClick={() => handleNetworkChange('aztec-testnet')}
          >
            <span className="green-dot"></span>
            <span>Aztec Testnet</span>
            {currentNetwork === 'aztec-testnet' && (
              <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none">
                <mask
                  id="mask0_2925_5594"
                  style={{ maskType: 'alpha' }}
                  maskUnits="userSpaceOnUse"
                  x="0"
                  y="0"
                  width="24"
                  height="24"
                >
                  <rect width="24" height="24" fill="#D9D9D9" />
                </mask>
                <g mask="url(#mask0_2925_5594)">
                  <path
                    d="M9.72476 14.5729L17.0923 7.20542C17.3019 6.99425 17.5479 6.88867 17.8303 6.88867C18.1124 6.88867 18.3592 6.99425 18.5705 7.20542C18.7817 7.41675 18.8873 7.66284 18.8873 7.94367C18.8873 8.22467 18.7849 8.46751 18.5803 8.67217L10.4765 16.7909C10.2667 17.0021 10.0207 17.1077 9.73851 17.1077C9.45618 17.1077 9.20935 17.0021 8.99801 16.7909L5.43501 13.2279C5.22385 13.0166 5.11726 12.7712 5.11526 12.4917C5.11326 12.2122 5.21785 11.9668 5.42901 11.7554C5.64035 11.5443 5.88718 11.4387 6.16951 11.4387C6.45168 11.4387 6.69568 11.5443 6.90151 11.7554L9.72476 14.5729Z"
                    fill="white"
                  />
                </g>
              </svg>
            )}
          </div>
          <div
            className={`network-dropdown-item ${currentNetwork === 'aztec-mainnet' ? 'selected' : ''}`}
            onClick={() => handleNetworkChange('aztec-mainnet')}
          >
            <span className="green-dot"></span>
            <span>Aztec Mainnet</span>
            {currentNetwork === 'aztec-mainnet' && (
              <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none">
                <mask
                  id="mask0_2925_5594"
                  style={{ maskType: 'alpha' }}
                  maskUnits="userSpaceOnUse"
                  x="0"
                  y="0"
                  width="24"
                  height="24"
                >
                  <rect width="24" height="24" fill="#D9D9D9" />
                </mask>
                <g mask="url(#mask0_2925_5594)">
                  <path
                    d="M9.72476 14.5729L17.0923 7.20542C17.3019 6.99425 17.5479 6.88867 17.8303 6.88867C18.1124 6.88867 18.3592 6.99425 18.5705 7.20542C18.7817 7.41675 18.8873 7.66284 18.8873 7.94367C18.8873 8.22467 18.7849 8.46751 18.5803 8.67217L10.4765 16.7909C10.2667 17.0021 10.0207 17.1077 9.73851 17.1077C9.45618 17.1077 9.20935 17.0021 8.99801 16.7909L5.43501 13.2279C5.22385 13.0166 5.11726 12.7712 5.11526 12.4917C5.11326 12.2122 5.21785 11.9668 5.42901 11.7554C5.64035 11.5443 5.88718 11.4387 6.16951 11.4387C6.45168 11.4387 6.69568 11.5443 6.90151 11.7554L9.72476 14.5729Z"
                    fill="white"
                  />
                </g>
              </svg>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default NetworkIndicator;
