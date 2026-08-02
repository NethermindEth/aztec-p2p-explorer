import Scene from '../../components/organisms/ThreeDWorld/Scene';
import ExploreMap from '../../components/atoms/Buttons/ExploreMap/ExploreMap';
import NetworkStats from '../../components/organisms/Statistics/NetworkStats/NetworkStats';
import PanelHeader from '../../components/molecules/PanelHeader/PanelHeader';
import PanelListBody from '../../components/molecules/PanelListBody/PanelListBody';
import PanelNodeDetails from '../../components/organisms/PanelNodeDetails/PanelNodeDetails';
import { useAtom } from 'jotai';
import { isMobileAtom, selectedNodeAtom, clusterPeersAtom, isTabletAtom } from '../../hooks/atoms';
import { API } from '../../api';
import { useEffect } from 'react';
import './ListView.css';

const ListView = () => {
  const [selectedNode] = useAtom(selectedNodeAtom);
  const [_isMobile] = useAtom(isMobileAtom);
  const [isTablet] = useAtom(isTabletAtom);

  const isMobile = _isMobile || isTablet;

  // Load peers map data for the 3D globe
  const { data: apiData, isLoading } = API.usePeersMap({});
  const [, setClusterPeers] = useAtom(clusterPeersAtom);

  useEffect(() => {
    if (!apiData) return;
    setClusterPeers(apiData.clusters);
  }, [apiData, isLoading, setClusterPeers]);

  return (
    <>
      {!isMobile && (
        <div className="list-wrapper fade-in">
          <div className="list-top-container">
            <div className="list-globe-container">
              {/* <div className="list-globe-background" /> */}
              <div className="list-canvas-container">
                <Scene />
              </div>
              <div className="list-map-button">
                <ExploreMap />
              </div>
            </div>

            <div className="list-panel-container">
              <PanelHeader />
              <div className="list-panel-content">
                <PanelListBody view="ListView" />
              </div>
            </div>
          </div>

          <div className="list-network-stats-container">
            <NetworkStats />
          </div>
        </div>
      )}

      {isMobile && (
        <div className="listview-mobile-container">
          {selectedNode ? (
            <div className="listview-selected-node">
              <PanelNodeDetails node={selectedNode} />
            </div>
          ) : (
            <div className="list-network-stats-container">
              <NetworkStats />
            </div>
          )}
        </div>
      )}
    </>
  );
};

export default ListView;
