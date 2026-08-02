import React, { useEffect, useState } from 'react';
import { Canvas } from '@react-three/fiber';
import { Suspense } from 'react';
import Earth from './Earth';
import InfoPanel from './InfoPanel';
import './Scene.css';
import { useAtom } from 'jotai';
import { clusterPeersAtom, globeRotationAtom, isMobileAtom } from '../../../hooks/atoms';
import { useHoverPanel } from '../../../hooks/useHoverPanel';
import { useLocation } from 'react-router-dom';
import { RawMarker } from '../../../utils/clusterMarkers';
import ReactDOM from 'react-dom';
import LoadingSpinner from '../../molecules/LoadingSpinner/LoadingSpinner';

type SceneProps = {
  isHoverable?: boolean;
};

type HoveredMarker = {
  data: RawMarker[];
  x: number;
  y: number;
};

const Scene = React.memo(function Scene({ isHoverable = false }: SceneProps) {
  const location = useLocation();
  const showListView = location.pathname === '/explore';
  const [mapPeers] = useAtom(clusterPeersAtom);
  const isMapView = location.pathname === '/';
  const portalTarget = document.getElementById('canvas-container');

  const [, setIsGlobeRotating] = useAtom(globeRotationAtom);
  const [isMobile] = useAtom(isMobileAtom);
  const [cameraSettings, setCameraSettings] = useState<{ position: [number, number, number]; fov?: number }>({
    position: [3.55, 3.55, 3.55],
  });
  const [hoveredMarker, setHoveredMarker] = useState<HoveredMarker | null>(null);

  const { isInfoPanelHovered, handlePointerOut, handlePointerOver, handleInfoPanelHover, forceCloseInfoPanel } =
    useHoverPanel(setHoveredMarker!, setIsGlobeRotating, { disableAutoClose: isMobile });

  //Not sure
  const sceneContainerClass = isHoverable || isInfoPanelHovered ? 'scene-container unset-z-index' : 'scene-container';

  useEffect(() => {
    if (isMobile) {
      setCameraSettings({ position: [3.55, 3.55, 3.55], fov: 60 });
    } else {
      setCameraSettings({ position: [3.55, 3.55, 3.55] });
    }
  }, [isMobile]);

  // Handle page visibility to pause/resume rendering when tab is hidden/shown
  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.hidden) {
        // Tab is hidden - Three.js will automatically pause with frameloop="demand"
        // No action needed as demand mode stops rendering when tab is hidden
      } else {
        // Tab is visible - scene will resume rendering on next interaction
        // No explicit action needed as demand mode will render on next invalidate()
      }
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    };
  }, []);

  const canvasProps = isMobile
    ? {
        camera: { position: cameraSettings.position, fov: cameraSettings.fov },
        dpr: 1,
        frameloop: 'demand' as const, // Render only when needed, not continuously
      }
    : {
        camera: { position: cameraSettings.position, fov: cameraSettings.fov },
        resize: { scroll: true, debounce: { scroll: 50, resize: 50 } },
        frameloop: 'demand' as const, // Render only when needed, not continuously
      };

  return (
    <>
      <div className={sceneContainerClass}>
        <Canvas {...canvasProps}>
          <Suspense fallback={null}>
            <Earth setHoveredMarker={handlePointerOver} onPointerOut={handlePointerOut} mapPeers={mapPeers || []} />
            <ambientLight intensity={1} />
          </Suspense>
        </Canvas>
        {isHoverable && hoveredMarker && !showListView && (
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
});

export default Scene;
