import { useState, useRef, useEffect } from 'react';
import PanelHeader from '../../components/molecules/PanelHeader/PanelHeader';
import PanelListBody from '../../components/molecules/PanelListBody/PanelListBody';
import FilterBar from '../../components/molecules/FilterBar/FilterBar';
import SwitchToggle from '../../components/molecules/Switch/Switch';
import { useAtom } from 'jotai';
import { clusterPeersAtom, isMobileAtom } from '../../hooks/atoms';
import './MapView.css';
import { API } from '../../api';
import Scene from '../../components/organisms/ThreeDWorld/Scene';
import NodeModal from '../../components/organisms/Statistics/NodeModal/NodeModal';
import Scene2D from '../../components/organisms/TwoDWorld/Scene2D';

export type IFilters = {
  region: string[];
  clients: string[];
};

const MapView = () => {
  const [is3D, setIs3D] = useState(false);

  const [isExpanded, setIsExpanded] = useState(false);
  const [activeTab, setActiveTab] = useState('topNetworks');
  const [isMobile] = useAtom(isMobileAtom);

  const startYRef = useRef(0);
  const barRef = useRef<HTMLDivElement>(null);
  const tabsContainerRef = useRef<HTMLDivElement>(null);
  const [flValues, setFlValues] = useState<IFilters>({
    region: localStorage.getItem('region') ? JSON.parse(localStorage.getItem('region')!) : ['All'],
    clients: localStorage.getItem('clients') ? JSON.parse(localStorage.getItem('clients')!) : ['All'],
  });

  const { data: apiData, isLoading } = API.usePeersMap({
    continents: flValues.region.includes('All') ? [] : flValues.region,
    clients: flValues.clients.includes('All') ? [] : flValues.clients,
  });
  const [, setClusterPeers] = useAtom(clusterPeersAtom);

  useEffect(() => {
    if (!apiData) return;
    setClusterPeers(apiData.clusters);
  }, [apiData, isLoading, setClusterPeers]);

  useEffect(() => {
    localStorage.setItem('region', JSON.stringify(flValues.region));
    localStorage.setItem('clients', JSON.stringify(flValues.clients));
  }, [flValues.region, flValues.clients]);

  function toggleView() {
    setIs3D(!is3D);
  }

  const handleTouchStart = (e: TouchEvent) => {
    startYRef.current = e.touches[0].clientY;

    if (isExpanded) {
      setIsExpanded(false);
    }
  };

  const handleTouchMove = (e: TouchEvent) => {
    e.preventDefault();
    const deltaY = e.touches[0].clientY - startYRef.current;

    if (barRef.current) {
      const newTop = Math.max(0, Math.min(deltaY, window.innerHeight)) / 7;
      if (newTop <= 10) {
        barRef.current.style.top = `${newTop}rem`;
      }
    }
  };

  const handleTouchEnd = (e: TouchEvent) => {
    const deltaY = e.changedTouches[0].clientY - startYRef.current;

    if (barRef.current) {
      if (!isExpanded && deltaY > 50) {
        setIsExpanded(true);
      } else if (isExpanded && deltaY < -50) {
        setIsExpanded(false);
      } else if (!isExpanded && deltaY < 50) {
        barRef.current.style.top = '0';
      }
    }
  };

  useEffect(() => {
    if (barRef.current) {
      const barElement = barRef.current;

      barElement.addEventListener('touchstart', handleTouchStart);
      barElement.addEventListener('touchmove', handleTouchMove);
      barElement.addEventListener('touchend', handleTouchEnd);

      return () => {
        barElement.removeEventListener('touchstart', handleTouchStart);
        barElement.removeEventListener('touchmove', handleTouchMove);
        barElement.addEventListener('touchend', handleTouchEnd);
      };
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isExpanded]);

  const updateBarPosition = () => {
    if (tabsContainerRef.current && barRef.current) {
      const tabsContainerHeight = tabsContainerRef.current.clientHeight;
      if (tabsContainerHeight != 0) {
        barRef.current.style.top = `${tabsContainerHeight / 16}rem`;
      }
    }
  };

  useEffect(() => {
    if (isExpanded) {
      updateBarPosition();
    } else {
      if (barRef.current) {
        barRef.current.style.top = '0';
      }
    }

    window.addEventListener('resize', updateBarPosition);
    return () => {
      window.removeEventListener('resize', updateBarPosition);
    };
  }, [isExpanded]);

  useEffect(() => {
    updateBarPosition();
  }, [activeTab]);

  return (
    <div className="mapview-container fade-in">
      <div className="mapview-wrapper">
        <div id="canvas-container" className={'canvas-container'}>
          <div className="scene-filter-wrapper">
            {is3D ? <Scene isHoverable /> : <Scene2D isHoverable />}

            <div className={'desktop-only filter-bar-container'}>
              <FilterBar onFilterChange={setFlValues} />
            </div>
          </div>
        </div>

        <div>
          <div className="sidebar">
            <div className={'panel-header-wrapper'}>
              <PanelHeader />
            </div>

            {!isMobile && (
              <div className="panel-body-wrapper">
                <PanelListBody view="MapView" />
              </div>
            )}

            {!isMobile && (
              <div className={'three-d-toggle'}>
                <SwitchToggle is3D={is3D!} toggleView={toggleView!} />
              </div>
            )}
          </div>

          {/* bar and tabs for mobile view */}
          <div className={`mobile-only ${isExpanded ? 'expanded' : ''}`}>
            <div className="bar" ref={barRef} onClick={() => setIsExpanded(!isExpanded)}>
              <div className="bar-icon">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="25"
                  height="24"
                  viewBox="0 0 25 24"
                  fill="none"
                  style={{
                    transform: isExpanded ? 'rotate(180deg)' : 'rotate(0deg)',
                    transition: 'transform 0.2s ease',
                  }}
                >
                  <mask
                    id="mask0_3857_19453"
                    style={{ maskType: 'alpha' }}
                    maskUnits="userSpaceOnUse"
                    x="0"
                    y="0"
                    width="25"
                    height="24"
                  >
                    <rect x="0.871094" width="24" height="24" fill="#D9D9D9" />
                  </mask>
                  <g mask="url(#mask0_3857_19453)">
                    <path
                      d="M12.8715 17.6538L7.21777 12L8.27152 10.9462L12.1215 14.7963V5.25H13.6215V14.7963L17.4715 10.9462L18.5253 12L12.8715 17.6538Z"
                      fill="#a9cc1f"
                    />
                  </g>
                </svg>
              </div>
            </div>
            <div className="tabs-container" ref={tabsContainerRef}>
              {isExpanded && (
                <>
                  <div className="tabs-header">
                    <div
                      className={`tab ${activeTab === 'topNetworks' ? 'active' : ''}`}
                      onClick={() => setActiveTab('topNetworks')}
                    >
                      Top Countries
                    </div>
                    <div
                      className={`tab ${activeTab === 'topClients' ? 'active' : ''}`}
                      onClick={() => setActiveTab('topClients')}
                    >
                      Top Versions
                    </div>
                  </div>
                  <div className={`tab-content ${activeTab === 'topNetworks' ? 'active' : ''}`}>
                    {activeTab === 'topNetworks' && <PanelListBody view="TopNetworks" />}
                  </div>
                  <div className={`tab-content ${activeTab === 'topClients' ? 'active' : ''}`}>
                    {activeTab === 'topClients' && <PanelListBody view="TopClients" />}
                  </div>
                </>
              )}
            </div>
          </div>
        </div>
        {isMobile && (
          <div className="mobile-filter-bar-container">
            <div className="mobile-switch-container">
              <span>Map view</span>
              <SwitchToggle is3D={is3D!} toggleView={toggleView!} />
            </div>
            <FilterBar onFilterChange={setFlValues} />
          </div>
        )}

        <NodeModal />
      </div>
    </div>
  );
};

export default MapView;
