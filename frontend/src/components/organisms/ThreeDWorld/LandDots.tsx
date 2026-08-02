import React, { useRef, useEffect } from 'react';
import { useAtom } from 'jotai';
import { positionsAtom } from '../../../hooks/atoms';
import { useLoadImageData } from '../../../hooks/useLoadImageData';
import * as THREE from 'three';

const LandDots: React.FC = React.memo(() => {
  const [positions] = useAtom(positionsAtom);
  const meshRef = useRef<THREE.InstancedMesh>(null);
  const geometryRef = useRef(new THREE.SphereGeometry(0.01, 0, 0));
  const tempObject = new THREE.Object3D();

  useLoadImageData();

  useEffect(() => {
    if (!meshRef.current || !positions) return;

    const numPoints = positions.length / 3;

    for (let i = 0; i < numPoints; i++) {
      tempObject.position.set(positions[i * 3], positions[i * 3 + 1], positions[i * 3 + 2]);
      tempObject.updateMatrix();
      meshRef.current.setMatrixAt(i, tempObject.matrix);
    }

    meshRef.current.instanceMatrix.needsUpdate = true;
  });

  if (!positions) return null;

  return (
    <>
      <instancedMesh ref={meshRef} args={[geometryRef.current, undefined, positions.length / 3]}>
        <meshStandardMaterial color="#c5c4ca" />
      </instancedMesh>
    </>
  );
});

export default LandDots;
