import Logo from '../../../assets/logos/aztecLight.svg?react';
import NethermindLogo from '../../../assets/logos/nethermindLogo.svg';
import { useAtom } from 'jotai';
import { showListViewAtom } from '../../../hooks/atoms';
import './MobileTopBar.css';
import RouterLink from '../../atoms/RouterLink/RouterLink';

const MobileTopBar = () => {
  const [, setShowListView] = useAtom(showListViewAtom);

  const handleClick = () => {
    setShowListView(false);
  };

  return (
    <RouterLink to={'/'} className="mobile-top-bar">
      <div className="mobile-logo-container" onClick={handleClick}>
        <Logo className="mobile-logo" />

        <div className="mobile-powered-by">
          <span>Powered by</span>
          <img src={NethermindLogo} alt="Nethermind Logo" />
        </div>
      </div>
    </RouterLink>
  );
};

export default MobileTopBar;
