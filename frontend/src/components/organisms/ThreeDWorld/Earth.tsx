import React, { useRef } from 'react';
import { useLoader, useFrame, useThree } from '@react-three/fiber';
import { OrbitControls } from '@react-three/drei';

import { TextureLoader, Group } from 'three';
import worldMap from '../../../assets/textures/world.jpg';
import LandDots from './LandDots';
import Atmosphere from './AtmosphereShader';
import Markers from './Markers';

import { useAtom } from 'jotai';
import { globeRotationAtom } from '../../../hooks/atoms';
import { RawMarker } from '../../../utils/clusterMarkers';

type EarthProps = {
  setHoveredMarker: React.Dispatch<React.SetStateAction<{ data: RawMarker[] | null; x: number; y: number } | null>>;
  onPointerOut: () => void;
  mapPeers: RawMarker[];
};

const Earth: React.FC<EarthProps> = ({ setHoveredMarker, onPointerOut, mapPeers }) => {
  const texture = useLoader(TextureLoader, worldMap);
  const earthRef = useRef<Group>(null!);
  const [isGlobeRotating] = useAtom(globeRotationAtom);
  const { invalidate } = useThree();

  useFrame(() => {
    if (earthRef.current && isGlobeRotating) {
      earthRef.current.rotation.y += 0.0005;
      invalidate(); // Request next frame only when rotating
    }
  });

  return (
    <>
      {/* <Stars radius={300} depth={60} count={10000} factor={7} saturation={0} fade /> */}
      <group ref={earthRef}>
        <mesh>
          <ambientLight intensity={0.5} />
          <sphereGeometry args={[3, 32, 32]} />
          <meshStandardMaterial visible={true} map={texture} />
        </mesh>
        <LandDots />
        <Atmosphere />
        <Markers setHoveredMarker={setHoveredMarker} onPointerOut={onPointerOut} markersData={mapPeers} />
      </group>
      <OrbitControls
        enableZoom={false}
        zoomSpeed={0.1}
        enableRotate={true}
        rotateSpeed={0.6}
        minDistance={5.4}
        maxDistance={7.2}
        enablePan={false}
        enableDamping={true}
        dampingFactor={0.05}
        onChange={() => invalidate()} // Trigger render on user interaction
      />
    </>
  );
};

export default Earth;
