import logoImg from '../../../assets/logos/newlogo.svg';
import { useAtom } from 'jotai';
import { showListViewAtom } from '../../../hooks/atoms';
import './TopMenu.css';
import RouterLink from '../../atoms/RouterLink/RouterLink';
import NetworkIndicator from '../NetworkIndicator/NetworkIndicator';

const TopMenu = () => {
  const [showListView, setShowListView] = useAtom(showListViewAtom);

  const handleOnClick = () => {
    if (!showListView) return;
    setShowListView(false);
  };

  return (
    <div className="top-menu-container">
      <div className="top-menu-wrapper">
        <RouterLink to={'/'} className="top-menu-logo-container" onClick={handleOnClick}>
          <img src={logoImg} className="logo" alt="logo" />
        </RouterLink>
        <div className="top-menu-network-container">
          <a
            href="https://docs.aztec.network/operate/operators"
            target="_blank"
            rel="noopener noreferrer"
            className="banner-link"
          >
            <span className="banner-link-text">Run an Aztec Node</span>
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 16 16" fill="none">
              <mask
                id="mask0_3098_704"
                style={{ maskType: 'alpha' }}
                maskUnits="userSpaceOnUse"
                x="0"
                y="0"
                width="16"
                height="16"
              >
                <rect width="16" height="16" fill="currentColor" />
              </mask>
              <g mask="url(#mask0_3098_704)">
                <path
                  d="M10.748 5.47317L4.57021 11.6512C4.42665 11.7946 4.26649 11.8629 4.08971 11.856C3.91293 11.8491 3.75282 11.7739 3.60938 11.6305C3.46582 11.4869 3.39404 11.3233 3.39404 11.1397C3.39404 10.956 3.46582 10.7924 3.60938 10.6488L9.76654 4.49167H4.63971C4.44582 4.49167 4.2821 4.42523 4.14854 4.29234C4.01488 4.15934 3.94804 3.99639 3.94804 3.8035C3.94804 3.61062 4.01488 3.4465 4.14854 3.31117C4.2821 3.17595 4.44582 3.10834 4.63971 3.10834H11.4397C11.6336 3.10834 11.7974 3.17512 11.931 3.30867C12.0646 3.44234 12.1314 3.60612 12.1314 3.8V10.6C12.1314 10.7939 12.0649 10.9577 11.932 11.0913C11.799 11.2249 11.6361 11.2917 11.4432 11.2917C11.2503 11.2917 11.0863 11.2249 10.951 11.0913C10.8157 10.9577 10.748 10.7939 10.748 10.6V5.47317Z"
                  fill="currentColor"
                />
              </g>
            </svg>
          </a>
          <NetworkIndicator />
        </div>
      </div>
    </div>
  );
};

export default TopMenu;
