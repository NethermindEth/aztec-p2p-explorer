import React, { useMemo, useCallback, useState, useRef } from 'react';
import { ComposableMap, Geographies, Geography, Marker, ZoomableGroup } from 'react-simple-maps';
import features from './features.json';
import './Scene2D.css';
import InfoPanel from '../ThreeDWorld/InfoPanel';
import { useHoverPanel } from '../../../hooks/useHoverPanel';
import { useAtom } from 'jotai';
import { clusterPeersAtom, isMobileAtom } from '../../../hooks/atoms';
import clusterMarkers, { ClusteredMarker, RawMarker } from '../../../utils/clusterMarkers';
import LoadingSpinner from '../../molecules/LoadingSpinner/LoadingSpinner';
import ReactDOM from 'react-dom';
import { useLocation } from 'react-router-dom';

type SceneProps = {
  isHoverable?: boolean;
};

type HoveredMarker = {
  data: RawMarker[];
  x: number;
  y: number;
};

const getMarkerAttributes = (count: number, zoom: number) => {
  const baseSize = 12;
  const range = [9, 10, 20, 30, 40, 50, 60, 70, 80, 99, 199, 299, 399, 499, 599, 699, 799, 899, 999];

  let r = Math.min((baseSize + 7) / zoom, 14); // default max radius size
  let text = '999+'; // default max text value

  for (let i = 0; i < range.length; i++) {
    if (count <= range[i]) {
      r = Math.min((baseSize + Math.min(i, 10)) / zoom, 6 + Math.min(i, 10));
      text = String(count);
      break;
    }
  }

  const textSize = Math.min(12 / zoom, 10);
  return { r, text, textSize };
};

const Scene2D: React.FC<SceneProps> = ({ isHoverable }) => {
  const [mapPeers] = useAtom(clusterPeersAtom);
  const [isMobile] = useAtom(isMobileAtom);
  const location = useLocation();
  const isMapView = location.pathname === '/';
  const [zoomLevel, setZoomLevel] = useState(1);
  const [hoveredMarker, setHoveredMarker] = useState<HoveredMarker | null>(null);
  const debounceRef = useRef<ReturnType<typeof setTimeout>>();

  const portalTarget = document.getElementById('canvas-container');

  const { handlePointerOut, handlePointerOver, handleInfoPanelHover, forceCloseInfoPanel } = useHoverPanel(
    setHoveredMarker,
    undefined,
    { disableAutoClose: isMobile }
  );

  const calculateEpsilon = useCallback((zoom: number) => {
    return 7 / zoom;
  }, []);

  const clusteredMarkers = useMemo(() => {
    if (!mapPeers || !Array.isArray(mapPeers)) {
      return [];
    }

    const validPeers = mapPeers.filter(
      (peer) => peer.latitude != null && peer.longitude != null && !(peer.latitude === 0 && peer.longitude === 0)
    );

    const epsilon = calculateEpsilon(zoomLevel);
    const clusters = clusterMarkers(validPeers, epsilon, 1);

    return clusters;
  }, [mapPeers, zoomLevel, calculateEpsilon]);

  const handleMarkerHover = (event: React.MouseEvent, cluster: ClusteredMarker) => {
    const rect = event.currentTarget.getBoundingClientRect();
    handlePointerOver({
      data: cluster.data,
      x: rect.x + rect.width / 2,
      y: rect.y,
    });
  };

  const sceneContainerClass = hoveredMarker ? 'map-2d unset-z-index' : 'map-2d';

  return (
    <>
      <div className={sceneContainerClass}>
        {/* <div className="gradient-background gradient-bg-1"></div>
        <div className="gradient-background gradient-bg-2"></div> */}
        <ComposableMap className="composable-map">
          <ZoomableGroup
            center={[0, 0]}
            zoom={1.2}
            minZoom={1}
            maxZoom={16}
            translateExtent={[
              [-20, -20],
              [900, 700],
            ]}
            onMoveEnd={({ zoom }) => {
              // Debounce zoom level updates to improve performance
              if (debounceRef.current) {
                clearTimeout(debounceRef.current);
              }
              debounceRef.current = setTimeout(() => {
                setZoomLevel(zoom);
              }, 150);
            }}
          >
            <Geographies geography={features}>
              {({ geographies }) => geographies.map((geo) => <Geography key={geo.rsmKey} geography={geo} />)}
            </Geographies>
            {clusteredMarkers.map((cluster, idx) => {
              const coordinates: [number, number] = [cluster.longitude, cluster.latitude];
              const { r, text, textSize } = getMarkerAttributes(cluster.count, zoomLevel);

              return (
                <Marker
                  className="marker-class"
                  tabIndex={1000}
                  key={idx}
                  coordinates={coordinates}
                  onMouseEnter={(e) => handleMarkerHover(e, cluster)}
                  onMouseLeave={handlePointerOut}
                >
                  <circle className="circle-class" r={r} fill="#A9CC1F" />
                  <text
                    className="marker-text"
                    textAnchor="middle"
                    fill="#000"
                    dy="0.33em"
                    style={{ fontSize: textSize }}
                  >
                    {text}
                  </text>
                </Marker>
              );
            })}
          </ZoomableGroup>
        </ComposableMap>

        {hoveredMarker && isHoverable && (
          <InfoPanel
            data={hoveredMarker.data}
            x={hoveredMarker.x}
            y={hoveredMarker.y}
            onHover={handleInfoPanelHover}
            onClose={forceCloseInfoPanel}
          />
        )}
      </div>
      {portalTarget &&
        isMapView &&
        ReactDOM.createPortal(
          <div className="map-loader-backdrop" style={{ display: mapPeers?.length ? 'none' : 'block' }}>
            <LoadingSpinner />
          </div>,
          portalTarget
        )}
    </>
  );
};

export default Scene2D;
